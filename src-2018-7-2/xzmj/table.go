package xzmj

import (
	"common"
	"container/heap"
	"container/list"
	"errorValue"
	"logger"
	"math/rand"
	"rpc"
	"sync"
	"time"
	"timer"
)

type Table struct {
	mTableState           rpc.TableState
	mPrevTableState       rpc.TableState
	mUsers                [4]*MjLogicUser
	mMjCardMgr            *MjCards
	mDice1                uint32
	mDice2                uint32
	mZhuang               uint32
	mNextZhuangSeatId     uint32
	mLastMoPaiUserSeatId  uint32
	mCurRelationSeatId    uint32
	mFinishedMatchNum     uint32
	mLastDealUserSeatId   uint32
	mLastGangUserSeatId   uint32
	mWaitBuGangTagMj      uint32
	mWaitBuGangUserSeatId uint32
	mDissolveApplySeatId  uint32
	mFinalDecision        rpc.TableFinalDecision
	mEscapeSettlement     bool
	mHuanSanZhangType     rpc.TASK_HSZ_TYPE

	mLastGameInfo *MjLastTableGameInfo

	l                        sync.RWMutex
	mRoomCfg                 *common.RoomConfig
	mWaitOperateTimerId      *timer.Timer
	mWaitOperateTimerOutTime int64
	mWaitRenewRoomTimerId    *timer.Timer

	mSettlementId uint64

	mDisagreeDissolveApplySeatIds IntHeap
	mAgreeDissolveApplySeatIds    IntHeap
	mWaitReadyTimer               *timer.Timer
	mReadyTimer                   *timer.Timer
	mRoom                         *Room
	mExit                         chan bool
	mReadyShutDown                bool
	mRecvRoomMsg                  chan interface{}
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
	self.mTableState = rpc.TableState_Table_State_Init
	self.mPrevTableState = rpc.TableState_Table_State_Init
	self.mRoomCfg = roomCfg
	self.mRecvRoomMsg = make(chan interface{}, 1000)
	self.mReadyTimer = nil
	self.mSettlementId = 0
	self.mMjCardMgr = NewMjCard()
	self.mDice1 = uint32(0)
	self.mDice2 = uint32(0)
	self.mZhuang = uint32(0)
	self.mCurRelationSeatId = uint32(0)
	self.mFinishedMatchNum = uint32(0)
	self.mLastDealUserSeatId = uint32(TABLE_SEAT_NONE)
	self.mLastGangUserSeatId = uint32(TABLE_SEAT_NONE)
	self.mNextZhuangSeatId = uint32(TABLE_SEAT_NONE)
	self.mWaitBuGangUserSeatId = uint32(TABLE_SEAT_NONE)
	self.mWaitBuGangTagMj = uint32(TABLE_SEAT_NONE)
	self.mDissolveApplySeatId = uint32(TABLE_SEAT_NONE)
	self.mHuanSanZhangType = rpc.TASK_HSZ_TYPE_TASK_HSZ_TYPE_NONE
	self.mFinalDecision = rpc.TableFinalDecision_Table_Decision_None
	self.mEscapeSettlement = false
	self.mExit = make(chan bool, 1)
	self.mWaitOperateTimerOutTime = time.Now().UnixNano()
	self.mDisagreeDissolveApplySeatIds.Clear()
	self.mAgreeDissolveApplySeatIds.Clear()
	self.mLastGameInfo = NewMjLastTableGameInfo()
	self.mReadyShutDown = false
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

func (self *Table) AddPlayerToTable(uid uint64) uint32 {
	logger.Info("Table:AddPlayerToTable<ENTER>,uid:%d", uid)
	var lRet uint32 = uint32(errorValue.ERET_OK)

	if self.mTableState != rpc.TableState_Table_State_Dissolve {
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
		logger.Info("Table:AddPlayerToTable<AFTER1>,uid:%d,lFreeSeatId:%d,RoomId:%d", uid, lFreeSeatId, self.mRoomCfg.SRoomId)
		//填充空闲位置玩家信息
		self.mUsers[lFreeSeatId].Init(uint32(lFreeSeatId), self.mRoomCfg.SRoomId, uid)
		self.mUsers[lFreeSeatId].SetUserState(rpc.MjUserState_Mj_User_State_Observer_Sit)
		lTableUserNum += 1
		self.NotifyAllUserTableChangInfo()

		if self.mTableState == rpc.TableState_Table_State_Init {
			//玩家人数达到最小开赛人数，则进入比赛状态，玩家可以开始准备
			if lTableUserNum >= self.mRoomCfg.SMinMatchUserNum {
				self.StartGame()
			}
		} else {
			//如果比赛已开始，则是中途进入比赛，通知其比赛数据
			self.NotifyUserTableCardInfo(lFreeSeatId)
			self.NotifyUserCurOperate(lFreeSeatId)
			if self.mTableState == rpc.TableState_Table_State_Wait_Dissolve {
				self.NotifyUserDissolveApplyInfo(lFreeSeatId)
			}
			self.NotifyUserReadyShutDownInfo(lFreeSeatId)
		}
	} else {
		lRet = uint32(errorValue.ERET_TABLE_DISSOLVE)
	}

	logger.Info("Table:AddPlayerToTable<LEAVE>,uid:%d", uid)
	return lRet
}

func (self *Table) HandleReadyGameMsg(msg *rpc.CSUserReadyGameRqst) uint32 {
	var lRet uint32 = errorValue.ERET_OK

	lSeatId := int(msg.GetReadySeatId())
	//lUid := msg.GetUid()

	if lSeatId > MAX_PLAYER_NUM_PER_TABLE {
		lRet = errorValue.ERET_INVALID_SEAT_ID
		return lRet
	}

	if self.mTableState != rpc.TableState_Table_State_Start {
		lRet = errorValue.ERET_TABLE_STATE
		return lRet
	}

	lRet = self.HandleMjReady(lSeatId)
	return lRet
}

func (self *Table) HandleMjReady(seatId int) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	self.mUsers[seatId].SetUserState(rpc.MjUserState_Mj_User_State_Ready)

	lRet = self.NotifyAllUserReady(seatId)
	if lRet != errorValue.ERET_OK {
		return lRet
	}

	lRet = self.NotifyAllUserTableChangInfo()
	if lRet != errorValue.ERET_OK {
		return lRet
	}

	if self.IsAllUserReadyOver() == true {
		lRet = self.GameBegin()
	}

	return lRet
}

func (self *Table) IsAllUserReadyOver() bool {
	lbRet := false

	if self.GetUserNum() < 2 {
		return lbRet
	}

	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsFree() == true {
			continue
		}
		if self.mUsers[i].GetUserState() != rpc.MjUserState_Mj_User_State_Ready {
			return lbRet
		}
	}
	lbRet = true
	return lbRet
}

