package common

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"
	"time"
)

func DebugInit(gcTime uint8, debugDns string) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	go func() {
		var m runtime.MemStats
		for {
			//HeapSys：程序向应用程序申请的内存
			//HeapAlloc：堆上目前分配的内存
			//HeapIdle：堆上目前没有使用的内存
			//HeapReleased：回收到操作系统的内存
			runtime.GC()
			debug.FreeOSMemory()
			runtime.ReadMemStats(&m)
			log.Println("Gc : HeapSys, HeapAlloc, HeapIdle, HeapReleased", m.HeapSys, m.HeapAlloc,
				m.HeapIdle, m.HeapReleased)
			time.Sleep(time.Second * time.Duration(gcTime))
		}
	}()

	go func() {
		log.Println("Debug Http Service :", debugDns)
		log.Println(http.ListenAndServe(debugDns, nil))
	}()
}
