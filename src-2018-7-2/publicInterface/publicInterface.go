package publicInterface

type CnInterface interface {
	SendMsgToPlayer(msg interface{}, uid uint64, method string)
}
