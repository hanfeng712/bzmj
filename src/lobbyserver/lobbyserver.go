package lobbyserver

import (
	//"fmt"
	"logger"
	//	"math/rand"
	"net"
	"rpc"
	"rpcplus"

	"rpc/proto"
	//	"strconv"
	"common"
	"dbclient"
	"net/http"
	"sync"
	//"time"

	"golang.org/x/net/websocket"
)

type serverInfo struct {
	PlayerCount uint16
	ServerIp    string
}

type LobbyServices struct {
	l            sync.RWMutex
	lgs          *rpcplus.Client       //日志服务器的连接段
	m            map[uint32]serverInfo //维护每个节点服务器的在线人数
	cnss         []*rpcplus.Client     //维护对每个节点的连接
	maincache    *common.CachePool
	clancache    *common.CachePool
	stableServer string //在线人数最小的节点服务器
}

var lobbyService *LobbyServices

//创建大厅对象
func NewLobbyServer(cfg common.LobbyServerCfg) (server *LobbyServices) {
	//数据库服务

	dbclient.Init()

	var logCfg common.LogServerCfg
	if err := common.ReadLogConfig(&logCfg); err != nil {
		logger.Fatal("%v", err)
	}
	logConn, err := net.Dial("tcp", logCfg.LogHost)
	if err != nil {
		logger.Fatal("connect logserver failed %s", err.Error())
	}
	server = &LobbyServices{
		lgs:  rpcplus.NewClient(logConn),
		cnss: make([]*rpcplus.Client, 0, 1),
		m:    make(map[uint32]serverInfo),
	}
	/*
		//初始化cache
		logger.Info("Init Cache %v", cfg.MainCacheProfile)
		server.maincache = common.NewCachePool(cfg.MainCacheProfile)

		logger.Info("Init Cache %v", cfg.ClanCacheProfile)
		server.clancache = common.NewCachePool(cfg.ClanCacheProfile)
	*/
	return server
}

//创建其他节点服务器连接服务器
var pLobbyServices *LobbyServices

func CreateLobbyServicesForCnserver(server *LobbyServices, listener net.Listener) *LobbyServices {
	pLobbyServices = server
	rpcServer := rpcplus.NewServer()

	rpcServer.Register(pLobbyServices)

	//rpcServer.HandleHTTP("/center/rpc", "/debug/rpcdebug/rpc")

	var uConnId uint32 = 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gateserver StartServices %s", err.Error())
			break
		}
		logger.Debug("other server connect lobby")
		uConnId++
		go func(uConnId uint32) {

			pLobbyServices.l.Lock()
			pLobbyServices.m[uConnId] = serverInfo{0, ""}
			pLobbyServices.l.Unlock()

			rpcServer.ServeConnWithContext(conn, uConnId)

			pLobbyServices.l.Lock()
			delete(pLobbyServices.m, uConnId)
			pLobbyServices.l.Unlock()

		}(uConnId)
	}

	return pLobbyServices
}

func (self *LobbyServices) LobbyConnCns(req *proto.CenterConnCns, reply *proto.CenterConnCnsResult) (err error) {
	logger.Info("Center:CenterConnCns:%s", req.Addr)

	conn, err := net.Dial("tcp", req.Addr)
	if err != nil {
		logger.Fatal("%s", err.Error())
		reply.Ret = false
		return
	}

	tmp := rpcplus.NewClient(conn)
	self.l.Lock()
	self.cnss = append(self.cnss, tmp)
	//self.theFirstUpdate(tmp)
	self.l.Unlock()
	reply.Ret = true

	return nil
}

//更新每个服务器的在线人数
func (self *LobbyServices) UpdateCnsPlayerCount(uConnId uint32, info *proto.SendCnsInfo, result *proto.SendCnsInfoResult) error {
	logger.Debug("LobbyServices:UpdateCnsPlayerCount:<ENTER>")
	self.l.Lock()
	self.m[uConnId] = serverInfo{info.PlayerCount, info.ServerIp}

	playerCountMax := uint16(0xffff) //不会有哪个服务器更大吧
	self.stableServer = ""
	for _, v := range self.m {
		if len(v.ServerIp) > 0 && v.PlayerCount < playerCountMax {
			playerCountMax = v.PlayerCount
			self.stableServer = v.ServerIp
		}
	}

	self.l.Unlock()
	logger.Debug("LobbyServices:UpdateCnsPlayerCount:<LEAVE>")
	//fmt.Printf("recv cns msg : server %d , player count %d, player ip = %s \n", info.ServerId, info.PlayerCount, info.ServerIp)
	return nil
}

func (self *LobbyServices) getStableCns() (cnsIp string) {
	self.l.RLock()
	defer self.l.RUnlock()
	return self.stableServer
}

type LobbyServicesForClient struct {
	m string
}

//创建客户端连接服务器
var lobbyServicesForClient *LobbyServicesForClient
var rpcServer *rpc.Server

func CreateLobbyServicesForClient(addr string, connType string) *LobbyServicesForClient {
	logger.Debug("client:", addr, "connType:", connType)
	lobbyServicesForClient = &LobbyServicesForClient{}
	rpcServer = rpc.NewServer()
	rpcServer.Register(lobbyServicesForClient)

	rpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			lobbyServicesForClient.onConn(conn)
		},
	)

	if connType == "webConn" {
		http.Handle("/", websocket.Handler(webConnHandler))
		err := http.ListenAndServe(":7850", nil)
		if err != nil {
			println("Listening to: ", addr, " failed !!")
			return nil
		}
	} else if connType == "tcpConn" {
		listenerForClient, err := net.Listen("tcp", addr)
		defer listenerForClient.Close()
		if err != nil {
			println("Listening to: ", addr, " failed !!")
			return nil
		}
		tcpConnHandler(listenerForClient)
	}

	return lobbyServicesForClient
}

//webSocket
/*
func webConnHandler(conn *websocket.Conn) {
	for {
		go func() {
			logger.Debug("client connect lobby")
			rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 4, 0)
			rpcServer.ServeConn(rpcConn)
		}()
	}
}
*/

func webConnHandler(conn *websocket.Conn) {
	defer conn.Close()
	logger.Debug("client connect lobby")
	rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 4, 0, false)
	rpcServer.ServeConn(rpcConn)
}

//tcpScoket
func tcpConnHandler(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gateserver StartServices %s", err.Error())
			break
		}
		go func() {
			logger.Debug("client connect lobby")
			rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 4, 0, false)
			rpcServer.ServeConn(rpcConn)
		}()
	}
}

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	err := conn.WriteObj(value)
	if err != nil {
		logger.Info("WriteResult Error %s", err.Error())
		return false
	}
	return true
}

func SendMsgToClient(conn rpc.RpcConn, value interface{}, fun string) bool {
	logger.Info("SendMsgToClient")
	common.WriteClientResult(conn, fun, value)
	return true
}
func (c *LobbyServicesForClient) onConn(conn rpc.RpcConn) {
}

func (c *LobbyServicesForClient) LobbyHandlePingMsg(conn rpc.RpcConn, msg rpc.CS_BetMsg) error {
	logger.Debug("LobbyHandlePingMsg:recv CS_BetMsg")
	betRsp := rpc.SC_BetMsg{}
	SendMsgToClient(conn, &betRsp, "hanfeng")
	return nil
}
