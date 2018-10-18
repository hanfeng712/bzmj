package common

import (
	"fmt"
	"logger"
	"rpc"
)

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	err := conn.WriteObj(value)
	if err != nil {
		logger.Info("WriteResult Error %s", err.Error())
		return false
	}
	return true
}

func WriteClientResult(conn rpc.RpcConn, serviceMethod string, value interface{}) bool {
	err := conn.Call(serviceMethod, value)
	if err != nil {
		logger.Info("WriteResult Error %s", err.Error())
		return false
	}
	return true
}

func SyncError(conn rpc.RpcConn, format string, args ...interface{}) {
	//tArgs := make([]interface{}, len(args))
	//for i, arg := range args {
	//	tArgs[i] = arg
	//}

	msg := rpc.SyncError{}
	value := fmt.Sprintf(format, args...)
	msg.Text = &value
	//msg.SetText(fmt.Sprintf(format, args...))

	WriteResult(conn, &msg)

	logger.Error(format, args...)
}

func SendMsg(conn rpc.RpcConn, code string) {
	/*
		msg := rpc.Msg{}
		msg.Code = &code

		WriteResult(conn, &msg)
	*/
}

func SendText(conn rpc.RpcConn, text string) {
	/*
		msg := rpc.Msg{}
		msg.Text = &text
		//msg.SetText(text)

		WriteResult(conn, &msg)
	*/
}
