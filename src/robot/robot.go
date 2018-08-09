package robot

import (
	"common"
	"fmt"
	"fsm"
	"logger"
	"net"
	"rpc"
	//"runtime/debug"
)

type SRobot struct {
	uid             uint64
	serverForClient *rpc.Server
	stateMatchine   *fsm.FSM
	conn            rpc.RpcConn
}

var robot *SRobot
var sendMsgCount uint64 = 1

const (
	addr = "127.0.0.1:7850"
)

func CreateRobot(id uint64) *SRobot {
	robot := &SRobot{}
	robot.uid = id
	robot.stateMatchine = fsm.CreateFSM()

	key1 := "connectserver"
	matchine1 := fsm.CreateMatchineState(key1, nil, key1, robot, "ConnectGameServer")
	robot.stateMatchine.AddState(key1, matchine1)

	key2 := "pinggame"
	matchine2 := fsm.CreateMatchineState(key2, nil, key2, robot, "SendPing")
	robot.stateMatchine.AddState(key2, matchine2)
	robot.stateMatchine.SetDefaultState(matchine2)
	robot.stateMatchine.Start()
	return robot
}

func (self *SRobot) ConnectGameServer(key string) {
	logger.Info("ConnectGameServer")

	lRpcServer := rpc.NewServer()
	self.serverForClient = lRpcServer
	lRpcServer.Register(self)

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

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("连接服务端失败:", err.Error())
		return
	}
	rpcConn := rpc.NewProtoBufConn(lRpcServer, conn, 128, 45)
	lRpcServer.ServeConn(rpcConn)
	/*
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
			}()
			lRpcServer.ServeConn(rpcConn)
		}()
	*/
}

func (self *SRobot) onConn(conn rpc.RpcConn) {
	logger.Info("=====onConn======")
	self.conn = conn
	self.stateMatchine.SwitchFsmState()
}

func (self *SRobot) onDisConn(conn rpc.RpcConn) {
	logger.Debug("onDisConn")
}

func (self *SRobot) sendLoginMsg(conn rpc.RpcConn) {
	logger.Debug("SendLoginMsg")

	return
}

func (self *SRobot) SendPing(key string) { //conn rpc.RpcConn) {
	logger.Debug("SendPing")

	pingReq := rpc.Ping{}
	pingReq.Id = &self.uid
	pingReq.Count = &sendMsgCount
	sendMsg(self.conn, &pingReq)
	return
}
func sendMsg(conn rpc.RpcConn, value interface{}) {
	common.WriteClientResult(conn, "LobbyServicesForClient.LobbyHandlePingMsg", value)
}
