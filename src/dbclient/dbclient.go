package dbclient

import (
	"common"
	"logger"
	"rpcplusclientpool"

	gp "github.com/golang/protobuf/proto"
)

var pPollBase *rpcplusclientpool.ClientPool
var pPollExtern *rpcplusclientpool.ClientPool

func Init() {
	//base
	var dbCfg common.DBConfig
	if err := common.ReadDbConfig("dbBase.json", &dbCfg); err != nil {
		logger.Fatal("%v", err)
	}

	aHosts := make([]string, 0)
	aHosts = append(aHosts, dbCfg.DBHost)
	pPollBase = rpcplusclientpool.CreateClientPool(aHosts)
	if pPollBase == nil {
		logger.Fatal("create failed")
	}

	//extern
	/*
		if err := common.ReadDbConfig("dbExtern.json", &dbCfg); err != nil {
			logger.Fatal("%v", err)
		}

		aHosts = make([]string, 0)
		aHosts = append(aHosts, dbCfg.DBHost)
		pPollExtern = rpcplusclientpool.CreateClientPool(aHosts)
		if pPollExtern == nil {
			logger.Fatal("create failed")
		}
	*/
}

//基础信息库
func KVQueryBase(table, uid string, value gp.Message) (exist bool, err error) {
	err, conn := pPollBase.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVQuery(conn, table, uid, value)
}

func KVWriteBase(table, uid string, value gp.Message) (result bool, err error) {
	err, conn := pPollBase.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVWrite(conn, table, uid, value)
}

func KVDeleteBase(table, uid string) (exist bool, err error) {
	err, conn := pPollBase.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVDelete(conn, table, uid)
}

//额外信息库
func KVQueryExt(table, uid string, value gp.Message) (exist bool, err error) {
	err, conn := pPollExtern.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVQuery(conn, table, uid, value)
}

func KVWriteExt(table, uid string, value gp.Message) (result bool, err error) {
	err, conn := pPollExtern.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVWrite(conn, table, uid, value)
}

func KVDeleteExt(table, uid string) (exist bool, err error) {
	err, conn := pPollExtern.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVDelete(conn, table, uid)
}
