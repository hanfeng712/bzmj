package xzmj

import (
	"error"
	"math/rand"
	"sync"
	"time"
)

type RoomMgr struct {
	rooms map[uint64]*Room
	l     sync.RWMutex
}

var XzmjRoomMgr *RoomMgr

func NewRoomMgr() *RoomMgr {
	lXzmjRoomMgr := &RoomMgr{
		rooms: make(map[uint64]*Room),
	}

	XzmjRoomMgr = lXzmjRoomMgr
	return XzmjRoomMgr
}
func (self *RoomMgr) CreateRoom() uint32 {
	var lRoomId uint64 = 0
	lRet := self.GenRoomId(&lRoomId)
	if lRet != uint32(error.ERET_OK) {
		return lRet
	}

	return lRet
}
func (self *RoomMgr) EnterRoom(roomId uint32, uid uint64) {

}

func (self *RoomMgr) GenRoomId(roomId *uint64) uint32 {
	self.l.RLock()
	defer self.l.RUnlock()

	lCount := 0
	var lRoomId uint64 = 0
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		lRoomIdTmp := r.Int63n(600000) + 600000
		if lCount > 10 {
			break
		}
		if self.rooms[uint64(lRoomIdTmp)] != nil {
			lCount++
			continue
		}
		lRoomId = uint64(lRoomIdTmp)
		break
	}
	if lRoomId <= 0 {
		return uint32(error.ERET_GENERATE_CUSTOM_ROOM_ID)
	}
	//创建房间
	lRoom := NewRoom(lRoomId)
	lRet := lRoom.Init(lRoomId)
	self.rooms[lRoomId] = lRoom
	*roomId = lRoomId
	return lRet
}
