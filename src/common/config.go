package common

//"rpc"

//gateserver配置
type LobbyServerCfg struct {
	GsIpForClient string
	GsIpForServer string
	DebugHost     string
	GcTime        uint8
}

/*
const (
	ROOM_TICK_TIME            = 2000
	WaitReadyTimeout          = 15000
	WaitRejectSuitTimeout     = 15000
	RenewRoomTimeout          = 240000
	WaitOperateTimeout        = 4000
	WaitOperateForDealTimeout = 4000
)


type RoomConfig struct {

	SRoomId uint32

	SRoomType                            rpc.ROOM_TYPE
	SType                                uint32
	SRoomCostCurrencyType                uint32 //房间花费货币类型
	SRoomCost                            uint32 //房间创建费用
	STableCostCurrencyType               uint32 //桌子花费货币类型
	STableCost                           uint32 //桌费
	SMatchCurrencyType                   uint32 //比赛花费货币类型
	SMatchNum                            uint32 //比赛局数，每次开局需要强制完成 的局数，
	SMinMatchUserNum                     uint32 //最小开赛人数
	SRoomPassword                        string //房间密码
	SDiZhu                               uint32
	SMaxPlayerNum                        uint32 //每个房间允许的最大人数
	SMaxBeiShu                           uint32
	SMaxCurrencyValue                    int64
	SMinCurrencyValue                    uint64
	SChallengeCount                      uint32 //比赛次数上限，用于调整赛
	SChallengeTimes                      uint32 //挑战体力值，用于调整赛
	SPointPerFan                         uint32 //每番对应的积分
	SMatchInitPoint                      int64  //每个比赛局初始积分
	SIsNeedReportMatchInfoToPlayer       bool   //是否需要上报比赛信息
	SIsAllowReportMultiMatchInfoToPlayer bool   //是否需要一轮比赛汇总信息到room
	SIsAllowRenew                        bool   //是否允许续约
	SIsAllowView                         bool   //是否允许围观
	SIsAllowEscape                       bool   //是否允许逃跑
	SDeposit                             uint32 //保证金(玩家逃跑时需要扣除保证金)
	SIsAllowAgentOperator                bool   //是否允许代打操作
	SIsAllowDissolveApply                bool   //是否允许解散申请
	SIsZiMoJiaFan                        bool   //是否自摸加番， 真:自摸加番，否:自摸加底
	SIsZiMoMoreThanMaxFan                bool   //自摸加番/加低是否可以大于最大番
	SIsJinGouDiao                        bool   //是否允许金钩吊
	SIsHaiDiLaoYue                       bool   //是否允许海底捞月
	SIsYaoJiu                            bool   //是否允许幺九
	SIsJiang                             bool   //是否允许将
	SIsDaXiaoYu                          bool
	SIsDianGangHuaZiMo                   bool
	SIsMenQing                           bool
	SIsZhongZhang                        bool
	SHuanSanZhangType                    rpc.HUAN_SAN_ZHANG_TYPE
	SRoomOwnerId                         uint64 //房主id
	SCreateRoomId                        uint64 //创建房间的人id
	SRoomClubId                          uint64 //房间所属队列id
	SWaitGameStartTimeOutTime            int64  //等待房间开始时间, 超时未开始，则自动销毁
	SIsNeedOwerInRoom                    bool
	SDanJuCurrencyLimmit                 int32
	SDanJuRoomCost                       int32 //单局房费
	SIsCheckDanJuCurrencyLimmit          bool
	SIsPrivateRoom                       bool
	SInvitedUid                          [4]uint64
	SEnsureConditionType                 uint32
	SRewardCoin                          uint64
	SJoinMatchFee                        uint64
	SMatchType                           uint32
}

func NewRoomCfg() *RoomConfig {
	lRoomCfg := &RoomConfig{}
	return lRoomCfg
}
func LoadRoomConfig() *RoomConfig {
	var lRoomCfg RoomConfig

	lRoomCfg.SRoomId = 0
	lRoomCfg.SMinMatchUserNum = 4
	return &lRoomCfg
}
*/