func (self *Table) StartGame() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	self.SetTableState(rpc.TableState_Table_State_Start)
	self.NotifyAllUserTableChangInfo()
	if self.mReadyTimer != nil {
		self.mReadyTimer.Stop()
	}
	self.mWaitReadyTimer = timer.NewTimer(time.Duration(common.WaitReadyTimeout))
	self.mWaitReadyTimer.Start(
		func() {
			self.WaitReadyTimeOut()
		},
	)
	//检查玩家金币，金币不足时进入等待充值状态
	self.CheckUserCurrencyForMatch()
	return lRet
}

func (self *Table) GameBegin() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	self.SetTableState(rpc.TableState_Table_State_Gaming)

	self.DanJuRoomCostSettlement()

	if self.mWaitReadyTimer != nil {
		self.mWaitReadyTimer.Stop()
	}

	if self.mSettlementId == 0 {
		self.mSettlementId = self.GenerateSettlementId()
	}

	self.mMjCardMgr.Init()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//lDice1 := r.Int63n(6) + 1
	self.mDice1 = r.Uint32()%6 + uint32(1)
	self.mDice2 = r.Uint32()%6 + uint32(1)

	//获取有效庄
	if self.mZhuang < 0 || self.mZhuang > 3 {
		if self.mNextZhuangSeatId == TABLE_SEAT_NONE {
			self.mNextZhuangSeatId = uint32(0)
		}
		self.mZhuang = self.mNextZhuangSeatId
		if self.mUsers[int(self.mZhuang)].IsMatch() == false {
			for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
				if self.mUsers[int(self.mZhuang)].IsMatch() == true {
					self.mZhuang = uint32(i)
					break
				}
			}
		}
	} else {
		for i := 0; i < MAX_PLAYER_NUM_PER_TABLE && (self.mUsers[int(self.mZhuang)].IsMatch() == false); i++ {
			self.mZhuang = self.mZhuang % MAX_PLAYER_NUM_PER_TABLE
		}
	}

	//fa pai
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsMatch() == false {
			continue
		}
		var lCardCount uint32 = 13
		if i == int(self.mZhuang) {
			lCardCount = uint32(14)
		}
		self.mMjCardMgr.FaPai(&(self.mUsers[i].mInitHandCards), lCardCount)
		self.mUsers[i].mHandCards = self.mUsers[i].mInitHandCards
		self.mUsers[i].SetUserState(rpc.MjUserState_Mj_User_State_Playing)
	}
	self.mLastMoPaiUserSeatId = self.mZhuang
	self.NotifyAllUserTableChangInfo()
	self.NotifyAllUserFaPai()

	//	lOperateTimeOut := common.WaitRejectSuitTimeout
	if self.mRoomCfg.SHuanSanZhangType == rpc.HUAN_SAN_ZHANG_TYPE_HUAN_SAN_ZHANG_TYPE_NONE {
		for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
			if self.mUsers[i].IsMatch() == false {
				continue
			}
			self.mUsers[i].SetOperateMask(rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_REJECTSUIT)
			self.NotifyAllUserWaitRejectSuit()
		}
	} else {
		//TODO:换三张
	}

	self.mCurRelationSeatId = TABLE_SEAT_NONE
	self.mWaitOperateTimerId = timer.NewTimer(time.Duration(common.WaitOperateTimeout))
	self.mWaitOperateTimerOutTime = time.Now().UnixNano() + int64(common.WaitOperateTimeout)
	self.mWaitOperateTimerId.Start(
		func() {
			self.WaitOperateTimeOut()
		},
	)
	return lRet
}

func (self *Table) UserGameEnd(seatId uint32) uint32 {
	self.NotifyUserAllSettlementInfo(seatId)
	return errorValue.ERET_OK
}

func (self *Table) GameEnd() uint32 {
	//TODO
	self.SetTableState(rpc.TableState_Table_State_Game_End)
	self.NotifyAllUserTableChangInfo()
	self.NotifyAllUserHandCards()
	self.NotifyAllUserSettlementInfo()
	self.CacheLastGameInfo()
	self.NotifyGameAllOperation(-1)
	self.mFinishedMatchNum += uint32(1)
	//TODO 记录每局的比赛输赢大小
	//TODO mongo
	//TODO 报告比赛结果给玩家
	//如果玩家能够继续开赛，则继续比赛
	if self.IsAllUserCanContinueGame() {
		self.GameRestart()
	} else {
		//如果玩家不能继续开赛，则要么解散，要么房主续桌
		//TODO
		self.DestroyTable()
	}
	return errorValue.ERET_OK
}

func (self *Table) IsAllUserCanContinueGame() bool {
	var lbRet bool = true
	if self.mReadyShutDown == true {
		return false
	}

	if self.mFinishedMatchNum >= self.mRoomCfg.SMatchNum {
		return false
	}

	if self.IsHavePlayerGiveUp() {
		lbRet = false
	}
	return lbRet
}

func (self *Table) IsHavePlayerGiveUp() bool {
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsGiveUp() == true {
			return true
		}
	}
	return false
}
func (self *Table) GameOver() uint32 {
	//TODO
	self.DestroyTable()
	return errorValue.ERET_OK
}

func (self *Table) GameRestart() uint32 {
	self.ClearTable(false)
	self.StartGame()
	return errorValue.ERET_OK
}

func (self *Table) GameWaitRenew() uint32 {
	var lRet uint32 = errorValue.ERET_OK
	lOperateTimeOut := common.RenewRoomTimeout
	self.ClearTable(true)
	self.SetTableState(rpc.TableState_Table_State_Wait_Renew)
	self.mWaitRenewRoomTimerId = timer.NewTimer(time.Duration(lOperateTimeOut))
	self.mWaitRenewRoomTimerId.Start(
		func() {
			self.WaitRenewTimeOut()
		},
	)
	self.NotifyAllUserTableChangInfo()
	return lRet
}
func (self *Table) GetUserNum() uint32 {
	var lNum uint32 = 0
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsFree() == true {
			continue
		}
		lNum++
	}
	return lNum
}

func (self *Table) GenerateSettlementId() uint64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	lSettleIdTmp := r.Int63n(900000) + 100000

	return uint64(lSettleIdTmp)
}
func (self *Table) SetTableState(state rpc.TableState) {
	self.mTableState = state
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
			logger.Info("table: HandleCreateRoomMsgs:uid:%d", msg.GetUid())
			lRet := self.AddPlayerToTable(msg.GetUid())
			if lRet != errorValue.ERET_OK {
				//TODO:Res error
				logger.Info("table: HandleCreateRoomMsgs:lRet:%d", lRet)
				self.AnswerClientError(lRet, msg.GetUid())
				return
			}
			self.ResUserCreateRoomRqst(msg.GetUid())
		}
	case *rpc.CSUserEnterRoomRqst:
		{
			var msg *rpc.CSUserEnterRoomRqst = m
			lRet := self.AddPlayerToTable(msg.GetUid())
			if lRet != errorValue.ERET_OK {
				//TODO:Res error
				logger.Info("table: HandleCreateRoomMsgs:lRet:%d", lRet)
				self.AnswerClientError(lRet, msg.GetUid())
				return
			}
			self.ResUserEnterRoomRqst(msg.GetUid(), true)
		}
	case *rpc.CSUserReadyGameRqst:
		{
			var msg *rpc.CSUserReadyGameRqst = m
			self.HandleReadyGameMsg(msg)
		}
	}
	return
}

