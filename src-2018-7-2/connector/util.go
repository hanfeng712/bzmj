package connector

import (
	"common"
	"logger"
	"math/rand"
	"sync/atomic"
	"time"
)

func GenLockMessage(sid uint8, tid uint8, value uint8) uint64 {
	return common.GenLockMessage(sid, tid, value)
}

func ParseLockMessage(lid uint64) (sid uint8, tid uint8, value uint8, t uint32, tmpid uint8) {
	return common.ParseLockMessage(lid)
}

var pid uint32 = 0

func GenPlayerId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&pid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(sid)<<56
}

var vid uint32 = 0

func GenVillageId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&vid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(vid)<<56
}

var bid uint32 = 0

func GenBattleId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&vid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(vid)<<56
}

var rid uint32 = 0

func GenReplayId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&vid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(vid)<<56
}

// UUID() provides unique identifier strings.
func GenUUID(sid uint8) string {
	return common.GenUUID(sid)
}

func CheckUUID(uid string) bool {
	return common.CheckUUID(uid)
}

const DEBUG = true

func dbgf(format string, items ...interface{}) {
	if DEBUG {
		logger.Info(format, items...)
	}
}

const TRACE = true

func ts(name string, items ...interface{}) {
	if TRACE {
		logger.Info("+%s %v\n", name, items)
	}
}
func te(name string, items ...interface{}) {
	if TRACE {
		logger.Info("-%s %v\n", name, items)
	}
}

func RandomNumber(start uint32, stop uint32) uint32 {
	if start > stop {
		start, stop = stop, start
	}

	//前闭后开
	total := stop - start + 1

	//同一时刻调用多次返回值一样
	//var randSource rand.Source = rand.NewSource(time.Now().Unix())
	//ran := rand.New(randSource)

	return uint32(rand.Intn(int(total))) + start
}

func RandomWeightTable(table map[interface{}]uint32) interface{} {
	tRandomTable := make(map[interface{}][2]uint32, 0)
	nSum := uint32(0)
	for k, v := range table {
		tRandomTable[k] = [2]uint32{nSum, nSum + v - 1}
		nSum += v
	}

	nRet := RandomNumber(0, nSum-1)

	for k, v := range tRandomTable {
		if nRet >= v[0] && nRet <= v[1] {
			return k
		}
	}

	return nil
}
