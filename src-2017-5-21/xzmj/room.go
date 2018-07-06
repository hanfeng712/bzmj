package xzmj

import (
	"common"
	"errorValue"
	"logger"
	"sync"
	"time"
	"timer"
)

type Room struct {
	mRoomId    uint32
	mRoomState uint32
	mTable     *Table
	mPlayers   map[uint64]*Table
	mRoomCfg   *common.RoomConfig
	l          sync.RWMutex

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

func (self *Room) Init(roomId uint32) uint32 {
	logger.Info("Room:Init<ENTER>, roomId:%d", roomId)
	self.mRoomId = roomId
	self.mRoomState = Room_State_Init
	self.mPlayers = make(map[uint64]*Table)
	self.mTable = NewTable(XzmjRoom)
	self.mTable.Init(roomId, self.mRoomCfg)
	self.mRoomCfg = common.NewRoomCfg()
	self.mRoomCfg.SRoomId = roomId

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

func (self *Room) AddPlayerToRoom(uid uint64, msg interface{}) uint32 {
	self.l.RLock()
	defer self.l.RUnlock()

	if self.mRoomState == Room_State_Dissolve {
		return uint32(errorValue.ERET_ROOM_DISSOLVE)
	}

	if self.IsHavePlayerInRoom(uid) == true {
		return uint32(errorValue.ERET_OK)
	}

	if uint32(len(self.mPlayers)) >= self.mRoomCfg.SMaxPlayerNum {
		return errorValue.ERET_ROOM_FULL
	}

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
		}
	}
	logger.Info("Room:RecvTableMsg<LEAVE>")
}

func (self *Room) HandleRecvTableMsgs(value interface{}) {
	logger.Info("room: HandleRecvTableMsgs: recv table msg")
	return
}