func (self *Table) HandleMjOperateMsg(userMjOperateRqst *rpc.CSUserMjOperateRqst) uint32 {

	var lUid uint64 = userMjOperateRqst.GetUid()
	var lSeatId uint32 = userMjOperateRqst.GetOperateSeatId()
	var lRet uint32 = errorValue.ERET_OK

	if self.mTableState == rpc.TableState_Table_State_Wait_Dissolve {
		return errorValue.ERET_MJ_OPERATE_IN_WAIT_DISSOLVE
	}

	//检查uid 与 SeateId 的对应关系
	if self.CheckUidSeatId(lUid, lSeatId) == false {
		lRet = errorValue.ERET_TABLE_UID_SEATID
		return lRet
	}

	//检查tagmj 有效性
	var lOperateMask rpc.MJ_OPERATE_MASK = userMjOperateRqst.GetOperateMask()
	var lTagMj uint32 = userMjOperateRqst.GetTagMj()
	if self.CheckMjVaild(lTagMj) {
		return errorValue.ERET_INVAILD_MJ
	}
	var lTagMj1 uint32 = userMjOperateRqst.GetTagMj1()
	var lTagMj2 uint32 = userMjOperateRqst.GetTagMj2()

	if lTagMj1 != uint32(0) {
		if self.CheckMjVaild(lTagMj1) == false {
			lRet = errorValue.ERET_INVAILD_MJ
		}
	}
	if lTagMj2 != uint32(0) {
		if self.CheckMjVaild(lTagMj2) == false {
			lRet = errorValue.ERET_INVAILD_MJ
		}
	}
	if lRet != errorValue.ERET_OK {
		return lRet
	}
	switch lOperateMask { //rpc.MJ_OPERATE_MASK
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_DEAL:
		lRet = self.HandleMjOperateDeal(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_PENG:
		lRet = self.HandleMjOperatePeng(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_MING_GANG:
		lRet = self.HandleMjOperateMingGang(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_AN_GANG:
		lRet = self.HandleMjOperateAnGang(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_BU_GANG:
		lRet = self.HandleMjOperateBuGang(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_REJECTSUIT:
		lRet = self.HandleMjOperateRejectSuit(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_HUAN_SAN_ZHANG:
		lRet = self.HandleMjOperateHuanSanZhang(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_HU:
		lRet = self.HandleMjOperateHu(lSeatId, lTagMj)
	case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_GUO:
		lRet = self.HandleMjOperateGuo(lSeatId, lTagMj)
	default:
		lRet = errorValue.ERET_NO_OPERATE_MASK
	}
	return lRet
}

func (self *Table) HandleMjOperateDeal(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	var lSeatId int = int(seatId)
	if lSeatId <= 0 || lSeatId > MAX_PLAYER_NUM_PER_TABLE {
		return errorValue.ERET_SYS_ERR
	}
	if self.mUsers[lSeatId].CanDealByTag(tagMj) == false {
		return errorValue.ERET_NO_OPERATE_MASK
	}
	lRet = self.mUsers[lSeatId].DealCard(tagMj)
	if errorValue.ERET_OK != lRet {
		return lRet
	}
	self.mLastDealUserSeatId = seatId
	self.mUsers[lSeatId].clearOperateMask()
	self.NotifyAllUserDeal(seatId, tagMj)
	//TODO self.AddUserOperateRecord(seatId,MJ_OPERATE_MASK_DEAL,tagMj)
	if seatId != self.mLastDealUserSeatId {
		self.mLastDealUserSeatId = TABLE_SEAT_NONE
	}
	self.mWaitOperateTimerId.Stop()
	if true == self.IsHaveUserCanOperateForDealCard(seatId, tagMj) {
		self.SetAllUserOperateMaskForDealCard(seatId, tagMj)
		self.NotifyUserWaitOperateForDealCard(seatId)
		self.mCurRelationSeatId = seatId
		self.mWaitOperateTimerOutTime = time.Now().UnixNano() + int64(common.WaitOperateTimeout)
		self.mWaitOperateTimerId = timer.NewTimer(time.Duration(common.WaitOperateTimeout))
		self.mWaitOperateTimerId.Start(
			func() {
				self.WaitOperateTimeOut()
			},
		)
	} else {
		self.UserMoPai(self.GetNextSeatId(seatId))
	}
	return lRet
}
func (self *Table) HandleMjOperatePeng(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) HandleMjOperateMingGang(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) HandleMjOperateAnGang(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) HandleMjOperateBuGang(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) HandleMjOperateRejectSuit(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) HandleMjOperateHuanSanZhang(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) HandleMjOperateHu(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) HandleMjOperateGuo(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	return lRet
}
func (self *Table) CheckUidSeatId(uid uint64, seatId uint32) bool {
	if seatId >= MAX_PLAYER_NUM_PER_TABLE {
		return false
	}
	if self.mUsers[int(seatId)].IsFree() {
		return false
	}
	return true
}

func (self *Table) CheckMjVaild(tagMj uint32) bool {
	if uint32(11) <= tagMj && tagMj <= uint32(39) {
		return true
	}
	if tagMj == MJ_BACK {
		return true
	}
	return false
}
func (self *Table) WaitRenewTimeOut() {
	if self.mTableState != rpc.TableState_Table_State_Dissolve {
		self.DestroyTable()
	}
}

func (self *Table) IsHaveUserCanOperateForDealCard(seatId uint32, tagMj uint32) bool {
	lbRet := false
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsMatch() == false {
			continue
		}
		lbRet = self.mUsers[i].IsCanOperateForDealCard(seatId, tagMj, self.mRoomCfg)
		if lbRet == true {
			break
		}
	}
	return lbRet
}

func (self *Table) SetAllUserOperateMaskForDealCard(seatId uint32, tagMj uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsMatch() == false {
			continue
		}
		var lOperateMask *list.List
		var lMjCountInPool uint32 = self.mMjCardMgr.GetLeftCardsNum()
		lOperateMask = self.mUsers[i].CalculateOperateMaskForDealCard(seatId, tagMj, lMjCountInPool, self.mRoomCfg)
		self.mUsers[i].SetOperateMaskList(lOperateMask)
	}

	return lRet
}

func (self *Table) GetNextSeatId(curSeatId uint32) uint32 {
	var lNextSeatId uint32 = curSeatId
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		lNextSeatId = (lNextSeatId + 1) % MAX_PLAYER_NUM_PER_TABLE
		if self.mUsers[lNextSeatId].IsMatch() == false {
			continue
		}
		if self.mUsers[lNextSeatId].IsEnd() {
			continue
		}
		break
	}
	return lNextSeatId
}
func (self *Table) WaitOperateTimeOut() {
	if self.mTableState == rpc.TableState_Table_State_Dissolve {
		return
	}

	if self.mRoomCfg.SIsAllowAgentOperator {
		//允许服务器代打
		var lSeatIds IntHeap
		for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
			if self.mUsers[i].CanOperater() {
				heap.Push(&lSeatIds, uint32(i))
			}
		}
		for seatId := range lSeatIds {
			var lTagMj uint32 = MJ_BACK
			var lTagMj1 uint32 = MJ_BACK
			var lTagMj2 uint32 = MJ_BACK
			//var lSeatId uint32 = uint32(seatId)
			var lOperateMask rpc.MJ_OPERATE_MASK = rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_NONE
			self.mUsers[seatId].AiCalculateOperateInfo(lOperateMask, lTagMj, lTagMj1, lTagMj2)
			switch lOperateMask {
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_DEAL:
				//HandleMjOperateDeal(lSeatId, lTagMj);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_PENG:
				//HandleMjOperatePeng(lSeatId, lTagMj);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_MING_GANG:
				//HandleMjOperateMingGang(lSeatId, lTagMj);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_AN_GANG:
				//HandleMjOperateAnGang(lSeatId, lTagMj);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_BU_GANG:
				//HandleMjOperateBuGang(lSeatId, lTagMj);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_REJECTSUIT:
				//HandleMjOperateRejectSuit(lSeatId, lTagMj);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_HUAN_SAN_ZHANG:
				//HandleMjOperateHuanSanZhang(lSeatId, lTagMj, lTagMj1, lTagMj2);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_HU:
				//HandleMjOperateHu(lSeatId, lTagMj);
				break
			case rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_GUO:
				//HandleMjOperateGuo(lSeatId, lTagMj);
				break
			}
		}
	}
	return
}
func (self *Table) ResUserCreateRoomRqst(uid uint64) {
	logger.Info("Table:ResUserCreateRoomRqst:<ENTER>, uid:%d", uid)
	lUserCreateRoomRsp := rpc.CSUserCreateRoomRsp{}
	lUserCreateRoomRsp.RoomId = &(self.mRoomCfg.SRoomId)
	lUserCreateRoomRsp.RoomType = &(self.mRoomCfg.SRoomType)
	lUserCreateRoomRsp.Uid = &(uid)
	self.mRoom.SendMsgToRoom(&lUserCreateRoomRsp)
}

func (self *Table) ResUserEnterRoomRqst(uid uint64, isEnterSuccess bool) {
	logger.Info("Table:ResUserEnterRoomRqst:<ENTER>, uid:%d", uid)
	lUserEnterRoomRsp := rpc.CSUserEnterRoomRsp{}
	lUserEnterRoomRsp.RoomId = &(self.mRoomCfg.SRoomId)
	lUserEnterRoomRsp.RoomType = &(self.mRoomCfg.SRoomType)
	lUserEnterRoomRsp.EnterRoomSuccess = &(isEnterSuccess)
	lUserEnterRoomRsp.MinCurrencyValue = &(self.mRoomCfg.SMinCurrencyValue)

	self.mRoom.SendMsgToRoom(&lUserEnterRoomRsp)
}

func (self *Table) AnswerClientError(value uint32, uid uint64) {
	logger.Info("player:AnswerClientError")
	var l uint32 = 1
	lCommonErrMsg := rpc.CSCommonErrMsg{}
	lCommonErrMsg.ErrorCode = &(value)
	lCommonErrMsg.RqstCmdID = &(l)
	lCommonErrMsg.Uid = &(uid)
}

func (self *Table) ClearTable(b bool) {
	if self.mReadyTimer != nil {
		self.mReadyTimer.Stop()
	}
}

func (self *Table) ClearMatchInfo(IsAllMatchOver bool) {
	self.mDice1 = uint32(0)
	self.mDice1 = uint32(0)
	self.mZhuang = TABLE_SEAT_NONE
	self.mHuanSanZhangType = rpc.TASK_HSZ_TYPE_TASK_HSZ_TYPE_NONE
	self.mLastMoPaiUserSeatId = uint32(TABLE_SEAT_NONE)
	self.mLastDealUserSeatId = uint32(TABLE_SEAT_NONE)
	self.mLastGangUserSeatId = uint32(TABLE_SEAT_NONE)
	self.mWaitBuGangUserSeatId = uint32(TABLE_SEAT_NONE)
	self.mCurRelationSeatId = uint32(TABLE_SEAT_NONE)
	self.mWaitBuGangTagMj = uint32(MJ_BACK)
	self.SetTableState(rpc.TableState_Table_State_Init)
	//TODO mCacheOperateInfo.clear()
	self.mMjCardMgr.Init()

	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		self.mUsers[i].ClearMatchInfo(IsAllMatchOver)
	}

	self.mDissolveApplySeatId = uint32(TABLE_SEAT_NONE)
	self.mFinalDecision = rpc.TableFinalDecision_Table_Decision_None
	self.mEscapeSettlement = false

}

func (self *Table) WaitReadyTimeOut() {
	if self.mRoomCfg.SIsAllowAgentOperator == true {
		self.SetTableState(rpc.TableState_Table_State_Game_End)
	}
	return
}

func (self *Table) DestroyTable() uint32 {
	if self.mTableState == rpc.TableState_Table_State_Dissolve {
		return errorValue.ERET_OK
	}

	self.SetTableState(rpc.TableState_Table_State_Dissolve)

	self.NotifyAllUserTableChangInfo()
	self.NotifyAllUserTableDismiss()

	return errorValue.ERET_OK
}

//检查玩家金币，金币不足时进入等待充值状态
func (self *Table) CheckUserCurrencyForMatch() {
	//TODO:
	return
}

func (self *Table) NotifyAllUserReady(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	var lSeatId uint32 = uint32(seatId)
	lUserReadyGameNotify := rpc.CSUserReadyGameNotify{}
	lUserReadyGameNotify.ReadySeatId = &(lSeatId)
	self.mRoom.SendMsgToRoom(&lUserReadyGameNotify)
	return lRet
}
func (self *Table) NotifyAllUserTableChangInfo() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() == true {
			self.NotifyUserTableInfo(i)
		}
	}
	return lRet
}

