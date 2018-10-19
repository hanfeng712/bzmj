package common

type MsgQueueType struct {
	//WebResult map[string]string
	WebResult string
}
type Connect_Type struct {
	WebSocket string
	TcpSocket string
	UdpSocket string
}

var ConnectType *Connect_Type = &Connect_Type{
	WebSocket: "webSocket",
	TcpSocket: "tcpSocket",
	UdpSocket: "udpSocket",
}
