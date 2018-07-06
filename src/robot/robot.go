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

func SendCreateRoomMsg(conn rpc.RpcConn, uid uint64) {
	logger.Info("SendCreateRoomMsg")
	var lRoomMatchNum uint32 = 4
	var lRoomDiZhu uint32 = 0
	var lRoomMaxBeiShu uint32 = 0
	var lMAppointRoomId uint32 = 0
	var lDeposit uint32 = 0
	var lMinMatchUserNum uint32 = 4
	var lRoomPassword string
	var lMinCurrencyValue uint32 = 0
	var lClubId uint32 = 0
	var lIsPrivateRoom bool = false
	var lRewardCoin uint64 = 0
	var lJoinMatchFee uint64 = 0
	var lMatchType uint32 = 0
	var lRoomType rpc.ROOM_TYPE = rpc.ROOM_TYPE_ROOM_TYPE_CUSTOM_XUE_ZHAN_MJ

	var lZiMoJiaFan bool = true
	var lZiMoMoreThanMaxFan bool = true
	var lJinGouDiao bool = true
	var lHaiDiLaoYue bool = true
	var lDaXiaoYu bool = true
	var lDianGangHuaZiMo bool = true
	var lYaoJiu bool = true
	var lJiang bool = true
	var lMenQing bool = true
	var lHuanSanZhangType uint32 = 3
	var lZhongZhang bool = true

	lRoomAdvanceParam := rpc.RoomAdvanceParam{}
	lRoomAdvanceParam.ZiMoJiaFan = &(lZiMoJiaFan)
	lRoomAdvanceParam.ZiMoMoreThanMaxFan = &(lZiMoMoreThanMaxFan)
	lRoomAdvanceParam.JinGouDiao = &(lJinGouDiao)
	lRoomAdvanceParam.HaiDiLaoYue = &(lHaiDiLaoYue)
	lRoomAdvanceParam.DaXiaoYu = &(lDaXiaoYu)
	lRoomAdvanceParam.DianGangHuaZiMo = &(lDianGangHuaZiMo)
	lRoomAdvanceParam.YaoJiu = &(lYaoJiu)
	lRoomAdvanceParam.Jiang = &(lJiang)
	lRoomAdvanceParam.MenQing = &(lMenQing)
	lRoomAdvanceParam.HuanSanZhangType = &(lHuanSanZhangType)
	lRoomAdvanceParam.ZhongZhang = &(lZhongZhang)
	createRoomReq := rpc.CSUserCreateRoomRqst{}
	createRoomReq.RoomType = &(lRoomType)
	createRoomReq.RoomMatchNum = &(lRoomMatchNum)
	createRoomReq.RoomDiZhu = &(lRoomDiZhu)
	createRoomReq.RoomMaxBeiShu = &(lRoomMaxBeiShu)
	createRoomReq.MAppointRoomId = &(lMAppointRoomId)
	createRoomReq.Deposit = &(lDeposit)
	createRoomReq.MinMatchUserNum = &(lMinMatchUserNum)
	createRoomReq.RoomPassword = &(lRoomPassword)
	createRoomReq.MinCurrencyValue = &(lMinCurrencyValue)
	createRoomReq.ClubId = &(lClubId)
	createRoomReq.IsPrivateRoom = &(lIsPrivateRoom)
	createRoomReq.RewardCoin = &(lRewardCoin)
	createRoomReq.JoinMatchFee = &(lJoinMatchFee)
	createRoomReq.MatchType = &(lMatchType)
	createRoomReq.Uid = &(uid)
	createRoomReq.AdvanceParam = &(lRoomAdvanceParam)

	SendCreateRoomMsgToNet(conn, &createRoomReq)

}

func SendCreateRoomMsgToNet(conn rpc.RpcConn, value interface{}) {
	logger.Info("SendCreateRoomMsg")
	common.WriteClientResult(conn, "CNServer.CreateRoom", value)
}

func SendMsg(conn rpc.RpcConn, value interface{}) {
	logger.Info("SendMsg")
	common.WriteClientResult(conn, "CNServer.Login", value)
}