func (self *Table) NotifyUserTableInfo(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	if self.mUsers[seatId].IsOnActivestate() == false {
		return lRet
	}

	lUserTableInfoChangeNotify := rpc.CSUserTableInfoChangeNotify{}
	lUserTableInfoChangeNotify.RoomType = &(self.mRoomCfg.SRoomType)
	lUserTableInfoChangeNotify.RoomId = &(self.mRoomCfg.SRoomId)
	lUserTableInfoChangeNotify.RoomOwnerUid = &(self.mRoomCfg.SRoomOwnerId)
	ltableinfo := lUserTableInfoChangeNotify.Info

	ltableinfo.BankerSeatId = &(self.mZhuang)
	ltableinfo.FirstDice = &(self.mDice1)
	ltableinfo.SecondDice = &(self.mDice2)
	lLeftCards := self.mMjCardMgr.GetLeftCardsNum()
	ltableinfo.CardNum = &(lLeftCards)
	ltableinfo.TableState = &(self.mTableState)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsFree() {
			continue
		}

		//TODO : 获取数据库数据,必须要访问数据库才行
		lSeatIdTmp := uint32(i)
		lUidTmp := uint64(1)
		lSexTmp := int32(1)
		lUserGoldTmp := int64(1111)
		lCustomRoomPointTmp := int64(1111)
		lConnectIpTmp := uint32(11)
		lLongitudeTmp := uint32(11)
		lLatitudeTmp := uint32(11)
		lStartMatchCurrencyValue := int64(1111)
		lIsOffLineFlag := self.mUsers[i].IsOffLine()
		lNickNameTmp := "aaaa"
		lHeadImage := "bbbb"
		lIdentityTmp := rpc.USER_IDENTITY_TYPE_IDENTITY_TYPE_NORMAL
		lUserStateTmp := self.mUsers[i].GetUserState()

		lSeatInfo := rpc.SeatInfo{}
		lSeatInfo.SeatId = &(lSeatIdTmp)
		lSeatInfo.UserState = &(lUserStateTmp)
		lSeatInfo.OffLineFlag = &(lIsOffLineFlag)
		lSeatInfo.BaseInfo.Uid = &(lUidTmp)
		lSeatInfo.BaseInfo.NickName = &(lNickNameTmp)
		lSeatInfo.BaseInfo.Sex = &(lSexTmp)
		lSeatInfo.BaseInfo.UserGold = &(lUserGoldTmp)
		lSeatInfo.BaseInfo.CustomRoomPoint = &(lCustomRoomPointTmp)
		lSeatInfo.BaseInfo.HeadImage = &(lHeadImage)
		lSeatInfo.BaseInfo.Identity = &(lIdentityTmp)
		lSeatInfo.BaseInfo.ConnectIp = &(lConnectIpTmp)
		lSeatInfo.BaseInfo.Longitude = &(lLongitudeTmp)
		lSeatInfo.BaseInfo.Latitude = &(lLatitudeTmp)
		lSeatInfo.StartMatchCurrencyValue = &(lStartMatchCurrencyValue)
		ltableinfo.SeatInfos = append(ltableinfo.SeatInfos, &(lSeatInfo))
	}
	lNum := uint64(self.mFinishedMatchNum)
	ltableinfo.SettlementId = &(self.mSettlementId)
	ltableinfo.FinishedMatchNum = &(lNum)
	ltableinfo.TotalMatchNum = &(self.mRoomCfg.SMatchNum)
	//TODO : 等待充值时间计算
	lWaitRechargeTimeOut := uint32(0)
	ltableinfo.WaitRechargeTimeOut = &(lWaitRechargeTimeOut)

	self.mRoom.SendMsgToRoom(&lUserTableInfoChangeNotify)
	return lRet
}

