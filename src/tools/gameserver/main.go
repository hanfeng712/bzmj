// bzmj project main.go
package main

import (
	"common"
	"connector"
	"flag"
	"logger"
	"net"
	_ "net/http/pprof"
	"os"
	"sync"
	"syscall"
)

var (
	//	laddr    = flag.String("l", "192.168.8.103:8820", "The address to bind to.")
	//	dbg_addr = flag.String("d", "127.0.0.1:8821", "The address to bind to.(for debug)")
	csvDir = flag.String("c", "config", "config dir")
)

func main() {
	logger.Info("cnserver start")

	flag.Parse()

	if err := common.ReadCnsServerConfig("gas1.json", &connector.Cfg); err != nil {
		logger.Fatal("load cns config error", *csvDir, err)
		return
	}

	common.DebugInit(connector.Cfg.GcTime, connector.Cfg.DebugHost)

	csock, err := net.Listen("tcp", connector.Cfg.CnsForCenter)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	//dbg_sock, err := net.Listen("tcp", *dbg_addr)
	//if err != nil {
	//	logger.Fatal("net.Listen: %s", err.Error())
	//}
	//go http.Serve(dbg_sock, nil)

	cnServer := connector.NewCNServer(&connector.Cfg)

	wg := &sync.WaitGroup{}
	connector.StartCenterService(cnServer, csock, &connector.Cfg)
	cnServer.StartClientService(&connector.Cfg, wg)

	handler := func(s os.Signal, arg interface{}) {
		logger.Info("cnserver handle signal: %d", s)
		cnServer.Quit()
		logger.Info("cnserver will close")
	}

	handlerArray := []os.Signal{syscall.SIGINT,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGSEGV,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGKILL}

	logger.Info("WatchSystemSignal!!!!!!!!!!!")
	go common.WatchSystemSignal(&handlerArray, handler)
	logger.Info("wait client Quit!!!!!!!!!!!")
	wg.Wait()
	logger.Info("all client Quit!!!!!!!!!!!")
	cnServer.EndService()
	logger.Info("cnserver end")
}

/*
package main

//protoc --go_out=. *.proto
import (
	"connector"
	"fmt"
	"sync"
)

func main() {
	fmt.Println("Hello World!")

	wg := &sync.WaitGroup{}
	cnServer := connector.NewCNServer()
	cnServer.StartClientService(1, wg)
	for {
	}
}
*/
