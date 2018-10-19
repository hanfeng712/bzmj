package rpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"logger"
	"net"
	"reflect"
	"sync"
	"time"
	"timer"

	"github.com/golang/snappy"

	"github.com/golang/protobuf/proto"
)

type ProtoBufConn struct {
	c            net.Conn
	id           uint64
	send         chan *Request
	t            *timer.Timer
	exit         chan bool
	last_time    int64
	time_out     uint32
	lockForClose sync.Mutex
	is_closed    bool
	sync.Mutex
	connMgr       *Server
	isCompression bool
}

func NewProtoBufConn(server *Server, c net.Conn, size int32, k uint32, isCompression bool) (conn RpcConn) {
	pbc := &ProtoBufConn{
		c:             c,
		send:          make(chan *Request, size),
		exit:          make(chan bool, 1),
		last_time:     time.Now().Unix(),
		time_out:      k,
		connMgr:       server,
		isCompression: isCompression,
	}

	if k > 0 {
		pbc.t = timer.NewTimer(time.Duration(k) * time.Second)
		pbc.t.Start(
			func() {
				pbc.OnCheck()
			},
		)
	}

	go pbc.mux()
	return pbc
}

func (conn *ProtoBufConn) OnCheck() {
	time_diff := uint32(time.Now().Unix() - conn.last_time)
	if time_diff > conn.time_out<<1 {
		//logger.Info("Conn %d TimeOut: %d", conn.GetId(), time_diff)
		//conn.connMgr.CloseConn(conn.GetId())
		conn.Close()
	}
}

func (conn *ProtoBufConn) mux() {
	for {
		select {
		case r := <-conn.send:
			buf, err := proto.Marshal(r)
			if err != nil {
				logger.Error("ProtoBufConn Marshal Error %s", err.Error())
				continue
			}
			dst := buf
			if conn.isCompression == true {
				dst = snappy.Encode(nil, buf)
			}
			if err != nil {
				logger.Error("ProtoBufConn snappy.Encode Error %s", err.Error())
				continue
			}

			conn.c.SetWriteDeadline(time.Now().Add(ConnWriteTimeOut))
			err = binary.Write(conn.c, binary.BigEndian, int32(len(dst)))
			if err != nil {
				//logger.Error("ProtoBufConn Write Error %s", err.Error())
				continue
			}

			conn.c.SetWriteDeadline(time.Now().Add(ConnWriteTimeOut))
			_, err = conn.c.Write(dst)
			if err != nil {
				//logger.Error("ProtoBufConn Write Error %s", err.Error())
				continue
			}
		case <-conn.exit:
			return
		}
	}
}

func (conn *ProtoBufConn) GetRemoteIp() string {
	return conn.c.RemoteAddr().String()
}

func (conn *ProtoBufConn) ReadRequest(req *Request) error {
	var size uint32

	conn.c.SetReadDeadline(time.Now().Add(ConnReadTimeOut))

	err := binary.Read(conn.c, binary.BigEndian, &size)
	if err != nil {
		return err
	}
	fmt.Printf("Recv Len:%d\n", size)
	buf := make([]byte, size)

	conn.c.SetReadDeadline(time.Now().Add(ConnReadTimeOut))

	_, err = io.ReadFull(conn.c, buf)
	if err != nil {
		return err
	}

	dst := buf
	if conn.isCompression == true {
		dst, err = snappy.Decode(nil, buf)

		if err != nil {
			return err
		}
	}

	if conn.isCompression == true {
		dst, err = snappy.Decode(nil, buf)

		if err != nil {
			return err
		}
	}
	conn.last_time = time.Now().Unix()

	return proto.Unmarshal(dst, req)
}

func (conn *ProtoBufConn) GetRequestBody(req *Request, body interface{}) error {
	if value, ok := body.(proto.Message); ok {
		return proto.Unmarshal(req.GetSerializedRequest(), value)
	}

	return fmt.Errorf("value type error %v", body)
}

func (conn *ProtoBufConn) writeRequest(r *Request) error {
	conn.send <- r
	return nil
}

func (conn *ProtoBufConn) Call(serviceMethod string, args interface{}) error {
	var msg proto.Message

	switch m := args.(type) {
	case proto.Message:
		msg = m
	default:
		return fmt.Errorf("Call args type error %v", args)
	}

	buf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	req := &Request{}
	req.Method = &serviceMethod
	req.SerializedRequest = buf

	return conn.writeRequest(req)
}

func (conn *ProtoBufConn) WriteObj(value interface{}) error {
	var msg proto.Message

	switch m := value.(type) {
	case proto.Message:
		msg = m
	default:
		return fmt.Errorf("WriteObj value type error %v", value)
	}

	buf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	req := &Request{}

	t := reflect.Indirect(reflect.ValueOf(msg)).Type()
	//req.SetMethod(t.PkgPath() + "." + t.Name())
	mehodValue := t.PkgPath() + "." + t.Name()
	req.Method = &(mehodValue)
	req.SerializedRequest = buf

	return conn.writeRequest(req)
}

func (conn *ProtoBufConn) SetId(id uint64) {
	conn.id = id
}

func (conn *ProtoBufConn) GetId() uint64 {
	return conn.id
}

func (conn *ProtoBufConn) Close() (errret error) {
	conn.lockForClose.Lock()

	if conn.is_closed {
		conn.lockForClose.Unlock()
		return nil
	}

	if err := conn.c.Close(); err != nil {
		//设置超时
		if err := conn.c.SetDeadline(time.Now()); err != nil {
			conn.lockForClose.Unlock()
			return err
		}

		//再尝试一次
		time.Sleep(10 * time.Millisecond)
		if err := conn.c.Close(); err != nil {
			conn.lockForClose.Unlock()
			return err
		}
	}

	conn.is_closed = true

	if conn.t != nil {
		conn.t.Stop()
	}

	conn.exit <- true

	conn.lockForClose.Unlock()

	return nil
}
