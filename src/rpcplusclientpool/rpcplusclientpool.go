package rpcplusclientpool

import (
	"net"
	"rpcplus"
	"logger"
	"math/rand"
	"sync"
	"time"
	"errors"
)

var NoServiceError = errors.New("no service, please wait.")

type CALLBACK func(conn *rpcplus.Client)

//加锁服务器的连接
type ServerInfo struct {
	Shost string
	Conn *rpcplus.Client
	Breconn bool
}

type ClientPool struct {
	//服务器列表
	aServerList []*ServerInfo
	//有效的服务器索引
	aValidServer []int
	l sync.RWMutex
	//连接成功失败的回调
	okcallback CALLBACK
	discallback CALLBACK
	//是否关闭
	bClose bool
}

//创建
func CreateClientPool(aServerHost []string) *ClientPool {
	pool := &ClientPool {
		aServerList : make([]*ServerInfo, len(aServerHost)),
		aValidServer : make([]int, 0),
		bClose : false,
	}
	
	for i, v := range aServerHost {
		if err := pool.addConnect(v, i); err != nil {
			logger.Fatal("dail lockserver failed", err)
			return nil
		}
	}
	
	return pool
}

//添加一个连接
func (self *ClientPool) addConnect(sHost string, iIndex int) error {
	if self.bClose {
		return nil
	}
	
	conn, err := net.Dial("tcp", sHost)
	if err != nil {
		return err
	}
	
	self.l.Lock()
	defer self.l.Unlock()
	
	if self.bClose {
		conn.Close()
		return nil
	}
	
	rpc := rpcplus.NewClient(conn)
	rpc.AddDisCallback(func(err error) {
		logger.Info("disconnected error:", err)
		
		self.reConnect(iIndex)
	})
	
	info := &ServerInfo {
		Shost : sHost,
		Conn : rpc,
		Breconn : false,
	}
	
	if self.okcallback != nil {
		go self.okcallback(rpc)
	}
	
	self.aServerList[iIndex] = info
	self.aValidServer = append(self.aValidServer, iIndex)
	
	//logger.Info("aServerList numbers:", len(self.aValidServer))
	
	return nil
}

//重新连接
func (self *ClientPool) reConnect(iIndex int) {
	self.l.Lock()
	defer self.l.Unlock()
	
	if iIndex >= len(self.aServerList) {
		return
	}
	
	//重复调用的情况
	info := self.aServerList[iIndex]
	//过期的多次传入
	if info.Breconn {
		return
	}
	
	if self.discallback != nil {
		go self.discallback(info.Conn)
	}
	
	logger.Info("reconnect...", iIndex)
	
	info.Breconn = true
	info.Conn.Close()
	for i, v := range self.aValidServer {
		if v == iIndex {
			self.aValidServer = append(self.aValidServer[:i], self.aValidServer[i + 1 :]...)
			break
		}
	}
	
	//重连接
	go func (sHost string) {
		for {
			if err := self.addConnect(sHost, iIndex); err == nil {
				break
			}
			
			time.Sleep(time.Second * 3)
		}
	}(info.Shost)
}

//随机取一个连接，后面根据负载来处理
func (self *ClientPool) RandomGetConn() (err error, conn *rpcplus.Client) {
	self.l.RLock()
	defer self.l.RUnlock()
	
	if len(self.aValidServer) == 0 {
		err = NoServiceError
		return
	}
	
	index := self.aValidServer[rand.Intn(len(self.aValidServer))]
	info := self.aServerList[index]
	if info.Breconn {
		err = NoServiceError
		return
	}
	conn = info.Conn
	
	return
}

//取得所有的连接
func (self *ClientPool) GetAllConn() []*rpcplus.Client {
	connlist := make([]*rpcplus.Client, 0)
	
	self.l.RLock()
	defer self.l.RUnlock()
	
	for _, index := range(self.aValidServer) {
		connlist = append(connlist, self.aServerList[index].Conn)
	}
	
	return connlist
}

//连接成功的回调
func (self *ClientPool) SetConnectedCallback(f CALLBACK) {
	self.okcallback = f
}

//连接断开的回调
func (self *ClientPool) SetDisconnectCallback(f CALLBACK) {
	self.discallback = f
}

//关闭所有连接
func (self *ClientPool) CloseAll() {
	self.l.Lock()
	
	self.bClose = true
	for _, info := range(self.aServerList) {
		info.Conn.Close()
	}
	self.aValidServer = make([]int, 0)
	
	self.l.Unlock()
}