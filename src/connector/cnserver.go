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
	id              uint8
	serverForClient *rpc.Server
	listenIp        string
	listener        net.Listener
	lobbyService    *LobbyService
}

func NewCNServer(cfg *common.CnsConfig) (server *CNServer) {
	logger.Info("CNServer:NewCNServer:<ENTER>")
	//数据库服务
	dbclient.Init()
	var lobbycfg common.LobbyServerCfg
	if err := common.ReadLobbyConfig(&lobbycfg); err != nil {
		return
	}
	lobbyConn, err := net.Dial("tcp", lobbycfg.LobbyIpForServer)
	if err != nil {
		logger.Fatal("CNServer:NewCNServer:res:%s", err.Error())
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
	}

	server.RegisterReconnect()
	cns = server

	//loadConfigFiles(common.GetDesignerDir())
	logger.Info("CNServer:NewCNServer:<LEAVE>")
	return
}

func StartCenterService(self *CNServer, listener net.Listener, cfg *common.CnsConfig) {
	logger.Info("CNServer:StartCenterService:<ENTER>")
	self.listener = listener
	//注册其他节点连接端
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
	logger.Info("CNServer:StartCenterService:<LEAVE>")
}

func (self *CNServer) StartClientService(cfg *common.CnsConfig, wg *sync.WaitGroup) {
	logger.Info("CNServer:StartClientService:<ENTER>")
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

	self.sendPlayerCountToGateServer()

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
	logger.Info("CNServer:StartClientService:<LEAVE>")
}

func (c *CNServer) onConn(conn rpc.RpcConn) {
}

func (self *CNServer) onDisConn(conn rpc.RpcConn) {
	ts("CNServer:onDisConn", conn.GetId())
	logger.Info("CNServer:StartClientService:<ENTER>,connId:%d", conn.GetId())
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

func (self *CNServer) loadConfigFiles() {
	return
}

func (self *CNServer) GetServerId() uint8 {
	return self.id
}

func (self *CNServer) sendPlayerCountToGateServer() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("sendPlayerCountToGateServer runtime error:", r)

				debug.PrintStack()
			}
		}()

		for {

			time.Sleep(5 * time.Second)

			self.l.RLock()
			playerCount := uint32(3) //len(self.players))
			self.l.RUnlock()

			var ret proto.SendCnsInfoResult

			err := self.lobbyserver.Call("lobbyService.UpdateCnsPlayerCount", proto.SendCnsInfo{999, uint16(playerCount), self.listenIp}, &ret)

			if err != nil {
				logger.Error("Error On lobbyService.UpdateCnsPlayerCount : %s", err.Error())
				return
			}

		}

	}()
}

func (self *CNServer) RegisterReconnect() {
	//注册大厅重连机制
	self.lobbyserver.AddDisCallback(func(err error) {
		logger.Info("disconnected error:", err)
		self.ReConnectLobby()
	})
}
func (self *CNServer) ReConnectLobby() {
	logger.Info("CNServer:ReConnectLobby:<ENTER>")
	var lobbycfg common.LobbyServerCfg
	if err := common.ReadLobbyConfig(&lobbycfg); err != nil {
		return
	}
	lobbyConn, err := net.Dial("tcp", lobbycfg.LobbyIpForServer)
	if err != nil {
		logger.Fatal("%s", err.Error())
	}
	self.lobbyserver = rpcplus.NewClient(lobbyConn)
	req := &proto.CenterConnCns{Addr: self.listener.Addr().String()}
	rst := &proto.CenterConnCnsResult{}
	self.lobbyserver.Go("LobbyServices.LobbyConnCns", req, rst, nil)
	logger.Info("CNServer:ReConnectLobby:<LEAVE>")
	return
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
