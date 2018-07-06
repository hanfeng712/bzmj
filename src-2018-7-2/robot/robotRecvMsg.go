package robot

import (
	"logger"
	"rpc"
)

func (self *SRobot) HandleLoginRsp(conn rpc.RpcConn, login rpc.Login) error {
	logger.Info("recv HandleLoginRsp uid:%d", login.GetUid())
	SendCreateRoomMsg(conn, login.GetUid())
	return nil
}

func (self *SRobot) HandleCreateRoomRsp(conn rpc.RpcConn, msg rpc.CSUserCreateRoomRsp) error {
	logger.Info("recv HandleCreateRoomRsp uid:%d", msg.GetUid())
	return nil
}

func (self *SRobot) HandleEnterRoomRsp(conn rpc.RpcConn, msg rpc.CSUserEnterRoomRsp) error {
	logger.Info("recv HandleEnterRoomRsp uid:%d", 1)
	return nil
}

func (self *SRobot) HandleReadyNotifyMsgRsp(conn rpc.RpcConn, msg rpc.CSUser) error {
	logger.Info("recv HandleReadyNotifyMsgRsp uid:%d", 1)
	return nil
}

func (self *SRobot) HandleErrorRsp(conn rpc.RpcConn, msg rpc.CSCommonErrMsg) error {
	logger.Info("recv HandleErrorRsp uid:%d", msg.ErrorCode)
	return nil
}
