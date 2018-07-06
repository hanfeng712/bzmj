package xzmj

import (
	"common"
	"errorValue"
	"logger"
	"publicInterface"
	"rpc"
	"sync"
	"time"
	"timer"
)

type Room struct {
	mRoomId    uint32
	mRoomState rpc.RoomState
	mTable     *Table
	mPlayers   map[uint64]*Table
	mRoomCfg   *common.RoomConfig
	l          sync.RWMutex
	cnServer   publicInterface.CnInterface
	mRoomTimer *timer.Timer

	mRecvTableMsg chan interface{}
	mExit         chan bool
}

var XzmjRoom *Room

func NewRoom(roomId uint32) *Room {
	lXzmjRoom := &Room{}
	XzmjRoom = lXzmjRoom
	return XzmjRoom
}

func (self *Room) Init(roomId uint32, cn publicInterface.CnInterface) uint32 {
	logger.Info("Room:Init<ENTER>, roomId:%d", roomId)
	self.mRoomId = roomId
	self.mRoomState = rpc.RoomState_Room_State_Init
	self.mPlayers = make(map[uint64]*Table)
	self.mRoomCfg = common.NewRoomCfg()
	self.mRoomCfg.SRoomId = roomId
	self.mRoomCfg.SMaxPlayerNum = 10000
	self.mTable = NewTable(XzmjRoom)
	self.mTable.Init(roomId, self.mRoomCfg)
	self.cnServer = cn

	self.mExit = make(chan bool, 1)
	self.mRecvTableMsg = make(chan interface{}, 1000)
	logger.Info("Room:Init<AFTER>, roomId:%d", roomId)
	self.Tick()
	go self.RecvTableMsg()
	logger.Info("Room:Init<LEAVE>, roomId:%d", roomId)
	return uint32(errorValue.ERET_OK)
}

func (self *Room) UpdateRoomState() uint32 {
	//logger.Info("Room:UpdateRoomState<ENTER>")
	//logger.Info("Room:UpdateRoomState<LEAVE>")
	return uint32(errorValue.ERET_OK)
}

func (self *Room) Tick() {
	//logger.Info("Room:Tick<ENTER>")
	self.UpdateRoomState()
	self.CustomRoomTick()
	//logger.Info("Room:Tick<LEAVE>")
	return
}

func (self *Room) CustomRoomTick() {
	//logger.Info("Room:CustomRoomTick<ENTER>")
	//设置房间定时器
	self.mRoomTimer = timer.NewTimer(time.Duration(common.ROOM_TICK_TIME))
	self.mRoomTimer.Start(
		func() {
			self.Tick()
		},
	)
	//logger.Info("Room:CustomRoomTick<LEAVE>")
	return
}

func (self *Room) SendRoomToTableMsg(msg interface{}) uint32 {
	logger.Info("Room:SendRoomToTableMsg:<ENTER>roomId:%d", self.mRoomId)
	self.mTable.SendMsgToTable(msg)
	return errorValue.ERET_OK
}

func (self *Room) SendMsgToRoom(msg interface{}) {
	self.mRecvTableMsg <- msg
	logger.Info("Table:SendMsgToTable:len:%d", len(self.mRecvTableMsg))
}

func (self *Room) SendMsgToPlayers(msg interface{}, uid uint64, method string) {
	logger.Info("Room:SendMsgToPlayers:uid:%d", uid)
	self.cnServer.SendMsgToPlayer(msg, uid, method)
}

func (self *Room) AddPlayerToRoom(uid uint64, msg interface{}) uint32 {
	self.l.RLock()
	defer self.l.RUnlock()

	if self.mRoomState == rpc.RoomState_Room_State_Dissolve {
		return uint32(errorValue.ERET_ROOM_DISSOLVE)
	}

	if self.IsHavePlayerInRoom(uid) == true {
		return uint32(errorValue.ERET_OK)
	}

	if uint32(len(self.mPlayers)) >= self.mRoomCfg.SMaxPlayerNum {
		return errorValue.ERET_ROOM_FULL
	}

	self.NotifyPlayerRoomInfo(uid)
	//加入房间
	self.mPlayers[uid] = self.mTable
	lRet := self.mTable.SendMsgToTable(msg)
	return lRet
}

func (self *Room) PlayerDissolveRoom(uid uint64) uint32 {

	return uint32(errorValue.ERET_OK)
}

func (self *Room) SetRoomExit() {
	self.mExit <- true
}

func (self *Room) IsHavePlayerInRoom(uid uint64) bool {
	self.l.RLock()
	defer self.l.RUnlock()

	if self.mPlayers[uid] != nil {
		return true
	}

	return false
}

func (self *Room) RecvTableMsg() {
	logger.Info("Room:RecvTableMsg<ENTER>")
	for {
		select {
		case r := <-self.mRecvTableMsg:
			self.HandleRecvTableMsgs(r)
		case <-self.mExit:
			return
		default:
			continue
		}
	}
	logger.Info("Room:RecvTableMsg<LEAVE>")
}

