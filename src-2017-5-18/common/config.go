package common

type RoomConfig struct {
	RoomId          uint64
	MinMatchUserNum int
}

func LoadRoomConfig() *RoomConfig {
	var lRoomCfg RoomConfig

	lRoomCfg.RoomId = 0
	lRoomCfg.MinMatchUserNum = 4
	return &lRoomCfg
}
