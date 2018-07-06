package connector

import (
	"common"
	"fmt"
	"rpc"
)

func (self *CNServer) Login(conn rpc.RpcConn, login rpc.Login) error {
	return self.login(conn, &login)
}

func (self *CNServer) login(conn rpc.RpcConn, login *rpc.Login) error {
	fmt.Print("recv login")
	lUid := login.GetUid()
	lCount := 1
	if lUid == uint64(0) {
		for {
			lUid = GenUUid()
			if self.players[lUid] == nil {
				break
			}
			lUid = 0
			if lCount >= 10 {
				break
			}
			lCount += 1
		}
	}
	if lUid == 0 {
		//生成Uid失败，返回客户端错误码
		return nil
	}
	LoadPlayer(lUid)

	loginReq := rpc.LoginRsp{}
	loginReq.Uid = &lUid
	common.WriteClientResult(conn, "SRobot.HandleLoginRsp", &loginReq)

	return nil
}

func (self *CNServer) CreateRoom(conn rpc.RpcConn, login *rpc.Login) error {

}

func (self *CNServer) EnterRoom(conn rpc.RpcConn, login *rpc.Login) error {

}

func (self *CNServer) ReadyGame(conn rpc.RpcConn, login *rpc.Login) error {

}
