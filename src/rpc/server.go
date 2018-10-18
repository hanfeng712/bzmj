package rpc

import (
	"errors"
	"logger"
	//	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	sysdebug "runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	ArgType    reflect.Type
	numCalls   uint
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

// Server represents an RPC Server.
type Server struct {
	mu           sync.RWMutex // protects the serviceMap
	serviceMap   map[string]*service
	reqLock      sync.Mutex // protects freeReq
	freeReq      *RequestWrap
	id           uint64
	connMap      map[uint64]RpcConn
	connLock     sync.RWMutex
	onConn       []func(conn RpcConn)
	onDisConn    []func(conn RpcConn)
	onCallBefore []func(conn RpcConn)
	onCallAfter  []func(conn RpcConn)
	quitSync     sync.RWMutex
	quit         bool
}

type RequestWrap struct {
	Request
	next *RequestWrap // for free list in Server
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{
		quit:         false,
		serviceMap:   make(map[string]*service),
		connMap:      make(map[uint64]RpcConn),
		onConn:       make([]func(conn RpcConn), 0),
		onDisConn:    make([]func(conn RpcConn), 0),
		onCallBefore: make([]func(conn RpcConn), 0),
		onCallAfter:  make([]func(conn RpcConn), 0),
	}
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

func (server *Server) RegCallBackOnConn(cb func(conn RpcConn)) {
	server.mu.Lock()
	server.onConn = append(server.onConn, cb)
	server.mu.Unlock()
}

func (server *Server) RegCallBackOnDisConn(cb func(conn RpcConn)) {
	server.mu.Lock()
	server.onDisConn = append(server.onDisConn, cb)
	server.mu.Unlock()
}

func (server *Server) RegCallBackOnCallBefore(cb func(conn RpcConn)) {
	server.mu.Lock()
	server.onCallBefore = append(server.onCallBefore, cb)
	server.mu.Unlock()
}

func (server *Server) RegCallBackOnCallAfter(cb func(conn RpcConn)) {
	server.mu.Lock()
	server.onCallAfter = append(server.onCallAfter, cb)
	server.mu.Unlock()
}

func (server *Server) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	server.mu.Lock()
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		logger.Fatal("rpc: no service name for type", s.typ.String())
	}
	if !isExported(sname) && !useName {
		s := "rpc Register: type " + sname + " is not exported"
		logger.Info(s)
		server.mu.Lock()
		return errors.New(s)
	}
	if _, present := server.serviceMap[sname]; present {
		server.mu.Lock()
		return errors.New("rpc: service already defined: " + sname)
	}
	s.name = sname
	s.method = make(map[string]*methodType)

	// Install the methods
	//fmt.Printf("Install the methods begine!")
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""
		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		logger.Info(str)
		server.mu.Unlock()
		return errors.New(str)
	}
	server.serviceMap[s.name] = s
	server.mu.Unlock()
	return nil
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using logger if reportErr is true.
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		//fmt.Printf("suitableMethods %s, %s, %s, %d \n", mtype, mname, method.PkgPath, mtype.NumIn())
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		// Method needs three ins: receiver, connid, *args.
		if mtype.NumIn() != 3 {
			if reportErr {
				logger.Info("method %s has wrong number of ins: %v", mname, mtype.NumIn())
			}
			continue
		}

		idType := mtype.In(1)

		if !idType.AssignableTo(reflect.TypeOf((*RpcConn)(nil)).Elem()) {
			if reportErr {
				logger.Info("%s conn %s must be %s", mname, idType.Name(), reflect.TypeOf((*RpcConn)(nil)).Elem().Name())
			}
			continue
		}

		// Second arg need not be a pointer.
		argType := mtype.In(2)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				logger.Info("%s argument type not exported: %s", mname, argType)
			}
			continue
		}

		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				logger.Info("method %s has wrong number of outs: %v", mname, mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				logger.Info("method %s returns %s not error", mname, returnType.String())
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType}
	}
	return methods
}

func (m *methodType) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

