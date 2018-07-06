package connector

import (
	"common"
	//	"errorValue"
	"logger"
	"rpc"
	"runtime"
	//"xzmj"
)

func (self *CNServer) Login(conn rpc.RpcConn, login rpc.Login) error {
	return self.login(conn, &login)
}

func (self *CNServer) login(conn rpc.RpcConn, login *rpc.Login) error {
	return nil
}
