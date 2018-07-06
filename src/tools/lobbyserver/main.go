package main

import (
	"common"
	"encoding/json"
	"lobbyserver"
	//	"log"
	"net"
	"os"
	"path"
	//"strconv"
	"fmt"
	"syscall"
)

var ipcfg common.LobbyServerCfg

func main() {

	cfgpath, _ := os.Getwd()
	cfg, err := os.Open(path.Join(cfgpath, "gscfg.json"))

	if err != nil {
		println("can't find gscfg.json")
		return
	}

	defer cfg.Close()

	deccfg := json.NewDecoder(cfg)

	if err := deccfg.Decode(&ipcfg); err != nil {
		println("can't find gscfg.json")
		return
	}

	common.DebugInit(ipcfg.GcTime, ipcfg.DebugHost)

	quitChan := make(chan int)

	listenerForClient, err := net.Listen("tcp", ipcfg.GsIpForClient)
	defer listenerForClient.Close()
	if err != nil {
		println("Listening to: ", ipcfg.GsIpForClient, " failed !!")
		return
	}
	//println("Listening to: ", listenerForClient.Addr().String())

	listenerForServer, err := net.Listen("tcp", ipcfg.GsIpForServer)
	defer listenerForServer.Close()
	if err != nil {
		println("Listening to: ", listenerForServer.Addr().String())
		return
	}
	//println("Listening to: ", listenerForServer.Addr().String())

	go lobbyserver.CreateLobbyServicesForCnserver(listenerForServer)
	go lobbyserver.CreateLobbyServicesForClient(listenerForClient)

	handler := func(s os.Signal, arg interface{}) {
		fmt.Printf("handle signal: %v\n", s)
		println("gateserver close")
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

		//println("nQuitCount = %s", strconv.Itoa(nQuitCount))
		if nQuitCount == 2 {
			break
		}
	}

	println("gateserver close")

}
