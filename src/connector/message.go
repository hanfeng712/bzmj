package connector

import (
	"common"
	"rpc"
	//	"time"
)

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	return common.WriteResult(conn, value)
}

/*
func WriteLoginResult(conn rpc.RpcConn, r rpc.LoginResult_Result) bool {
	rep := rpc.LoginResult{}
	rep.Result = &r
	rep.ServerTime = NewUint(uint32(time.Now().Unix()))
	return WriteResult(conn, &rep)
}

func WriteLoginResultWithErrorMsg(conn rpc.RpcConn, r rpc.LoginResult_Result, msg string) bool {
	rep := rpc.LoginResult{}
	rep.Result = &r
	rep.SetErrmsg(msg)
	rep.ServerTime = NewUint(uint32(time.Now().Unix()))
	return WriteResult(conn, &rep)
}

func WriteMatchResult(conn rpc.RpcConn, r rpc.MatchPlayerResult_Result) bool {
	rep := rpc.MatchPlayerResult{}
	rep.Result = &r
	return WriteResult(conn, &rep)
}
*/

func SyncError(conn rpc.RpcConn, format string, args ...interface{}) {
	common.SyncError(conn, format, args...)
}

func SendMsg(conn rpc.RpcConn, code string) {
	common.SendMsg(conn, code)
}

func SendText(conn rpc.RpcConn, text string) {
	common.SendText(conn, text)
}
