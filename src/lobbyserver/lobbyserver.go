package lobbyserver

import (
	"fmt"
	"logger"
	//	"math/rand"
	"net"
	"rpc"
	"rpcplus"

	"rpc/proto"
	//	"strconv"
	"common"
	//"dbclient"
	"sync"
	"time"
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
	/*
		dbclient.Init()
		var logCfg common.LogServerCfg
		if err := common.ReadLogConfig(&logCfg); err != nil {
			logger.Fatal("%v", err)
		}
		logConn, err := net.Dial("tcp", logCfg.LogHost)
		if err != nil {
			logger.Fatal("connect logserver failed %s", err.Error())
		}
	*/
	server = &LobbyServices{
		//	lgs:  rpcplus.NewClient(logConn),
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
	logger.Info("LobbyServices:UpdateCnsPlayerCount:<ENTER>")
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
	logger.Info("LobbyServices:UpdateCnsPlayerCount:<LEAVE>")
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

func CreateLobbyServicesForClient(listener net.Listener) *LobbyServicesForClient {

	lobbyServicesForClient = &LobbyServicesForClient{}
	rpcServer := rpc.NewServer()
	rpcServer.Register(lobbyServicesForClient)

	rpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			lobbyServicesForClient.onConn(conn)
		},
	)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gateserver StartServices %s", err.Error())
			break
		}
		go func() {
			rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 4, 0)
			rpcServer.ServeConn(rpcConn)
		}()
	}

	return lobbyServicesForClient
}

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	err := conn.WriteObj(value)
	if err != nil {
		logger.Info("WriteResult Error %s", err.Error())
		return false
	}
	return true
}

func (c *LobbyServicesForClient) onConn(conn rpc.RpcConn) {
	rep := rpc.LoginCnsInfo{}

	cnsIp := pLobbyServices.getStableCns()
	rep.CnsIp = &cnsIp
	gasinfo := fmt.Sprintf("%s;%d", conn.GetRemoteIp(), time.Now().Unix())
	logger.Info("Client(%s) -> CnServer(%s)", conn.GetRemoteIp(), cnsIp)
	// encode
	encodeInfo := common.Base64Encode([]byte(gasinfo))

	gasinfo = fmt.Sprintf("%s;%s", gasinfo, encodeInfo)

	//fmt.Printf("%s \n", gasinfo)

	rep.GsInfo = &gasinfo

	WriteResult(conn, &rep)

	time.Sleep(10 * time.Second)
	conn.Close()
}
