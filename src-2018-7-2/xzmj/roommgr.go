package xzmj

import (
	"errorValue"
	"logger"
	"math/rand"
	"publicInterface"
	"sync"
	"time"
)

type RoomMgr struct {
	rooms    map[uint32]*Room
	cnServer publicInterface.CnInterface
	l        sync.RWMutex
}

var XzmjRoomMgr *RoomMgr

func NewRoomMgr(cn publicInterface.CnInterface) *RoomMgr {
	lXzmjRoomMgr := &RoomMgr{
		rooms:    make(map[uint32]*Room),
		cnServer: cn}

	XzmjRoomMgr = lXzmjRoomMgr
	return lXzmjRoomMgr
}

func (self *RoomMgr) CreateRoom(roomId *uint32) uint32 {
	logger.Debug("RoomMgr:CreateRoom")
	self.l.RLock()
	defer self.l.RUnlock()

	var lRoomId uint32 = 0
	lRet := self.GenRoomId(&lRoomId)
	if lRet != uint32(errorValue.ERET_OK) {
		return lRet
	}
	*roomId = lRoomId
	return lRet
}

func (self *RoomMgr) EnterRoom(roomId uint32, uid uint64) {

}

func (self *RoomMgr) GenRoomId(roomId *uint32) uint32 {
	logger.Info("Room:GenRoomId<ENTER>")
	self.l.RLock()
	defer self.l.RUnlock()

	lCount := 0
	var lRoomId uint32 = 0
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		lRoomIdTmp := r.Int63n(600000) + 600000
		if lCount > 10 {
			break
		}
		if self.rooms[uint32(lRoomIdTmp)] != nil {
			lCount++
			continue
		}
		lRoomId = uint32(lRoomIdTmp)
		break
	}
	if lRoomId <= 0 {
		logger.Info("Room:GenRoomId<LEAVE1>")
		return uint32(errorValue.ERET_GENERATE_CUSTOM_ROOM_ID)
	}
	//创建房间
	lRoom := NewRoom(lRoomId)
	lRet := lRoom.Init(lRoomId, self.cnServer)
	self.rooms[lRoomId] = lRoom
	*roomId = lRoomId
	logger.Info("Room:GenRoomId<LEAVE2>")
	return lRet
}

func (self *RoomMgr) HandleRoomToTableMsg(roomId uint32, msg interface{}) uint32 {
	logger.Info("RoomMgr:HandleRoomToTableMsg:<ENTER>roomId:%d", roomId)
	self.l.RLock()
	lRoom := self.rooms[roomId]
	if lRoom == nil {
		logger.Info("RoomMgr:HandleRoomToTableMsg:<LEAVE>roomId:%d", roomId)
		return errorValue.ERET_SYS_ERR
	}
	lRet := lRoom.SendRoomToTableMsg(msg)
	return lRet
}

func (self *RoomMgr) DestroyCustomRoom(roomId uint32) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	self.l.RLock()
	lRoom, lOk := self.rooms[roomId]
	if lOk == false {
		return uint32(errorValue.ERET_INVALID_ROOM_ID)
	}

	lRoom.SetRoomExit()
	delete(self.rooms, roomId)
	return lRet

}

func (self *RoomMgr) GetRoomById(roomId uint32) *Room {
	self.l.RLock()
	defer self.l.RUnlock()

	lRoom := self.rooms[roomId]
	if lRoom != nil {
		return lRoom
	}

	return nil
}