func (self *Table) NotifyAllUserFaPai() uint32 {
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() == false {
			continue
		}

		lUserMjAssignNotify := rpc.CSUserMjAssignNotify{}
		lUserMjAssignNotify.SeatId = &(self.mUsers[i].mSeatId)
		lUserMjAssignNotify.BankerSeatId = &(self.mZhuang)
		lUserMjAssignNotify.FirstDice = &(self.mDice1)
		lUserMjAssignNotify.SecondDice = &(self.mDice2)
		//lUserMjAssignNotify.Uid = &(self.mUsers[i].mUid)
		j := 0
		for lTmp := range self.mUsers[i].mHandCards {
			lUserMjAssignNotify.Mjs[j] = uint32(lTmp)
			j += 1
		}
		self.mRoom.SendMsgToRoom(&lUserMjAssignNotify)
	}
	return errorValue.ERET_OK
}

func (self *Table) NotifyAllUserWaitRejectSuit() uint32 {
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() {
			continue
		}

		lUserMjWaitForOperateNotify := rpc.CSUserMjWaitForOperateNotify{}
		lUserMjWaitForOperateNotify.RelationSeatId = &(self.mUsers[i].mSeatId)
		lWaitRejectSuitTimeout := uint32(common.WaitRejectSuitTimeout)
		lUserMjWaitForOperateNotify.OperateTimeOut = &(lWaitRejectSuitTimeout)

		for e := self.mUsers[i].mOperateMask.Front(); e != nil; e = e.Next() {
			lUserMjWaitForOperateNotify.OperateMask[i] = e.Value.(rpc.MJ_OPERATE_MASK)
		}
		self.mRoom.SendMsgToRoom(&lUserMjWaitForOperateNotify)
	}

	return errorValue.ERET_OK
}

func (self *Table) NotifyAllUserDeal(seatId uint32, tagMj uint32) uint32 {
	lOperateMask := rpc.MJ_OPERATE_MASK_MJ_OPERATE_MASK_DEAL

	lUserMjOperateNotify := rpc.CSUserMjOperateNotify{}
	lUserMjOperateNotify.OperateMask = &(lOperateMask)
	lUserMjOperateNotify.OperateSeatId = &(seatId)
	lUserMjOperateNotify.TagMj = &(tagMj)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() == false {
			continue
		}
		//lUserMjOperateNotify.Uid = &(self.mUsers[i].mUid)
		self.mRoom.SendMsgToRoom(&lUserMjOperateNotify)
	}
	return errorValue.ERET_OK
}

func (self *Table) NotifyUserWaitOperateForDealfuwCard(seatId uint32) uint32 {
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() == false {
			continue
		}

		lUserMjWaitForOperateNotify := rpc.CSUserMjWaitForOperateNotify{}
		lUserMjWaitForOperateNotify.RelationSeatId = &(seatId)
		lWaitOperateForDealTimeout := uint32(common.WaitOperateForDealTimeout)
		lUserMjWaitForOperateNotify.OperateTimeOut = &(lWaitOperateForDealTimeout)
		//lUserMjWaitForOperateNotify.Uid = &(self.mUsers[i].mUid)
		lOperateMask := self.mUsers[i].GetOperateMask()
		var index int = 0
		for e := lOperateMask.Front(); e != nil; e = e.Next() {
			lUserMjWaitForOperateNotify.OperateMask[index] = e.Value.(rpc.MJ_OPERATE_MASK)
			index++
		}
		self.mRoom.SendMsgToRoom(&lUserMjWaitForOperateNotify)
	}
	return errorValue.ERET_OK
}

//TODO:扣除单局房费
func (self *Table) DanJuRoomCostSettlement() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)

	//self.mRoom.SendMsgToRoom(&lUserMjWaitForOperateNotify)
	return lRet
}