func (self *Room) HandleRecvTableMsgs(value interface{}) {
	logger.Info("room: HandleRecvTableMsgs: recv table msg")
	switch m := value.(type) {
	case *rpc.CSUserCreateRoomRsp:
		{
			var msg *rpc.CSUserCreateRoomRsp = m
			logger.Info("Room: CSUserCreateRoomRsp:uid:%lld", msg.GetUid())
			self.SendMsgToPlayers(msg, msg.GetUid(), "SRobot.HandleCreateRoomRsp")
		}
	case *rpc.CSUserEnterRoomRsp:
		{
			var msg *rpc.CSUserEnterRoomRsp = m
			//logger.Info("Room: CSUserEnterRoomRsp:uid:%lld", msg.GetUid())
			self.SendMsgToPlayers(msg /*msg.GetUid()*/, uint64(1), "SRobot.HandleEnterRoomRsp")
		}
	case *rpc.CSUserReadyGameNotify:
		{
			var msg *rpc.CSUserReadyGameNotify = m
			//logger.Info("Room: CSUserEnterRoomRsp:uid:%lld", msg.GetUid())
			self.SendMsgToPlayers(msg /*msg.GetUid()*/, uint64(1), "SRobot.HandleReadyNotifyMsgRsp")
		}
	case *rpc.CSCommonErrMsg:
		{
			var msg *rpc.CSCommonErrMsg = m
			self.SendMsgToPlayers(msg /*msg.GetUid()*/, uint64(1), "SRobot.HandleErrorRsp")
		}
	}
	return
}

func (self *Room) NotifyPlayerRoomInfo(uid uint64) {

	lRoomAdvanceParam := rpc.RoomAdvanceParam{}
	lRoomAdvanceParam.ZiMoJiaFan = &(self.mRoomCfg.SIsZiMoJiaFan)
	lRoomAdvanceParam.ZiMoMoreThanMaxFan = &(self.mRoomCfg.SIsZiMoMoreThanMaxFan)
	lRoomAdvanceParam.JinGouDiao = &(self.mRoomCfg.SIsJinGouDiao)
	lRoomAdvanceParam.HaiDiLaoYue = &(self.mRoomCfg.SIsHaiDiLaoYue)
	lRoomAdvanceParam.DaXiaoYu = &(self.mRoomCfg.SIsDaXiaoYu)
	lRoomAdvanceParam.DianGangHuaZiMo = &(self.mRoomCfg.SIsDianGangHuaZiMo)
	lRoomAdvanceParam.YaoJiu = &(self.mRoomCfg.SIsYaoJiu)
	lRoomAdvanceParam.Jiang = &(self.mRoomCfg.SIsJiang)
	lRoomAdvanceParam.MenQing = &(self.mRoomCfg.SIsMenQing)
	lRoomAdvanceParam.HuanSanZhangType = &(self.mRoomCfg.SHuanSanZhangType)
	lRoomAdvanceParam.ZhongZhang = &(self.mRoomCfg.SIsZhongZhang)

	lUserRoomInfoChangeNotify := rpc.CSUserRoomInfoChangeNotify{}
	lUserRoomInfoChangeNotify.RoomType = &(self.mRoomCfg.SRoomType)
	lUserRoomInfoChangeNotify.RoomId = &(self.mRoomCfg.SRoomId)
	lUserRoomInfoChangeNotify.RoomState = &(self.mRoomState)
	lUserRoomInfoChangeNotify.RoomMatchNum = &(self.mRoomCfg.SMatchNum)
	lUserRoomInfoChangeNotify.RoomDiZhu = &(self.mRoomCfg.SDiZhu)
	lUserRoomInfoChangeNotify.RoomMaxBeiShu = &(self.mRoomCfg.SMaxBeiShu)
	lUserRoomInfoChangeNotify.RoomOwnerUid = &(uid)
	lPlayerNum := uint64(len(self.mPlayers))
	lUserRoomInfoChangeNotify.RoomPlayerNum = &(lPlayerNum)
	lUserRoomInfoChangeNotify.Deposit = &(self.mRoomCfg.SDeposit)
	lUserRoomInfoChangeNotify.ClubId = &(self.mRoomCfg.SRoomClubId)
	lUserRoomInfoChangeNotify.MinMatchUserNum = &(self.mRoomCfg.SMinMatchUserNum)
	lMinCurrencyValue := uint32(self.mRoomCfg.SMinCurrencyValue)
	lUserRoomInfoChangeNotify.MinCurrencyValue = &(lMinCurrencyValue) /////////////////wenti
	lUserRoomInfoChangeNotify.AdvanceParam = &(lRoomAdvanceParam)

	self.SendMsgToPlayers(&lUserRoomInfoChangeNotify, uid, "SRobot.HandleNotifyPlayerRoomInfo")
	return
}
