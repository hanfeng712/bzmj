package connector

import (
	//	"common"
	"fmt"
	"logger"
	"rpc"
)

type player struct {
	mUid  uint64
	mConn rpc.RpcConn
}

func LoadPlayer(uid uint64, conn rpc.RpcConn) *player {
	lPlayer := NewPlayer(uid)
	lPlayer.mConn = conn
	return lPlayer
}

func NewPlayer(uid uint64) *player {
	lRet := &player{mUid: uid}
	return lRet
}

func (p *player) GetUid() uint64 {
	return p.mUid
}

func (p *player) OnQuit() {
	ts("player:OnQuit", p.GetUid())
	defer te("player:OnQuit", p.GetUid())
	fmt.Println("退出 ")

	if p.mConn != nil {
		p.mConn.Lock()
		defer p.mConn.Unlock()
	}

	logger.Info("OnQuit p.conn.Lock() end")
}

/*
func (self *player) AnswerClientError(value uint32) {
	logger.Info("player:AnswerClientError")
	var l uint32 = 1
	lCommonErrMsg := rpc.CSCommonErrMsg{}
	lCommonErrMsg.ErrorCode = &(value)
	lCommonErrMsg.RqstCmdID = &(l)

	self.SendMsgToClient(&lCommonErrMsg, "SRobot.HandleErrorRsp")
}

func (self *player) SendMsgToClient(msg interface{}, method string) {
	logger.Info("player:SendMsgToClient;uid:%d", self.GetUid())
	common.WriteClientResult(self.mConn, method, msg)
}
*/
