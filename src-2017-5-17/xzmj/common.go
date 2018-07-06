package xzmj

//	"errors"

const (
	Room_State_Init       = 1
	Room_State_Game       = 2
	Room_State_Dissolve   = 3
	Room_State_Wait_Renew = 4
)

const (
	Table_State_Init          = 1
	Table_State_Start         = 2
	Table_State_Gaming        = 3
	Table_State_Game_End      = 4
	Table_State_Dissolve      = 5
	Table_State_Wait_Renew    = 6
	Table_State_Wait_Dissolve = 7
	Table_State_Wait_Recharge = 8
)

const (
	Mj_User_State_Init          = 1
	Mj_User_State_Sit           = 2
	Mj_User_State_Ready         = 3
	Mj_User_State_Playing       = 4
	Mj_User_State_Hu            = 5
	Mj_User_State_GiveUp        = 6
	Mj_User_State_Observer_Sit  = 7
	Mj_User_State_Wait_Recharge = 8
)

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
	MJ_OPERATE_MASK_NONE           = 1
	MJ_OPERATE_MASK_DEAL           = 2
	MJ_OPERATE_MASK_GUO            = 3
	MJ_OPERATE_MASK_PENG           = 4
	MJ_OPERATE_MASK_MING_GANG      = 5
	MJ_OPERATE_MASK_AN_GANG        = 6
	MJ_OPERATE_MASK_BU_GANG        = 7
	MJ_OPERATE_MASK_REJECTSUIT     = 8
	MJ_OPERATE_MASK_HU             = 9
	MJ_OPERATE_MASK_MO_PAI         = 10
	MJ_OPERATE_MASK_GIVE_UP        = 11
	MJ_OPERATE_MASK_HUAN_SAN_ZHANG = 12
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
