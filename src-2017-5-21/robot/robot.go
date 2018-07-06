package robot

import (
	"common"
	"fmt"
	"logger"
	"net"
	"rpc"
	"runtime/debug"
)

type SRobot struct {
	uid             uint64
	serverForClient *rpc.Server
}

var robot *SRobot

const (
	addr = "127.0.0.1:7900"
)

func CreateRobot() *SRobot {
	robot := &SRobot{}
	return robot
}

func (self *SRobot) ConnectGameServer(l int, k int) {

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

}

func (c *SRobot) onConn(conn rpc.RpcConn) {
	logger.Info("onConn")
	SendLoginMsg(conn, 1111)
}

func (self *SRobot) onDisConn(conn rpc.RpcConn) {
	logger.Info("onDisConn")
}

func SendLoginMsg(conn rpc.RpcConn, uid uint64) {
	logger.Info("SendLoginMsg")
	loginReq := rpc.Login{}
	loginReq.Uid = &uid

	SendMsg(conn, &loginReq)
	return
}

func SendMsg(conn rpc.RpcConn, value interface{}) {
	logger.Info("SendMsg")
	common.WriteClientResult(conn, "CNServer.Login", value)
}
