package robot

import (
	//	"common"
	//	"fmt"
	"fsm"
	"logger"
	//	"net"
	"rpc"
	//	"runtime/debug"
)

type SRobot struct {
	uid             uint64
	serverForClient *rpc.Server
	stateMatchine   *fsm.FSM
	conn            rpc.RpcConn
}

var robot *SRobot

const (
	addr = "127.0.0.1:8850"
)

func CreateRobot() *SRobot {
	robot := &SRobot{}
	robot.stateMatchine = fsm.CreateFSM()

	key1 := "connectserver"
	matchine1 := fsm.CreateMatchineState(key1, nil, key1, robot, "ConnectGameServer")
	robot.stateMatchine.AddState(key1, matchine1)

	key2 := "pinggame"
	matchine2 := fsm.CreateMatchineState(key2, nil, key2, robot, "SendPing")
	robot.stateMatchine.AddState(key1, matchine2)
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
	self.conn = rpcConn
}

func (c *SRobot) onConn(conn rpc.RpcConn) {
	logger.Info("onConn")
}

func (self *SRobot) onDisConn(conn rpc.RpcConn) {
	logger.Info("onDisConn")
}

func (self *SRobot) SendLoginMsg(conn rpc.RpcConn) {
	logger.Info("SendLoginMsg")

	return
}

func (self *SRobot) SendPing(conn rpc.RpcConn) {
	logger.Info("SendPing")
	return
}
func SendMsg(conn rpc.RpcConn, value interface{}) {
	logger.Info("SendMsg")
	common.WriteClientResult(conn, "CNServer.Login", value)
}
