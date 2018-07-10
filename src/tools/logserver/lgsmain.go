package main

import (
	"common"
	"logger"
	"logserver"
	"net"
	"os"
	"syscall"
)

var lgsConfig common.LogServerCfg

func main() {

	if err := common.ReadLogConfig(&lgsConfig); err != nil {
		logger.Fatal("can't find lgscfg.json: %v", err)
		return
	}
	common.DebugInit(lgsConfig.GcTime, lgsConfig.DebugHost)

	quitChan := make(chan int)

	listener, err := net.Listen("tcp", lgsConfig.LogHost)
	defer listener.Close()
	if err != nil {
		logger.Fatal("Listening to: ", lgsConfig.LogHost, " failed !!")
		return
	}
	logger.Info("Listening to: ", lgsConfig.LogHost, "Success !!")

	go logserver.CreateServices(listener, lgsConfig.Host,
		lgsConfig.Port,
		lgsConfig.User,
		lgsConfig.Pass,
		lgsConfig.Dbname,
		lgsConfig.Charset)

	handler := func(s os.Signal, arg interface{}) {
		logger.Info("handle signal: %v\n", s)
		logger.Info("logserver close")
		os.Exit(0)
	}

	handlerArray := []os.Signal{syscall.SIGINT,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGSEGV,
		syscall.SIGTERM,
		syscall.SIGABRT}

	common.WatchSystemSignal(&handlerArray, handler)

	nQuitCount := 0
	for {
		select {
		case <-quitChan:
			nQuitCount = nQuitCount + 1
		}

		if nQuitCount == 2 {
			break
		}
	}

	logger.Info("logserver close")
}
