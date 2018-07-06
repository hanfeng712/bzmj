package connector

import (
	//"common"
	"logger"
	"net"
	"rpc"
	"runtime/debug"
	"sync"
	"time"
)

const (
	ip   = ""
	port = 3333
)

var cns *CNServer

type CNServer struct {
	players     map[uint64]*player
	playersbyid map[uint64]*player

	serverForClient *rpc.Server
	listenIp        string
	listener        net.Listener
	l               sync.RWMutex
}

func NewCNServer() *CNServer {
	server := &CNServer{
		players:     make(map[uint64]*player),
		playersbyid: make(map[uint64]*player),
	}

	cns = server

	return cns
}

func (self *CNServer) StartClientService(l int, wg *sync.WaitGroup) {
	lListerIp := "127.0.0.1:5300"

	lRpcServer := rpc.NewServer()
	self.serverForClient = lRpcServer
	lRpcServer.Register(cns)

	lRpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			self.onConn(conn)
		},
	)

	lRpcServer.RegCallBackOnDisConn(
		func(conn rpc.RpcConn) {
			self.onDisConn(conn)
		},
	)

	lRpcServer.RegCallBackOnCallBefore(
		func(conn rpc.RpcConn) {
			conn.Lock()
		},
	)

	lRpcServer.RegCallBackOnCallAfter(
		func(conn rpc.RpcConn) {
			conn.Unlock()
		},
	)

	listener, err := net.Listen("tcp", "127.0.0.1:7900")
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	self.listener = listener
	self.listenIp = lListerIp

	wg.Add(1) //监听client要算一个
	go func() {
		for {
			time.Sleep(time.Millisecond * 5)
			conn, err := self.listener.Accept()
			logger.Info("Accept one")
			if err != nil {
				logger.Error("cns StartServices %s", err.Error())
				wg.Done() //退出监听就要减去一个
			}

			wg.Add(1) // 这里是给客户端增加计数
			go func() {
				rpcConn := rpc.NewProtoBufConn(lRpcServer, conn, 128, 45)
				defer func() {
					if r := recover(); r != nil {
						logger.Error("player rpc runtime error begin:", r)
						debug.PrintStack()
						self.onDisConn(rpcConn)
						rpcConn.Close()

						logger.Error("player rpc runtime error end ")
					}
					wg.Done() // 客户端退出减去计数
				}()
				lRpcServer.ServeConn(rpcConn)
			}()
		}
	}()
}

func (c *CNServer) onConn(conn rpc.RpcConn) {
}

func (self *CNServer) onDisConn(conn rpc.RpcConn) {
	ts("CNServer:onDisConn", conn.GetId())
	defer te("CNServer:onDisConn", conn.GetId())

	self.delPlayer(conn.GetId())
}

//销毁玩家
func (self *CNServer) delPlayer(connId uint64) {
	ts("CNServer:delPlayer", connId)
	defer te("CNServer:delPlayer", connId)

	p, exist := self.players[connId]
	if exist {
		p.OnQuit()

		//self.l.Lock()
		delete(self.players, connId)
		delete(self.playersbyid, p.GetUid())
		//self.l.Unlock()
	}
}
