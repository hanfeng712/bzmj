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
	//var lRet uint32 = uint32(errorValue.ERET_OK)
	logger.Info("recv login\n")
	lUid := login.GetUid()
	lCount := 1
	if lUid == uint64(0) {
		for {
			lUid = common.GenUUid()
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
		//lRet = uint32(errorValue.ERET_GENERATE_CUSTOM_ROOM_ID)
		return nil
	}
	_, file, line, _ := runtime.Caller(1)
	logger.Info("%s,line:%d,uid:%d", file, line, lUid)
	lPlayer := LoadPlayer(lUid, conn)
	self.l.RLock()
	self.players[lUid] = lPlayer
	loginRsp := rpc.LoginRsp{}
	loginRsp.Uid = &lUid
	common.WriteClientResult(conn, "SRobot.HandleLoginRsp", &loginRsp)

	return nil
}

func (self *CNServer) CreateRoom(conn rpc.RpcConn, createRoomRqst *rpc.CSUserCreateRoomRqst) error {
	return nil
}

func (self *CNServer) EnterRoom(conn rpc.RpcConn, enterRoom *rpc.CSUserEnterRoomRqst) error {
	return nil
}

func (self *CNServer) ReadyGame(conn rpc.RpcConn, readyMsg *rpc.CSUserReadyGameRqst) error {
	return nil
}
