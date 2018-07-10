package logserver

var sqlMap = map[string]string{
	"PlayerLoginLogoutGame": "insert into tbl_playerlogin set pl_ChannelId = ?, pl_CharId = ?, pl_DateTime = FROM_UNIXTIME(?) , pl_Info = ?;",
	"GainResources":         "insert into log_res_gain set pl_ChannelId = ?, pl_UId = ?, pl_Time = FROM_UNIXTIME(?), pl_ResType = ?, pl_ResNum = ?, pl_ResWay = ?;",
	"LoseResources":         "insert into log_res_lose set pl_ChannelId = ?, pl_UId = ?, pl_Time = FROM_UNIXTIME(?), pl_ResType = ?, pl_ResNum = ?, pl_ResWay = ?;",
	"LogTaobaoPay":          "insert into log_taobao_pay set tp_UId = ?, pl_ChannelId = ?, tp_TradeTime = FROM_UNIXTIME(?), tp_TradeError = ?, tp_TradeEnd = ?, tp_TradeNumber = ?, tp_ItemName = ?, tp_TotoalPee = ?;",
}