func (s *service) call(server *Server, mtype *methodType, req *RequestWrap, argv reflect.Value, conn RpcConn) {
	mtype.Lock()
	mtype.numCalls++
	mtype.Unlock()
	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{s.rcvr, reflect.ValueOf(conn), argv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	errmsg := ""
	if errInter != nil {
		errmsg = errInter.(error).Error()
		server.sendErrorResponse(req, conn, errmsg)
	}
	server.freeRequest(req)
}

func (server *Server) CloseConn(id uint64) bool {
	server.connLock.Lock()
	conn, exist := server.connMap[id]
	server.connLock.Unlock()

	if exist {
		conn.Close()
		/*for _, v := range server.onDisConn {
			v(conn)
		}*/

		return true
	}

	return false
}

func (server *Server) GetConn(id uint64) RpcConn {
	server.connLock.RLock()
	conn, exist := server.connMap[id]
	server.connLock.RUnlock()

	if exist {
		return conn
	}
	return nil
}

func (server *Server) ServerBroadcast(data interface{}) {
	server.connLock.RLock()

	for _, conn := range server.connMap {
		conn.WriteObj(data)
	}
	server.connLock.RUnlock()
}

func (server *Server) Quit() {
	server.quitSync.Lock()
	server.quit = true
	server.quitSync.Unlock()
}

func (server *Server) ServeConn(conn RpcConn) {
	id := atomic.AddUint64(&server.id, 1)
	conn.SetId(id)

	server.connLock.Lock()
	server.connMap[id] = conn
	server.connLock.Unlock()
	for _, v := range server.onConn {
		v(conn)
	}

	for {

		server.quitSync.RLock()
		bQuit := server.quit
		server.quitSync.RUnlock()
		if bQuit {
			break
		}

		service, mtype, req, argv, keepReading, err := server.readRequest(conn)
		if err != nil {
			if e2, ok := err.(*net.OpError); ok && (e2.Timeout() || e2.Temporary()) {
				//logger.Info("Read timeout %v", e2) // This will happen frequently
				continue
			}

			if err != io.EOF {
				logger.Info("rpc: %s", err.Error())
			}
			if !keepReading {
				break
			}
			// send a response if we actually managed to read a header.
			if req != nil {
				server.sendErrorResponse(req, conn, err.Error())
				server.freeRequest(req)
			}
			continue
		}

		for _, v := range server.onCallBefore {
			v(conn)
		}

		service.call(server, mtype, req, argv, conn)

		for _, v := range server.onCallAfter {
			v(conn)
		}

	}

	for _, v := range server.onDisConn {
		v(conn)
	}

	conn.Close()

	server.connLock.Lock()
	delete(server.connMap, id)
	server.connLock.Unlock()

}

func (server *Server) sendErrorResponse(req *RequestWrap, conn RpcConn, errmsg string) {

	// Encode the response header

	resp := RpcErrorResponse{}

	resp.Method = req.Method
	resp.Text = &errmsg

	err := conn.WriteObj(resp)

	if err != nil {
		logger.Error("rpc: writing ErrorResponse: %s", err.Error())
		sysdebug.PrintStack()
	}
}

func (server *Server) readRequest(conn RpcConn) (service *service, mtype *methodType, req *RequestWrap, argv reflect.Value, keepReading bool, err error) {
	req = server.getRequest()
	err = conn.ReadRequest(&req.Request)

	if err != nil {
		req = nil
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}

		if e2, ok := err.(*net.OpError); ok && (e2.Timeout() || e2.Temporary()) {
			//logger.Info("Read timeout %v", e2) // This will happen frequently
			return
		}

		err = errors.New("rpc: server cannot decode request: " + err.Error())
		return
	}

	// We read the header successfully.  If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	serviceMethod := strings.Split(req.GetMethod(), ".")
	if len(serviceMethod) != 2 {
		err = errors.New("rpc: service/method request ill-formed: " + req.GetMethod())
		return
	}

	// Look up the request.
	server.mu.RLock()
	service = server.serviceMap[serviceMethod[0]]
	server.mu.RUnlock()
	if service == nil {
		err = errors.New("rpc: can't find service " + req.GetMethod())
		return
	}

	mtype = service.method[serviceMethod[1]]
	if mtype == nil {
		err = errors.New("rpc: can't find method " + req.GetMethod())
		return
	}

	// Decode the argument value.
	argIsValue := false // if true, need to indirect before calling.
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	// argv guaranteed to be a pointer now.
	if err = conn.GetRequestBody(&req.Request, argv.Interface()); err != nil {
		return
	}

	if argIsValue {
		argv = argv.Elem()
	}

	return
}

func (server *Server) getRequest() *RequestWrap {
	server.reqLock.Lock()
	req := server.freeReq
	if req == nil {
		req = new(RequestWrap)
	} else {
		server.freeReq = req.next
		*req = RequestWrap{}
	}
	server.reqLock.Unlock()
	return req
	//return new(RequestWrap)
}

func (server *Server) freeRequest(req *RequestWrap) {
	server.reqLock.Lock()
	req.next = server.freeReq
	server.freeReq = req
	server.reqLock.Unlock()
	req = nil
}

func (server *Server) HandleDebugHTTP(debugPath string) {
	http.Handle(debugPath, debugHTTP{server})
}

type RpcConn interface {
	ReadRequest(*Request) error

	Call(string, interface{}) error

	GetRemoteIp() string

	GetRequestBody(*Request, interface{}) error

	WriteObj(interface{}) error

	SetId(uint64)
	GetId() uint64

	Close() error

	Lock()
	Unlock()
}
