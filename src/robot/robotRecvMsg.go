package robot

import (
	"logger"
	"rpc"
)

func (self *SRobot) HandlePongRsp(conn rpc.RpcConn, msg rpc.Pong) error {
	logger.Info("recv HandlePongRsp : uid:%d, count:%d", self.uid, msg.Count)
	self.stateMatchine.SwitchFsmState()
	return nil
}
