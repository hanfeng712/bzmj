package robot

import (
	"logger"
	"rpc"
)

func (self *SRobot) HandlePongRsp(conn rpc.RpcConn, msg rpc.Pong) error {
	logger.Debug("recv HandlePongRsp :")
	//uid:%d, count:%d", self.uid, msg.Count)
	self.stateMatchine.SwitchFsmState()
	return nil
}

func (self *SRobot) LoginCnsInfo(conn rpc.RpcConn, msg rpc.LoginCnsInfo) error {
	logger.Debug("recv LoginCnsInfo")
	return nil
}