func (self *Table) UserMoPai(seatId uint32) uint32 {
	var lRet uint32 = errorValue.ERET_OK
	if self.mMjCardMgr.IsEmpty() == true {
		lRet = self.GameEnd()
		return lRet
	}

	var lCard uint32
	self.mMjCardMgr.MoPai(&lCard)
	self.mLastMoPaiUserSeatId = seatId
	self.mUsers[int(seatId)].AddCard(lCard)

	lMjCountInPool := self.mMjCardMgr.GetLeftCardsNum()
	lOperateMask := self.mUsers[int(seatId)].CalculateOperateMask(lMjCountInPool)
	self.mUsers[int(seatId)].SetOperateMaskList(lOperateMask)
	self.NotifyUserMoPai(seatId, lCard)
	return errorValue.ERET_OK
}

func (self *Table) NotifyUserMoPai(seatId uint32, tagMj uint32) uint32 {
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() == false {
			continue
		}
		lUserMjGetNotify := rpc.CSUserMjGetNotify{}
		lUserMjGetNotify.RelationSeatId = &seatId
		lWaitOperateForDealTimeout := uint32(common.WaitOperateForDealTimeout) //*uint32
		lUserMjGetNotify.OperateTimeOut = &(lWaitOperateForDealTimeout)        //*uint32
		if seatId == self.mUsers[i].mSeatId {
			lUserMjGetNotify.TagMj = &tagMj
		} else {
			BaiPai := uint32(MJ_BACK)
			lUserMjGetNotify.TagMj = &(BaiPai)
		}
		lOperateMask := self.mUsers[i].GetOperateMask()
		for e := lOperateMask.Front(); e != nil; e = e.Next() {
			lMask := e.Value.(rpc.MJ_OPERATE_MASK)
			lUserMjGetNotify.OperateMask = append(lUserMjGetNotify.OperateMask, lMask)
		}
		self.mRoom.SendMsgToRoom(&lUserMjGetNotify)
	}
	return errorValue.ERET_OK
}
func (self *Table) NotifyUserTableCardInfo(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)

	if self.mUsers[seatId].IsOnActivestate() == false {
		return lRet
	}
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsFree() {
			continue
		}
		lTagSeatIdTmp := uint32(seatId)
		lSeatIdTmp := uint32(i)
		lUserTableCardInfoNotify := rpc.CSUserTableCardInfoNotify{}
		lUserTableCardInfoNotify.RoomType = &(self.mRoomCfg.SRoomType) //*ROOM_TYPE
		lUserTableCardInfoNotify.RoomId = &(self.mRoomCfg.SRoomId)     //*uint32
		lUserTableCardInfoNotify.SeatId = &(lSeatIdTmp)                //*uint32
		lUserTableCardInfoNotify.TagSeatId = &(lTagSeatIdTmp)          //*uint32
		lUserState := self.mUsers[i].GetUserState()
		lUserTableCardInfoNotify.UserState = &(lUserState) //*MjUserState
		lRejectSuit := self.mUsers[i].GetRejectSuit()
		lUserTableCardInfoNotify.RejectSuit = &(lRejectSuit) //*uint32
		for lIts := range *(self.mUsers[i].GetDealCards()) {
			lTmp := uint32(lIts)
			lUserTableCardInfoNotify.DealCards = append(lUserTableCardInfoNotify.DealCards, lTmp)
		}
		for lIts := range *(self.mUsers[i].GetMingGangCards()) {
			lUserTableCardInfoNotify.MingCards = append(lUserTableCardInfoNotify.MingCards, uint32(lIts))
		}
		for lIts := range *(self.mUsers[i].GetAnGangCards()) {
			lUserTableCardInfoNotify.AnGangCards = append(lUserTableCardInfoNotify.AnGangCards, uint32(lIts))
		}
		for lIts := range *(self.mUsers[i].GetBuGangCards()) {
			lUserTableCardInfoNotify.BuGangCards = append(lUserTableCardInfoNotify.BuGangCards, uint32(lIts))
		}
		for lIts := range *(self.mUsers[i].GetPengCards()) {
			lUserTableCardInfoNotify.PengCards = append(lUserTableCardInfoNotify.PengCards, uint32(lIts))
		}
		if seatId == i {
			for lIts := range self.mUsers[i].GetHandCards() {
				lUserTableCardInfoNotify.HandCards = append(lUserTableCardInfoNotify.HandCards, uint32(lIts))
			}
			for lIts := range *(self.mUsers[i].GetHuanCards()) {
				lUserTableCardInfoNotify.HuanCards = append(lUserTableCardInfoNotify.HuanCards, uint32(lIts))
			}
			for lIts := range *(self.mUsers[i].GetHuanInCards()) {
				lUserTableCardInfoNotify.HuanInCards = append(lUserTableCardInfoNotify.HuanInCards, uint32(lIts))
			}
		} else {
			for range self.mUsers[i].GetHandCards() {
				lUserTableCardInfoNotify.HandCards = append(lUserTableCardInfoNotify.HandCards, MJ_BACK)
			}
			for range *(self.mUsers[i].GetHuanCards()) {
				lUserTableCardInfoNotify.HuanCards = append(lUserTableCardInfoNotify.HuanCards, MJ_BACK)
			}
			for range *(self.mUsers[i].GetHuanInCards()) {
				lUserTableCardInfoNotify.HuanInCards = append(lUserTableCardInfoNotify.HuanInCards, MJ_BACK)
			}
		}
		if self.mUsers[i].IsHaveOperateRecord() {
			var m MjOperateRecord = (self.mUsers[i].GetLastOperateRecord()).(MjOperateRecord)
			lUserTableCardInfoNotify.LastOperateRecord.SeatId = &(m.SeatId)           //*uint32
			lUserTableCardInfoNotify.LastOperateRecord.OperateMask = &(m.OperateMask) //*MJ_OPERATE_MASK
			lUserTableCardInfoNotify.LastOperateRecord.TagMj = &(m.TagMj)             //*uint32
			lUserTableCardInfoNotify.LastOperateRecord.TagMj1 = &(m.TagMj1)           //*uint32
			lUserTableCardInfoNotify.LastOperateRecord.TagMj2 = &(m.TagMj2)           //*uint32
		}
		lSettlementType := self.mUsers[i].GetSettlementType()
		lUserTableCardInfoNotify.SettlementType = &(lSettlementType)
		self.mRoom.SendMsgToRoom(&lUserTableCardInfoNotify)
	}
	return lRet
}

