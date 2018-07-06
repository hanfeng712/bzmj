package xzmj

import (
	"container/heap"
	"container/list"
	"rpc"
)

//	"errorValue"

const (
	Mj_User_Action_Init     = 0
	Mj_User_Action_MoPai    = 1
	Mj_User_Action_DealPai  = 2
	Mj_User_Action_PengPai  = 3
	Mj_User_Action_AnGang   = 4
	Mj_User_Action_MingGang = 5
	Mj_User_Action_BuGang   = 6
	Mj_User_Action_HuPai    = 7
	Mj_User_Action_GuoPai   = 8
)

const (
	MJ_BACK      = 0
	MJ_WANG_SUIT = 1
	MJ_TONG_SUIT = 2
	MJ_TIAO_SUIT = 3

	MJ_1_WANG = 11
	MJ_2_WANG = 12
	MJ_3_WANG = 13
	MJ_4_WANG = 14
	MJ_5_WANG = 15
	MJ_6_WANG = 16
	MJ_7_WANG = 17
	MJ_8_WANG = 18
	MJ_9_WANG = 19

	MJ_1_TONG = 21
	MJ_2_TONG = 22
	MJ_3_TONG = 23
	MJ_4_TONG = 24
	MJ_5_TONG = 25
	MJ_6_TONG = 26
	MJ_7_TONG = 27
	MJ_8_TONG = 28
	MJ_9_TONG = 29

	MJ_1_TIAO = 31
	MJ_2_TIAO = 32
	MJ_3_TIAO = 33
	MJ_4_TIAO = 34
	MJ_5_TIAO = 35
	MJ_6_TIAO = 36
	MJ_7_TIAO = 37
	MJ_8_TIAO = 38
	MJ_9_TIAO = 39
)

const (
	//		杠牌数量
	GANG_MJ_COUNT = 4
	//		碰牌数量
	PENG_MJ_COUNT = 3
	//		对子数量
	DUIZI_MJ_COUNT = 2
	//		顺子麻将数量
	HUA_MJ_COUNT = 3
	//		起手麻将数量
	HAND_MJ_COUNT = 14
	//		座位号的最大值
	SIT_NUM_MAX = 3
	//
	MAX_PLAYER_NUM_PER_TABLE = 4
	//
	TABLE_SEAT_NONE = 100
)

func MJ_MAKE(suit uint32, rank uint32) uint32 {
	mj := suit*10 + rank
	return mj
}

func MJ_GET_SUIT(mj uint32) uint32 {
	return (mj % 100) / 10
}

type IntHeap []uint32

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(uint32))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *IntHeap) Clear() {
	for h.Len() > 0 {
		h.Pop()
	}
}

func (h *IntHeap) Find(Tag uint32) int {
	var i int = 0
	for ; i < h.Len(); i++ {
		if (*h)[i] == Tag {
			return i
		}
	}
	return i
}

///////////////////////////////////////////////////////////////////////////////////
type IntOperateHeap []rpc.MJ_OPERATE_MASK

func (h IntOperateHeap) Len() int           { return len(h) }
func (h IntOperateHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntOperateHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntOperateHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(rpc.MJ_OPERATE_MASK))
}

func (h *IntOperateHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *IntOperateHeap) Clear() {
	for h.Len() > 0 {
		h.Pop()
	}
}

func (h *IntOperateHeap) Find(Tag rpc.MJ_OPERATE_MASK) int {
	var i int = 0
	for ; i < h.Len(); i++ {
		if (*h)[i] == Tag {
			return i
		}
	}
	return i
}

type MjOperateRecord struct {
	SeatId      uint32
	OperateMask rpc.MJ_OPERATE_MASK
	TagMj       uint32
	TagMj1      uint32
	TagMj2      uint32
}

type MajSettlementDetail struct {
	SeatId      int
	ChangeValue int32
}

type MajSettlementInfo struct {
	Type     rpc.MJ_SETTLEMENT_TYPE
	HuType   uint32
	GenCount int32
	Multiple uint32
	Detail   *list.List
}

type MjLastTableGameInfo struct {
	Zhuang        uint32
	Dice1         uint32
	Dice2         uint32
	RemainCardNum uint32
	SettlementId  uint64
	HSZType       rpc.HUAN_SAN_ZHANG_TYPE
	TaskHSZType   rpc.TASK_HSZ_TYPE
}

func NewMjLastTableGameInfo() *MjLastTableGameInfo {
	lTmp := MjLastTableGameInfo{}
	lTmp.Zhuang = uint32(0)
	lTmp.Dice1 = uint32(0)
	lTmp.Dice2 = uint32(0)
	lTmp.RemainCardNum = uint32(0)
	lTmp.SettlementId = uint64(0)
	lTmp.HSZType = rpc.HUAN_SAN_ZHANG_TYPE_HUAN_SAN_ZHANG_TYPE_NONE
	lTmp.TaskHSZType = rpc.TASK_HSZ_TYPE_TASK_HSZ_TYPE_NONE
	return &lTmp
}

type MjLastUserGameInfo struct {
	RejectSuit     uint32
	HuCard         uint32  //胡牌
	DealCards      IntHeap //玩家的出牌
	HandCards      IntHeap //玩家的手牌
	MingGangCards  IntHeap //玩家的明杠
	AnGangCards    IntHeap //玩家的暗杠
	BuGangCards    IntHeap //玩家的补杠
	PengCards      IntHeap //玩家的碰
	SettlementInfo *list.List
	InitHandCards  IntHeap
	HuanCards      IntHeap
	HuanInCards    IntHeap
	OperateRecord  *list.List
}

func NewMjLastUserGameInfo() *MjLastUserGameInfo {
	lTmp := MjLastUserGameInfo{}
	lTmp.RejectSuit = uint32(0)
	lTmp.HuCard = uint32(0)

	heap.Init(&(lTmp.DealCards))

	heap.Init(&(lTmp.HandCards))

	heap.Init(&(lTmp.DealCards))

	heap.Init(&(lTmp.MingGangCards))

	heap.Init(&(lTmp.AnGangCards))

	heap.Init(&(lTmp.BuGangCards))

	heap.Init(&(lTmp.PengCards))

	lTmp.SettlementInfo = list.New()
	lTmp.SettlementInfo.Init()

	heap.Init(&(lTmp.InitHandCards))

	heap.Init(&(lTmp.HuanCards))

	heap.Init(&(lTmp.HuanInCards))

	lTmp.OperateRecord = list.New()
	lTmp.OperateRecord.Init()

	return &lTmp
}
