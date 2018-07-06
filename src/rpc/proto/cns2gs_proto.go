package proto

type SendCnsInfo struct {
	ServerId    uint16
	PlayerCount uint16
	ServerIp    string
}

type SendCnsInfoResult struct {
	SendResult uint8
}
