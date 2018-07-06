package xzmj

import (
	"common"
	"errorValue"
	"logger"
	"rpc"
	"sync"
)

type Table struct {
	mTableState uint32
	mUsers      [4]*MjLogicUser
	l           sync.RWMutex
	mRoomCfg    *common.RoomConfig

	mRoom        *Room
	mExit        chan bool
	mRecvRoomMsg chan interface{}
}

var XzmjTable *Table

func NewTable(room *Room) *Table {
	lXzmjTable := &Table{
		mRoom: room}
	XzmjTable = lXzmjTable
	return XzmjTable
}

func (self *Table) Init(RoomId uint32, roomCfg *common.RoomConfig) {
	logger.Info("Table:Init<ENTER>,RoomId:%d", RoomId)
	self.mTableState = Table_State_Init
	self.mRoomCfg = roomCfg
	self.mRecvRoomMsg = make(chan interface{}, 1000)
	self.mExit = make(chan bool, 1)
	for i := 0; i < 4; i++ {
		self.mUsers[i] = NewUser()
		self.mUsers[i].Init(uint32(i), RoomId, uint64(0))
	}
	go self.Run()
	logger.Info("Table:Init<LEAVE>,RoomId:%d", RoomId)
	return
}

//Table协程开始执行
func (self *Table) Run() {
	logger.Info("Table:Run<ENTER>,len:%d", len(self.mRecvRoomMsg))
	for {
		select {
		case r := <-self.mRecvRoomMsg:
			self.HandleRecvRoomMsgs(r)
		case <-self.mExit:
			return
		default:
			continue
		}
	}
}

func (self *Table) SendMsgToTable(msg interface{}) uint32 {
	self.mRecvRoomMsg <- msg
	logger.Info("Table:SendMsgToTable:len:%d", len(self.mRecvRoomMsg))
	return errorValue.ERET_OK
}

func (self *Table) HandleRecvRoomMsgs(value interface{}) {
	logger.Info("table recv msg")
	switch m := value.(type) {
	case *rpc.Login:
		{
			var msg *rpc.Login = m
			logger.Debug("table: HandleRoomMsgs: uid:%lld", msg.GetUid())
		}
	case *rpc.CSUserCreateRoomRqst:
		{
			var msg *rpc.CSUserCreateRoomRqst = m
			logger.Info("table: HandleCreateRoomMsgs:uid:%lld", msg.GetUid())
			self.AddPlayerToTable(msg.GetUid())
		}
	case *rpc.CSUserEnterRoomRqst:
		{
			var msg *rpc.CSUserEnterRoomRqst = m
			logger.Debug("table: HandleRoomMsgs:uid:%lld", msg.GetUid())
			self.AddPlayerToTable(msg.GetUid())
		}
	}
	return
}

func (self *Table) AddPlayerToTable(uid uint64) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)

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
			return uint32(errorValue.ERET_TABLE_FULL)
		}
		//填充空闲位置玩家信息
		self.mUsers[lFreeSeatId].Init(uint32(lFreeSeatId), self.mRoomCfg.SRoomId, uid)
		self.mUsers[lFreeSeatId].setUserState(Mj_User_State_Observer_Sit)
		lTableUserNum += 1
		self.NotifyAllUserTableChangInfo()

		if self.mTableState == Table_State_Init {
			//玩家人数达到最小开赛人数，则进入比赛状态，玩家可以开始准备
			if lTableUserNum >= self.mRoomCfg.SMinMatchUserNum {
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
		lRet = uint32(errorValue.ERET_TABLE_DISSOLVE)
	}
	return lRet
}

func (self *Table) StartGame() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	return lRet
}

func (self *Table) ResUserCreateRoomRqst(msg *rpc.CSUserCreateRoomRqst) {
	lUserCreateRoomRsp := rpc.CSUserCreateRoomRsp{}
	lUserCreateRoomRsp.RoomId = &(self.mRoomCfg.SRoomId)
	lUserCreateRoomRsp.RoomType = &(self.mRoomCfg.SRoomType)

	self.mRoom.SendRoomToTableMsg(lUserCreateRoomRsp)
}

func (self *Table) NotifyAllUserTableChangInfo() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)

	return lRet
}

func (self *Table) NotifyUserTableCardInfo(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	return lRet
}

func (self *Table) NotifyUserCurOperate(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	return lRet
}

func (self *Table) NotifyUserDissolveApplyInfo(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	return lRet
}

func (self *Table) NotifyUserReadyShutDownInfo(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	return lRet
}
