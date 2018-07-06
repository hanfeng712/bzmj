package connector

/*
const PLAYERLEVEL_MAX = 200
const RANGE_LEN = 4

var TimeRange [RANGE_LEN]uint32 = [RANGE_LEN]uint32{60, 3600, 86400, 604800}
var YuanBaoCost [RANGE_LEN]uint32 = [RANGE_LEN]uint32{1, 20, 260, 1000}

func NewPosition(x, y uint32) *rpc.Position {
	p := &rpc.Position{}
	p.X = &x
	p.Y = &y

	return p
}
*/
func NewUint(x uint32) *uint32 {
	return &x
}

/*
func GetYuanBaoCountFromTime(seconds uint32) (ret uint32) {
	if seconds >= TimeRange[RANGE_LEN-2] {
		return (seconds-TimeRange[RANGE_LEN-2])/((TimeRange[RANGE_LEN-1]-TimeRange[RANGE_LEN-2])/(YuanBaoCost[RANGE_LEN-1]-YuanBaoCost[RANGE_LEN-2])) + YuanBaoCost[RANGE_LEN-2]
	}

	if seconds < TimeRange[0] {
		return YuanBaoCost[0]
	}

	for i := 0; i < RANGE_LEN-2; i++ {
		if seconds >= TimeRange[i] && seconds < TimeRange[i+1] {
			return (seconds-TimeRange[i])/((TimeRange[i+1]-TimeRange[i])/(YuanBaoCost[i+1]-YuanBaoCost[i])) + YuanBaoCost[i]
		}
	}
	return 0
}

func GetBuildingCfgByTypeId(typeId rpc.BuildingId_IdType, level uint32) *BuildingCfg {
	if level == 0 {
		return nil
	}

	switch typeId {
	case rpc.BuildingId_Barrack:
		return GetBuildingCfg("Barrack", level)
	case rpc.BuildingId_Center:
		return GetBuildingCfg("TownHall", level)
	case rpc.BuildingId_Farm:
		return GetBuildingCfg("Farm", level)
	case rpc.BuildingId_Laboratory:
		return GetBuildingCfg("Laboratory", level)
	case rpc.BuildingId_Wall:
		return GetBuildingCfg("Walls", level)
	case rpc.BuildingId_Worker:
		return GetBuildingCfg("Worker", level)
	case rpc.BuildingId_FoodStorage:
		return GetBuildingCfg("FoodStorage", level)
	case rpc.BuildingId_GoldMine:
		return GetBuildingCfg("GoldMine", level)
	case rpc.BuildingId_GoldStorage:
		return GetBuildingCfg("GoldStorage", level)
	case rpc.BuildingId_TroopHousing:
		return GetBuildingCfg("TroopHousing", level)
	case rpc.BuildingId_ArcherTower:
		return GetBuildingCfg("ArcherTower", level)
	case rpc.BuildingId_Cannon:
		return GetBuildingCfg("Cannon", level)
	case rpc.BuildingId_WizardTower:
		return GetBuildingCfg("WizardTower", level)
	case rpc.BuildingId_AirDefense:
		return GetBuildingCfg("AirDefense", level)
	case rpc.BuildingId_Mortar:
		return GetBuildingCfg("Mortar", level)
	case rpc.BuildingId_TeslaTower:
		return GetBuildingCfg("TeslaTower", level)
	case rpc.BuildingId_XBow:
		return GetBuildingCfg("XBow", level)
	case rpc.BuildingId_AllianceCastle:
		return GetBuildingCfg("AllianceCastle", level)
	case rpc.BuildingId_SpellForge:
		return GetBuildingCfg("SpellForge", level)
	case rpc.BuildingId_Deco1:
		return GetBuildingCfg("Deco1", level)
	case rpc.BuildingId_Deco2:
		return GetBuildingCfg("Deco2", level)
	case rpc.BuildingId_Deco3:
		return GetBuildingCfg("Deco3", level)
	case rpc.BuildingId_Deco4:
		return GetBuildingCfg("Deco4", level)
	case rpc.BuildingId_Deco5:
		return GetBuildingCfg("Deco5", level)
	case rpc.BuildingId_Deco6:
		return GetBuildingCfg("Deco6", level)
	case rpc.BuildingId_Deco7:
		return GetBuildingCfg("Deco7", level)
	case rpc.BuildingId_Deco8:
		return GetBuildingCfg("Deco8", level)
	case rpc.BuildingId_Deco9:
		return GetBuildingCfg("Deco9", level)
	case rpc.BuildingId_Deco10:
		return GetBuildingCfg("Deco10", level)
	case rpc.BuildingId_Deco11:
		return GetBuildingCfg("Deco11", level)
	case rpc.BuildingId_Deco12:
		return GetBuildingCfg("Deco12", level)
	case rpc.BuildingId_Deco13:
		return GetBuildingCfg("Deco13", level)
	case rpc.BuildingId_Deco14:
		return GetBuildingCfg("Deco14", level)
	case rpc.BuildingId_Deco15:
		return GetBuildingCfg("Deco15", level)
	case rpc.BuildingId_Deco16:
		return GetBuildingCfg("Deco16", level)
	case rpc.BuildingId_Deco17:
		return GetBuildingCfg("Deco17", level)
	case rpc.BuildingId_Deco18:
		return GetBuildingCfg("Deco18", level)
	case rpc.BuildingId_Deco19:
		return GetBuildingCfg("Deco19", level)
	case rpc.BuildingId_Deco20:
		return GetBuildingCfg("Deco20", level)
	case rpc.BuildingId_Deco21:
		return GetBuildingCfg("Deco21", level)
	case rpc.BuildingId_Deco22:
		return GetBuildingCfg("Deco22", level)
	case rpc.BuildingId_Deco23:
		return GetBuildingCfg("Deco23", level)
	case rpc.BuildingId_Deco24:
		return GetBuildingCfg("Deco24", level)
	case rpc.BuildingId_Bomb:
		return GetBuildingCfg("Bomb", level)
	case rpc.BuildingId_GiantBomb:
		return GetBuildingCfg("GiantBomb", level)
	case rpc.BuildingId_Eject:
		return GetBuildingCfg("Eject", level)
	case rpc.BuildingId_GeneralHouse:
		return GetBuildingCfg("GeneralHouse", level)
	case rpc.BuildingId_Barrier1:
		return GetBuildingCfg("Barrier1", level)
	case rpc.BuildingId_Barrier2:
		return GetBuildingCfg("Barrier2", level)
	case rpc.BuildingId_Barrier3:
		return GetBuildingCfg("Barrier3", level)
	case rpc.BuildingId_Barrier4:
		return GetBuildingCfg("Barrier4", level)
	case rpc.BuildingId_Barrier5:
		return GetBuildingCfg("Barrier5", level)
	case rpc.BuildingId_Barrier6:
		return GetBuildingCfg("Barrier6", level)
	}
	return nil
}

func GetAttackCost(center_level uint32) uint32 {
	cfg, exist := townhallLevelsCfg[center_level]
	if !exist {
		return 0
	}

	return (*cfg)[0].AttackCost
}

func GetGlocalChatCost() uint32 {
	cfg, exist := globalCfg[strings.ToLower("GLOBAL_CHAT_COST")]
	if !exist {
		return 0
	}

	return (*cfg)[0].Value
}

func GetBuildingCntLimitByTypeId(typeId rpc.BuildingId_IdType, center_level uint32) uint32 {
	cfg, exist := townhallLevelsCfg[center_level]
	if !exist {
		return 0
	}

	switch typeId {
	case rpc.BuildingId_Barrack:
		return (*cfg)[0].Barrack
	case rpc.BuildingId_Center:
	case rpc.BuildingId_Farm:
		return (*cfg)[0].Farm
	case rpc.BuildingId_Laboratory:
		return (*cfg)[0].Laboratory
	case rpc.BuildingId_Wall:
		return (*cfg)[0].Walls
	case rpc.BuildingId_Worker:
		return (*cfg)[0].Worker
	case rpc.BuildingId_FoodStorage:
		return (*cfg)[0].FoodStorage
	case rpc.BuildingId_GoldMine:
		return (*cfg)[0].GoldMine
	case rpc.BuildingId_GoldStorage:
		return (*cfg)[0].GoldStorage
	case rpc.BuildingId_TroopHousing:
		return (*cfg)[0].TroopHousing
	case rpc.BuildingId_ArcherTower:
		return (*cfg)[0].ArcherTower
	case rpc.BuildingId_Cannon:
		return (*cfg)[0].Cannon
	case rpc.BuildingId_WizardTower:
		return (*cfg)[0].WizardTower
	case rpc.BuildingId_AirDefense:
		return (*cfg)[0].AirDefense
	case rpc.BuildingId_Mortar:
		return (*cfg)[0].Mortar
	case rpc.BuildingId_TeslaTower:
		return (*cfg)[0].TeslaTower
	case rpc.BuildingId_XBow:
		return (*cfg)[0].XBow
	case rpc.BuildingId_AllianceCastle:
		return (*cfg)[0].AllianceCastle
	case rpc.BuildingId_SpellForge:
		return (*cfg)[0].SpellForge
	case rpc.BuildingId_Deco1:
		return (*cfg)[0].Deco1
	case rpc.BuildingId_Deco2:
		return (*cfg)[0].Deco2
	case rpc.BuildingId_Deco3:
		return (*cfg)[0].Deco3
	case rpc.BuildingId_Deco4:
		return (*cfg)[0].Deco4
	case rpc.BuildingId_Deco5:
		return (*cfg)[0].Deco5
	case rpc.BuildingId_Deco6:
		return (*cfg)[0].Deco6
	case rpc.BuildingId_Deco7:
		return (*cfg)[0].Deco7
	case rpc.BuildingId_Deco8:
		return (*cfg)[0].Deco8
	case rpc.BuildingId_Deco9:
		return (*cfg)[0].Deco9
	case rpc.BuildingId_Deco10:
		return (*cfg)[0].Deco10
	case rpc.BuildingId_Deco11:
		return (*cfg)[0].Deco11
	case rpc.BuildingId_Deco12:
		return (*cfg)[0].Deco12
	case rpc.BuildingId_Deco13:
		return (*cfg)[0].Deco13
	case rpc.BuildingId_Deco14:
		return (*cfg)[0].Deco14
	case rpc.BuildingId_Deco15:
		return (*cfg)[0].Deco15
	case rpc.BuildingId_Deco16:
		return (*cfg)[0].Deco16
	case rpc.BuildingId_Deco17:
		return (*cfg)[0].Deco17
	case rpc.BuildingId_Deco18:
		return (*cfg)[0].Deco18
	case rpc.BuildingId_Deco19:
		return (*cfg)[0].Deco19
	case rpc.BuildingId_Deco20:
		return (*cfg)[0].Deco20
	case rpc.BuildingId_Deco21:
		return (*cfg)[0].Deco21
	case rpc.BuildingId_Deco22:
		return (*cfg)[0].Deco22
	case rpc.BuildingId_Deco23:
		return (*cfg)[0].Deco23
	case rpc.BuildingId_Deco24:
		return (*cfg)[0].Deco24
	case rpc.BuildingId_Bomb:
		return (*cfg)[0].Bomb
	case rpc.BuildingId_GiantBomb:
		return (*cfg)[0].GiantBomb
	case rpc.BuildingId_Eject:
		return (*cfg)[0].Eject
	case rpc.BuildingId_GeneralHouse:
		return (*cfg)[0].GeneralHouse

	}
	return 0
}

func GetCharacterCfgByTypeId(chType rpc.CharacterType, level uint32) *CharacterCfg {
	if level == 0 {
		return nil
	}

	switch chType {
	case rpc.CharacterType_Barbarian:
		return GetCharacterCfg("Barbarian", level)
	case rpc.CharacterType_Archer:
		return GetCharacterCfg("Archer", level)
	case rpc.CharacterType_Goblin:
		return GetCharacterCfg("Goblin", level)
	case rpc.CharacterType_Giant:
		return GetCharacterCfg("Giant", level)
	case rpc.CharacterType_WallBreaker:
		return GetCharacterCfg("WallBreaker", level)
	case rpc.CharacterType_Balloon:
		return GetCharacterCfg("Balloon", level)
	case rpc.CharacterType_Wizard:
		return GetCharacterCfg("Wizard", level)
	case rpc.CharacterType_Healer:
		return GetCharacterCfg("Healer", level)
	case rpc.CharacterType_Dragon:
		return GetCharacterCfg("Dragon", level)
	case rpc.CharacterType_PEKKA:
		return GetCharacterCfg("PEKKA", level)
	case rpc.CharacterType_Yuanfang:
		return GetCharacterCfg("Yuanfang", level)
	case rpc.CharacterType_Lvbu:
		return GetCharacterCfg("Lvbu", level)
	case rpc.CharacterType_Diaochan:
		return GetCharacterCfg("Diaochan", level)
	case rpc.CharacterType_Guanyu:
		return GetCharacterCfg("Guanyu", level)
	case rpc.CharacterType_Yuanxiuqi:
		return GetCharacterCfg("Yuanxiuqi", level)
	}

	return nil
}

func GetSpellCfgByTypeId(spType rpc.SpellType, level uint32) *SpellCfg {
	if level == 0 {
		return nil
	}

	switch spType {
	case rpc.SpellType_Haste:
		return GetSpellCfg("Haste", level)
	case rpc.SpellType_Jump:
		return GetSpellCfg("Jump", level)
	case rpc.SpellType_Xmas:
		return GetSpellCfg("xmas", level)
	case rpc.SpellType_HealingWave:
		return GetSpellCfg("HealingWave", level)
	case rpc.SpellType_LighningStorm:
		return GetSpellCfg("LighningStorm", level)
	}

	return nil
}
func GetBuffCfgByTypeId(spType rpc.TTTBuff_TTTBuffType) *TttBuffCfg {
	if spType == 0 {
		return nil
	}

	switch spType {
	case rpc.TTTBuff_TTTBuffAddBattleTime:
		return GetTTTBuffCfg("TTTBuffAddBattleTime")
	case rpc.TTTBuff_TTTBuffAddArmy:
		return GetTTTBuffCfg("TTTBuffAddArmy")
	case rpc.TTTBuff_TTTBuffAddSpell:
		return GetTTTBuffCfg("TTTBuffAddSpell")
	case rpc.TTTBuff_TTTBuffJumpCheckPoint:
		return GetTTTBuffCfg("TTTBuffJumpCheckPoint")
	case rpc.TTTBuff_TTTBuffDetectEye:
		return GetTTTBuffCfg("TTTBuffDetectEye")
	case rpc.TTTBuff_TTTBuffExpandInitilialCount:
		return GetTTTBuffCfg("TTTBuffExpandInitilialCount")
	case rpc.TTTBuff_TTTBuffExpandAlifeCount:
		return GetTTTBuffCfg("TTTBuffExpandAlifeCount")
	}

	return nil
}
func GetSpellCfg(cfg string, level uint32) *SpellCfg {
	if level == 0 {
		return nil
	}

	if cfgs, exist := spellCfg[strings.ToLower(cfg)]; exist {
		if level > uint32(len(*cfgs)) {
			return nil
		}

		return &(*cfgs)[level-1]
	}

	return nil
}

func GetCharacterCfg(cfg string, level uint32) *CharacterCfg {
	if level == 0 {
		return nil
	}

	if cfgs, exist := charactorCfg[strings.ToLower(cfg)]; exist {
		if level > uint32(len(*cfgs)) {
			return nil
		}

		return &(*cfgs)[level-1]
	}

	return nil
}

func GetBuildingCfg(cfg string, level uint32) *BuildingCfg {
	if level == 0 {
		return nil
	}

	if cfgs, exist := buildingCfg[strings.ToLower(cfg)]; exist {
		if level > uint32(len(*cfgs)) {
			return nil
		}

		return &(*cfgs)[level-1]
	}

	return nil
}

func GetTaskCfg(key string) *TaskCfg {
	cfg, exist := taskCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}
func GetTTTCfg(key string) *TttCfg {
	cfg, exist := tttCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}
func GetTTTBuffCfg(key string) *TttBuffCfg {
	cfg, exist := tttbuffCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}
func GetPassHour(time_unix uint32) float64 {
	return (float64(time.Now().Unix()) - float64(time_unix)) / 3600
}

func GetExpPoints(level uint32) uint32 {
	cfg, exist := expCfg[level]
	if !exist {
		return 0
	}

	return (*cfg)[0].ExpPoints
}

func GetGlobalCfg(key string) uint32 {
	cfg, exist := globalCfg[strings.ToLower(key)]
	if !exist {
		return 0
	}

	return (*cfg)[0].Value
}

func GetShareAwardCfg(key uint32) *ShareAwardCfg {
	cfg, exist := shareawardCfg[key]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

func GetLandAwardCfg(key uint32) *LandAwardCfg {
	cfg, exist := landawardCfg[key]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

//pve配置
func GetPVEStageCfg(stageid uint32) *PVEStageCfg {
	key := strconv.FormatUint(uint64(stageid), 10)
	if cfg, exist := gPVEStageCfg[key]; exist {
		return &(*cfg)[0]
	}

	return nil
}
*/
