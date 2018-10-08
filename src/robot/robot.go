package robot

import (
	"common"
	"container/list"
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
	sendMsgCount    uint64
	conn            rpc.RpcConn
}

//var robotList *SRobot[]
var robotList *list.List = list.New()

const (
	addr = "127.0.0.1:7850"
)

func CreateRobot(id uint64) *SRobot {

	robot := &SRobot{}
	robotList.PushBack(robot)
	robot.uid = id
	robot.sendMsgCount = 1
	robot.stateMatchine = fsm.CreateFSM()
	key1 := "connectserver"
	matchine1 := fsm.CreateMatchineState(key1, nil, key1, robot, "ConnectGameServer")
	robot.stateMatchine.AddState(key1, matchine1)

	key2 := "pinggame"
	matchine2 := fsm.CreateMatchineState(key2, nil, key2, robot, "SendPing")
	robot.stateMatchine.AddState(key2, matchine2)
	robot.stateMatchine.SetDefaultState(matchine2)
	robot.stateMatchine.Start()
	logger.Debug("id:%d", id)

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
	/*
		logger.Debug("SendPing:uid:%d,sendMsgCount:%d", self.uid, self.sendMsgCount)
		count := self.sendMsgCount + uint64(1)
		pingReq := rpc.Ping{}
		pingReq.Id = &self.uid
		pingReq.Count = &(count)
		sendMsg(self.conn, &pingReq)
		self.sendMsgCount = count
	*/
	return
}
func sendMsg(conn rpc.RpcConn, value interface{}) {
	common.WriteClientResult(conn, "LobbyServicesForClient.LobbyHandlePingMsg", value)
}
