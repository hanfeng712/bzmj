package main

import (
	"common"
	db "dbserver"
	"flag"
	"logger"
	"net"
)

var (
	//laddr = flag.String("l", "127.0.0.1:8800", "The address to bind to.")
	//dbg_addr     = flag.String("d", "127.0.0.1:8801", "The address to bind to.(for debug)")
	dbConfigFile = flag.String("c", "", "config file name for the dbserver")
)

var dbServer *db.DBServer

func main() {
	logger.Info("dbsserver start")
	flag.Parse()

	var dbcfg common.DBConfig
	if err := common.ReadDbConfig(*dbConfigFile, &dbcfg); err != nil {
		logger.Fatal("load config failed, error is: %v", err)
		return
	}

	common.DebugInit(dbcfg.GcTime, dbcfg.DebugHost)

	dbServer = db.NewDBServer(dbcfg)

	tsock, err := net.Listen("tcp", dbcfg.DBHost)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	logger.Info("dbsserver Listen %s ", dbcfg.DBHost)

	//dbg_sock, err := net.Listen("tcp", *dbg_addr)

	//if err != nil {
	//	logger.Fatal("net.Listen: %s", err.Error())
	//}

	//go http.Serve(dbg_sock, nil)

	go db.StartServices(dbServer, tsock)

	db.WaitForExit(dbServer)

	tsock.Close()

	logger.Info("dbsserver end")
}