func (self *Table) NotifyUserCurOperate(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	if self.mTableState == rpc.TableState_Table_State_Gaming || (self.mTableState == rpc.TableState_Table_State_Wait_Dissolve && self.mPrevTableState == rpc.TableState_Table_State_Gaming) {
		lUserMjWaitForOperateNotify := rpc.CSUserMjWaitForOperateNotify{}

		if self.mCurRelationSeatId == uint32(TABLE_SEAT_NONE) {
			lSeatTmp := uint32(seatId)
			lUserMjWaitForOperateNotify.RelationSeatId = &(lSeatTmp)
		} else {
			lUserMjWaitForOperateNotify.RelationSeatId = &(self.mCurRelationSeatId)
		}

		lTimeOut := self.mWaitOperateTimerOutTime - time.Now().UnixNano()
		if lTimeOut > int64(0) {
			lOperateTimeOut := uint32(lTimeOut)
			lUserMjWaitForOperateNotify.OperateTimeOut = &(lOperateTimeOut)
		} else {
			lOperateTimeOut := uint32(0)
			lUserMjWaitForOperateNotify.OperateTimeOut = &(lOperateTimeOut)
		}

		lOperateMask := self.mUsers[seatId].GetOperateMask()
		for e := lOperateMask.Front(); e != nil; e = e.Next() {
			lIts := e.Value.(rpc.MJ_OPERATE_MASK)
			lUserMjWaitForOperateNotify.OperateMask = append(lUserMjWaitForOperateNotify.OperateMask, lIts)
		}
		self.mRoom.SendMsgToRoom(&lUserMjWaitForOperateNotify)
	}
	return lRet
}

func (self *Table) NotifyUserDissolveApplyInfo(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	lUserDissolveApplyNotify := rpc.CSUserTableDissolveApplyNotify{}
	lUserDissolveApplyNotify.ApplySeatId = &(self.mDissolveApplySeatId) //*uint32
	lUserDissolveApplyNotify.FinalDecision = &(self.mFinalDecision)     //*TableFinalDecision
	lTimeOut := self.mWaitOperateTimerOutTime - time.Now().UnixNano()
	if lTimeOut > int64(0) {
		lApplyTimeOut := uint32(lTimeOut)
		lUserDissolveApplyNotify.ApplyTimeOut = &(lApplyTimeOut)
	} else {
		lApplyTimeOut := uint32(0)
		lUserDissolveApplyNotify.ApplyTimeOut = &(lApplyTimeOut)
	}
	for p1 := range self.mDisagreeDissolveApplySeatIds {
		p := uint32(p1)
		lUserDissolveApplyNotify.DisagreeApplySeatIds = append(lUserDissolveApplyNotify.DisagreeApplySeatIds, p)
	}
	for p1 := range self.mAgreeDissolveApplySeatIds {
		p := uint32(p1)
		lUserDissolveApplyNotify.AgreeApplySeatIds = append(lUserDissolveApplyNotify.AgreeApplySeatIds, p)
	}
	self.mRoom.SendMsgToRoom(&lUserDissolveApplyNotify)
	return lRet
}

func (self *Table) NotifyUserReadyShutDownInfo(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	lUserTableReadyShutDownNotify := rpc.CSUserTableReadyShutDownNotify{}

	lTimeOut := self.mWaitOperateTimerOutTime - time.Now().UnixNano()
	if lTimeOut > int64(0) {
		lReadyShutDwonTimeOut := uint32(lTimeOut)
		lUserTableReadyShutDownNotify.ReadyShutDwonTimeOut = &(lReadyShutDwonTimeOut)
	} else {
		lReadyShutDwonTimeOut := uint32(0)
		lUserTableReadyShutDownNotify.ReadyShutDwonTimeOut = &(lReadyShutDwonTimeOut)
	}

	blTrue := true
	lUserTableReadyShutDownNotify.ReadyShutDown = &(blTrue)
	self.mRoom.SendMsgToRoom(&lUserTableReadyShutDownNotify)
	return lRet
}

func (self *Table) NotifyAllUserTableDismiss() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() == false {
			continue
		}

		lUserTableDIsmissNotify := rpc.CSUserTableDIsmissNotify{}
		lUserTableDIsmissNotify.RoomType = &(self.mRoomCfg.SRoomType)
		lUserTableDIsmissNotify.RoomId = &(self.mRoomCfg.SRoomId)
		lSeatId := uint32(i)
		lUserTableDIsmissNotify.SeatId = &(lSeatId)
		self.mRoom.SendMsgToRoom(&lUserTableDIsmissNotify)
	}
	return lRet
}

func (self *Table) NotifyAllUserHandCards() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsMatch() {
			continue
		}
		lSettlementInfoNotify := rpc.CSUserMjLastHandNotify{}
		lSettlementInfoNotify.RoomType = &(self.mRoomCfg.SRoomType)

		lSettlementInfoNotify.RoomId = &(self.mRoomCfg.SRoomId)
		lSeatId := uint32(i)
		lSettlementInfoNotify.SeatId = &(lSeatId)
		lHandCards := self.mUsers[i].GetHandCards()
		for p1 := range lHandCards {
			p := uint32(p1)
			lSettlementInfoNotify.Mjs = append(lSettlementInfoNotify.Mjs, p)
		}
		for j := 0; j < MAX_PLAYER_NUM_PER_TABLE; j++ {
			if self.mUsers[j].IsOnActivestate() {
				continue
			}
			self.mRoom.SendMsgToRoom(&lSettlementInfoNotify)
		}
	}
	return lRet
}

func (self *Table) NotifyUserWaitOperateForDealCard(seatId uint32) uint32 {
	lRet := uint32(errorValue.ERET_OK)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsOnActivestate() == false {
			continue
		}

		lUserMjWaitForOperateNotify := rpc.CSUserMjWaitForOperateNotify{}
		lUserMjWaitForOperateNotify.RelationSeatId = &(seatId) //*uint32
		lTimeOut := uint32(common.WaitOperateForDealTimeout)
		lUserMjWaitForOperateNotify.OperateTimeOut = &(lTimeOut) //*uint32
		var lOperateMask *list.List = self.mUsers[i].GetOperateMask()
		for e := lOperateMask.Front(); e != nil; e = e.Next() {
			lValue := e.Value.(rpc.MJ_OPERATE_MASK)
			lUserMjWaitForOperateNotify.OperateMask = append(lUserMjWaitForOperateNotify.OperateMask, lValue)
		}
		self.mRoom.SendMsgToRoom(&lUserMjWaitForOperateNotify)
	}
	return lRet
}

