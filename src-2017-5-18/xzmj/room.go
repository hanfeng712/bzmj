package xzmj

import (
	"common"
	"container/list"
	"error"
	"sync"
)

type Room struct {
	lRoomId      uint64
	lRoomState   uint32
	lTable       *Table
	playerInRoom map[uint64]*player
	lRoomCfg     *common.RoomConfig
	l            sync.RWMutex
}

var XzmjRoom *Room

func NewRoom(roomId uint64) *Room {
	lXzmjRoom := &Room{}
	XzmjRoom = lXzmjRoom
	return XzmjRoom
}

func (self *Room) Init(roomId uint64) uint32 {
	self.lRoomId = roomId
	self.lRoomState = Room_State_Init
	self.playerInRoom = list.New()
	self.lTable = NewTable()
	self.lTable.Init(roomId, self.lRoomCfg)
	self.lRoomCfg.RoomId = roomId

	return uint32(error.ERET_OK)
}

func (self *Room) UpdateRoomState() uint32 {
	return uint32(error.ERET_OK)
}

func (self *Room) AddPlayerToRoom(p *player) uint32 {
	if self.lRoomState == Room_State_Dissolve {
		return uint32(error.ERET_ROOM_DISSOLVE)
	}

	if self.IsHavePlayerInRoom(p.GetUid()) == true {
		return uint32(error.ERET_OK)
	}
	self.l.RLock()
	self.playerInRoom[p.GetUid()] = p
	self.l.RUnlock()

	lTable
	return uint32(error.ERET_OK)
}

func (self *Room) PlayerDissolveRoom(uid uint64) uint32 {

	return uint32(error.ERET_OK)
}

func (self *Room) IsHavePlayerInRoom(uid uint64) bool {
	self.l.RLock()
	defer self.l.RUnlock()

	if self.playerInRoom[uid] != nil {
		return true
	}

	return false
}
