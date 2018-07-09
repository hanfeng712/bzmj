package connector

import (
	"common"
	"fmt"
	"logger"
	"net"
	"os"
	"rpc"
	proto "rpc/proto"
	"rpcplus"
	"runtime/debug"
	"sync"
	"time"
)

const (
	ip   = ""
	port = 3333
)

var cns *CNServer
var Cfg common.CnsConfig

type LobbyService struct {
}

type CNServer struct {
	lobbyserver     *rpcplus.Client
	players         map[uint64]*player
	playersbyid     map[uint64]*player
	l               sync.RWMutex
	serverForClient *rpc.Server
	listenIp        string
	listener        net.Listener
	lobbyService    *LobbyService
}

/*
func NewCNServer() *CNServer {
	server := &CNServer{
		players:     make(map[uint64]*player),
		playersbyid: make(map[uint64]*player),
	}

	cns = server
	return cns
}
*/

func NewCNServer(cfg *common.CnsConfig) (server *CNServer) {
	//数据库服务

	//	dbclient.Init()
	var lobbycfg common.LobbyServerCfg
	if err := common.ReadLobbyConfig(&lobbycfg); err != nil {
		return
	}
	lobbyConn, err := net.Dial("tcp", lobbycfg.LobbyIpForServer)
	if err != nil {
		logger.Fatal("%s", err.Error())
	}
	/*
		var logCfg common.LogServerCfg
		if err := common.ReadLogConfig(&logCfg); err != nil {
			logger.Fatal("%v", err)
		}
		logConn, err := net.Dial("tcp", logCfg.LogHost)
		if err != nil {
			logger.Fatal("connect logserver failed %s", err.Error())
		}

			var chatcfg common.ChatServerCfg
			if err = common.ReadChatConfig(&chatcfg); err != nil {
				return
			}
			chatConn, err := net.Dial("tcp", chatcfg.ListenForServer)
			if err != nil {
				logger.Fatal("connect chatserver failed %s", err.Error())
			}
	*/
	server = &CNServer{
		lobbyserver: rpcplus.NewClient(lobbyConn),
		//logRpcConn:    rpcplus.NewClient(logConn),
		players:      make(map[uint64]*player),
		playersbyid:  make(map[uint64]*player),
		lobbyService: &LobbyService{},
		//chatRpcConn:   rpcplus.NewClient(chatConn),
		//rankMgr:       CreateRankMgr()
	}

	cns = server

	//loadConfigFiles(common.GetDesignerDir())

	return
}

func StartCenterService(self *CNServer, listener net.Listener, cfg *common.CnsConfig) {
	//连接center
	rpcLobbyServer := rpcplus.NewServer()
	rpcLobbyServer.Register(self.lobbyService)

	req := &proto.CenterConnCns{Addr: listener.Addr().String()}
	rst := &proto.CenterConnCnsResult{}
	self.lobbyserver.Go("LobbyServices.LobbyConnCns", req, rst, nil)

	connLobby, err := listener.Accept()
	if err != nil {
		logger.Error("StartCenterServices %s", err.Error())
		os.Exit(0)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("StartCenterService runtime error:", r)

				debug.PrintStack()
			}
		}()
		rpcLobbyServer.ServeConn(connLobby)
		connLobby.Close()
	}()

}

func (self *CNServer) StartClientService(cfg *common.CnsConfig, wg *sync.WaitGroup) {

	rpcServer := rpc.NewServer()
	self.serverForClient = rpcServer

	//lockclient.Init()
	//accountclient.Init()

	rpcServer.Register(cns)
	rpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			self.onConn(conn)
		},
	)

	rpcServer.RegCallBackOnDisConn(
		func(conn rpc.RpcConn) {
			self.onDisConn(conn)
		},
	)

	rpcServer.RegCallBackOnCallBefore(
		func(conn rpc.RpcConn) {
			conn.Lock()
		},
	)

	rpcServer.RegCallBackOnCallAfter(
		func(conn rpc.RpcConn) {
			conn.Unlock()
		},
	)

	listener, err := net.Listen("tcp", Cfg.CnsHost)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	self.listener = listener
	self.listenIp = cfg.CnsHostForClient

	//self.sendPlayerCountToGateServer()

	wg.Add(1) //监听client要算一个
	go func() {
		for {
			//For Client/////////////////////////////
			time.Sleep(time.Millisecond * 5)
			conn, err := self.listener.Accept()

			if err != nil {
				logger.Error("cns StartServices %s", err.Error())
				wg.Done() // 退出监听就要减去一个
				break
			}

			wg.Add(1) // 这里是给客户端增加计数
			go func() {
				rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 128, 45)
				defer func() {
					if r := recover(); r != nil {
						logger.Error("player rpc runtime error begin:", r)

						rpcConn.Unlock()
						debug.PrintStack()
						self.onDisConn(rpcConn)
						rpcConn.Close()

						logger.Error("player rpc runtime error end ")
					}
					wg.Done() // 客户端退出减去计数
				}()

				rpcServer.ServeConn(rpcConn)
			}()
		}
	}()
}

/*
func (self *CNServer) StartClientService(cfg *common.CnsConfig, wg *sync.WaitGroup) {
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
*/
func (c *CNServer) onConn(conn rpc.RpcConn) {
}

func (self *CNServer) onDisConn(conn rpc.RpcConn) {
	ts("CNServer:onDisConn", conn.GetId())
	defer te("CNServer:onDisConn", conn.GetId())

	self.delPlayer(conn.GetId())
}

func (self *CNServer) EndService() {
	self.lobbyserver.Close()
}

func (self *CNServer) Quit() {
	self.listener.Close()
	self.serverForClient.Quit()
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

/*
func (self *CNServer) AnswerClientError(conn rpc.RpcConn, value uint32) {
	logger.Info("player:AnswerClientError")
	var l uint32 = 1
	lCommonErrMsg := rpc.CSCommonErrMsg{}
	lCommonErrMsg.ErrorCode = &(value)
	lCommonErrMsg.RqstCmdID = &(l)

	common.WriteClientResult(conn, "SRobot.HandleErrorRsp", &lCommonErrMsg)
}

func (self *CNServer) SendMsgToPlayer(msg interface{}, uid uint64, method string) {
	self.l.RLock()
	defer self.l.RUnlock()
	logger.Info("CNServer:SendMsgToPlayer:uid:%d", uid)
	lPlayer := self.players[uid]

	lPlayer.SendMsgToClient(msg, method)
}
*/
