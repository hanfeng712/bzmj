package xzmj

import (
	"common"
	"error"
	"sync"
)

type Table struct {
	mTableState uint32
	mUsers      [4]*MjLogicUser
	l           sync.RWMutex
	mRoomCfg    *common.RoomConfig
	
	mExit         chan bool
	mRecvMsg		chan {}interface
}

var XzmjTable *Table

func NewTable() *Table {
	lXzmjTable := &Table{}
	XzmjTable = lXzmjTable
	return XzmjTable
}

func (self *Table) Init(RoomId uint64, roomCfg *common.RoomConfig) {
	self.mTableState = Table_State_Init
	self.mRoomCfg = roomCfg
	self.mRecvMsg = make(chan {}interface,1000)
	self.mExit = make(chan bool, 1)
	var i uint32
	for i = 0; i < 4; i++ {
		self.mUsers[i].Init(i, RoomId, uint64(0))
	}
}

func (self *Table) AddPlayerToTable(p *player) uint32 {
	var lRet uint32 = uint32(error.ERET_OK)

	if self.mTableState != Table_State_Dissolve {
		var lFreeSeatId int = -1
		var lTableUserNum uint32 = 0
		for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
			if self.mUsers[i].IsFree() == true {
				if lFreeSeatId == -1 {
					lFreeSeatId = i
				}
			} else {
				lTableUserNum++
			}
		}
		//桌子已满座
		if lTableUserNum == MAX_PLAYER_NUM_PER_TABLE {
			return uint32(error.ERET_TABLE_FULL)
		}
		//填充空闲位置玩家信息
		self.mUsers[lFreeSeatId].Init(lFreeSeatId, self.mRoomCfg.RoomId)
		self.mUsers[lFreeSeatId].setUserState(Mj_User_State_Observer_Sit)
		lTableUserNum += 1

		if self.mTableState == Table_State_Init {
			//玩家人数达到最小开赛人数，则进入比赛状态，玩家可以开始准备
			if lTableUserNum >= self.mRoomCfg.MinMatchUserNum {
				self.StartGame()
			}
		} else {
			//如果比赛已开始，则是中途进入比赛，通知其比赛数据
			self.NotifyUserTableCardInfo(lFreeSeatId)
			self.NotifyUserCurOperate(lFreeSeatId)
			if self.mTableState == Table_State_Wait_Dissolve {
				self.NotifyUserDissolveApplyInfo(lFreeSeatId)
			}
			self.NotifyUserReadyShutDownInfo(lFreeSeatId)
		}
	} else {
		lRet = uint32(error.ERET_TABLE_DISSOLVE)
	}
	return lRet
}

func (self *Table) StartGame() uint32 {
	var lRet uint32 = uint32(error.ERET_OK)
	return lRet
}

func (self *Table) GameBegin() uint32 {
	SetTableState(Table_State_Gaming)
	go self.Run()
}

func (self *Table) Run(){
	for{
		select{
			case r := <-self.mRecvMsg:
				
			case <-self.mExit:
				return
		}
	}
}
func (self *Table) SetTableState(state uint32) uint32{
	self.mTableState = state
}

func (self *Table) NotifyUserTableCardInfo(seatId int) uint32 {
	var lRet uint32 = uint32(error.ERET_OK)
	return lRet
}

func (self *Table) NotifyUserCurOperate(seatId int) uint32 {
	var lRet uint32 = uint32(error.ERET_OK)
	return lRet
}

func (self *Table) NotifyUserDissolveApplyInfo(seatId int) uint32 {
	var lRet uint32 = uint32(error.ERET_OK)
	return lRet
}

func (self *Table) NotifyUserReadyShutDownInfo(seatId int) uint32 {
	var lRet uint32 = uint32(error.ERET_OK)
	return lRet
}
