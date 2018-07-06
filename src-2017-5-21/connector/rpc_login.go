package connector

import (
	"common"
	"errorValue"
	"fmt"
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
	fmt.Printf("recv login\n")
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
	self.testCreateRoom(login)
	LoadPlayer(lUid)
	loginReq := rpc.LoginRsp{}
	loginReq.Uid = &lUid
	common.WriteClientResult(conn, "SRobot.HandleLoginRsp", &loginReq)

	return nil
}

func (self *CNServer) testCreateRoom(login *rpc.Login) error {
	logger.Info("TestCreateRoom<ENTER>")

	var lRoomId uint64 = 0
	if self.xzmjRoomMgr == nil {
		logger.Info("TestCreateRoom<after>,roomMgr is nil")
		return nil
	}
	self.xzmjRoomMgr.CreateRoom(&lRoomId)
	self.xzmjRoomMgr.HandleRoomToTableMsg(lRoomId, login)

	return nil
}

func (self *CNServer) CreateRoom(conn rpc.RpcConn, createRoomRqst *rpc.CSUserCreateRoomRqst) error {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	logger.Info("CreateRoom<ENTER>")

	var lRoomId uint64 = 0
	if self.xzmjRoomMgr == nil {
		logger.Info("error<after>,roomMgr is nil")
		return nil
	}
	lRet = self.xzmjRoomMgr.CreateRoom(&lRoomId)
	if lRet != errorValue.ERET_OK {
		//TODO:send error msg to client
		return nil
	}

	lRoom := self.xzmjRoomMgr.GetRoomById(lRoomId)
	lRet := lRoom.AddPlayerToRoom(login.GetUid(), createRoomRqst)
	if lRet != errorValue.ERET_OK {
		//TODO:resposen user create room
		return nil
	}

	return nil
}

func (self *CNServer) EnterRoom(conn rpc.RpcConn, login *rpc.CSUserEnterRoomRsp) error {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	return lRet
}

func (self *CNServer) ReadyGame(conn rpc.RpcConn, login *rpc.CSUserReadyGameRqst) error {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	return lRet
}