func (self *Table) NotifyUserAllSettlementInfo(seatId uint32) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	if self.mUsers[seatId].IsMatch() == false {
		return lRet
	}

	lSettlementInfoNotify := rpc.CSUserMjSettlementInfoNotify{}
	lSettlementInfos := self.mUsers[seatId].GetSettlementInfo()
	lSettlementInfoNotify.RoomType = &(self.mRoomCfg.SRoomType)
	lSettlementInfoNotify.RoomId = &(self.mRoomCfg.SRoomId)
	lSeatIdTmp := uint32(seatId)
	lSettlementInfoNotify.SeatId = &(lSeatIdTmp)
	for p := lSettlementInfos.Front(); p != nil; p = p.Next() {
		var lTmp MajSettlementInfo = p.Value.(MajSettlementInfo)

		lInfo := rpc.MjSettlementInfo{}
		lInfo.Type = &(lTmp.Type)         //*MJ_SETTLEMENT_TYPE
		lInfo.HuType = &(lTmp.HuType)     //*uint32
		lInfo.Multiple = &(lTmp.Multiple) //*uint32
		lInfo.GenCount = &(lTmp.GenCount) //*int32

		for q := lTmp.Detail.Front(); q != nil; q = q.Next() {
			var lDetail MajSettlementDetail = q.Value.(MajSettlementDetail)
			lSettlementDetail := rpc.MjSettlementDetail{}
			lSeatIdTmp := uint32(lDetail.SeatId)
			lSettlementDetail.SeatId = &(lSeatIdTmp)
			lSettlementDetail.Value = &(lDetail.ChangeValue)
			lInfo.Detail = append(lInfo.Detail, (&lSettlementDetail)) //[]*MjSettlementDetail
		}

		lSettlementInfoNotify.SettlementInfo = append(lSettlementInfoNotify.SettlementInfo, (&lInfo))
	}
	self.mRoom.SendMsgToRoom(&lSettlementInfoNotify)
	return lRet
}
func (self *Table) NotifyAllUserSettlementInfo() uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsMatch() == false {
			continue
		}
		lSettlementInfoNotify := rpc.CSUserMjSettlementInfoNotify{}
		lSettlementInfos := self.mUsers[i].GetSettlementInfo()
		lSettlementInfoNotify.RoomType = &(self.mRoomCfg.SRoomType)
		lSettlementInfoNotify.RoomId = &(self.mRoomCfg.SRoomId)
		li := uint32(i)
		lSettlementInfoNotify.SeatId = &(li) //&((uint32)i)

		for p := lSettlementInfos.Front(); p != nil; p = p.Next() {
			var lTmp MajSettlementInfo = p.Value.(MajSettlementInfo)

			lInfo := rpc.MjSettlementInfo{}
			lInfo.Type = &(lTmp.Type)         //*MJ_SETTLEMENT_TYPE
			lInfo.HuType = &(lTmp.HuType)     //*uint32
			lInfo.Multiple = &(lTmp.Multiple) //*uint32
			lInfo.GenCount = &(lTmp.GenCount) //*int32

			for q := lTmp.Detail.Front(); q != nil; q = q.Next() {
				var lDetail MajSettlementDetail = q.Value.(MajSettlementDetail)
				lSettlementDetail := rpc.MjSettlementDetail{}
				lSeatIdTmp := uint32(lDetail.SeatId)
				lSettlementDetail.SeatId = &(lSeatIdTmp)
				lSettlementDetail.Value = &(lDetail.ChangeValue)
				lInfo.Detail = append(lInfo.Detail, (&lSettlementDetail)) //[]*MjSettlementDetail
			}

			lSettlementInfoNotify.SettlementInfo = append(lSettlementInfoNotify.SettlementInfo, (&lInfo))
		}
		for j := 0; j < MAX_PLAYER_NUM_PER_TABLE; j++ {
			if self.mUsers[j].IsOnActivestate() == false {
				continue
			}
			self.mRoom.SendMsgToRoom(&lSettlementInfoNotify)
		}
	}
	return lRet
}

func (self *Table) NotifyGameAllOperation(seatId int) uint32 {
	var lRet uint32 = uint32(errorValue.ERET_OK)
	gameAllOperationNotify := rpc.CSUserGameAllOperationNotify{}
	gameAllOperationNotify.ZhuangSeatId = &(self.mLastGameInfo.Zhuang) //*uint32
	gameAllOperationNotify.Dice1 = &(self.mLastGameInfo.Dice1)         //*uint32
	gameAllOperationNotify.Dice2 = &(self.mLastGameInfo.Dice2)         //*uint32
	gameAllOperationNotify.HSZType = &(self.mRoomCfg.SHuanSanZhangType)
	gameAllOperationNotify.TaskHSZType = &(self.mHuanSanZhangType)      //*uint32
	gameAllOperationNotify.FinishedMatchNum = &(self.mFinishedMatchNum) //*uint32

	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		if self.mUsers[i].IsMatch() == false {
			continue
		}
		lastGameInfo := self.mUsers[i].GetLastGameInfo()
		userGameInitInfo := rpc.UserGameInitInfo{}
		li := uint32(i)
		userGameInitInfo.SeatId = &(li)                          //*uint32
		userGameInitInfo.RejectSuit = &(lastGameInfo.RejectSuit) //*uint32

		for c := range lastGameInfo.InitHandCards {
			var lTmp uint32 = uint32(c)
			userGameInitInfo.InitHandCards = append(userGameInitInfo.InitHandCards, lTmp)
		}
		for c := range lastGameInfo.HuanCards {
			var lTmp uint32 = uint32(c)
			userGameInitInfo.HuanCards = append(userGameInitInfo.HuanCards, lTmp)
		}
		for c := range lastGameInfo.HuanInCards {
			var lTmp uint32 = uint32(c)
			userGameInitInfo.HuanInCards = append(userGameInitInfo.HuanInCards, lTmp)
		}
		gameAllOperationNotify.GameInitInfo = append(gameAllOperationNotify.GameInitInfo, &userGameInitInfo)
	}
	//GameOperation    //[]*UserGameOperation
	allOperation := self.mUsers[0].GetLastGameInfo().OperateRecord
	for op := allOperation.Front(); op != nil; op = op.Next() {
		var lTmpOp MjOperateRecord = MjOperateRecord{}
		gameOperation := rpc.UserGameOperation{}
		lSeatIdTmp := uint32(lTmpOp.SeatId)
		gameOperation.SeatId = (&lSeatIdTmp) //*uint32
		lOperateMask := uint32(lTmpOp.OperateMask)
		gameOperation.OperateMask = (&lOperateMask) //*uint32
		lTagMj := uint32(lTmpOp.TagMj)
		gameOperation.TagMj = (&lTagMj) //*uint32
		gameAllOperationNotify.GameOperation = append(gameAllOperationNotify.GameOperation, &gameOperation)
	}

	if seatId >= 0 && seatId <= 3 {
		//通知个人
		if self.mUsers[seatId].IsOnActivestate() {
			self.mRoom.SendMsgToRoom(&gameAllOperationNotify)
			return lRet
		}
	} else {
		//通知所有人
		for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
			if self.mUsers[i].IsOnActivestate() == false {
				continue
			}
			self.mRoom.SendMsgToRoom(&gameAllOperationNotify)
		}
		return lRet
	}
	return lRet
}

func (self *Table) CacheLastGameInfo() {
	self.mLastGameInfo.Zhuang = self.mZhuang
	self.mLastGameInfo.Dice1 = self.mDice1
	self.mLastGameInfo.Dice2 = self.mDice2
	self.mLastGameInfo.RemainCardNum = self.mMjCardMgr.GetLeftCardsNum()
	self.mLastGameInfo.SettlementId = (self.mSettlementId)
	self.mLastGameInfo.HSZType = (self.mRoomCfg.SHuanSanZhangType)
	for i := 0; i < MAX_PLAYER_NUM_PER_TABLE; i++ {
		self.mUsers[i].CacheLastGameInfo()
	}
	return
}
