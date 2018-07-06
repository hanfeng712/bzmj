package common

import (
	"crypto/rc4"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	gp "github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
)

func EncodeMessage(value gp.Message) (result []byte, err error) {
	//ts("KVWrite", table, uid)
	//defer te("KVWrite", table, uid)

	buf, err := gp.Marshal(value)

	if err != nil {
		return
	}

	result = snappy.Encode(nil, buf)

	return
}

func DecodeMessage(value []byte, result gp.Message) (err error) {
	var dst []byte

	dst, err = snappy.Decode(nil, value)

	if err != nil {
		return
	}

	err = gp.Unmarshal(dst, result)

	return
}

//唯一id生成
var uuid uint32 = 0

// UUID() provides unique identifier strings.
func GenUUid() uint64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	lUid := r.Int63n(600000) + 600000
	return uint64(lUid)
}
func GenUUID(sid uint8) string {
	b := make([]byte, 16)

	t := time.Now().Unix()
	tmpid := uint16(atomic.AddUint32(&uuid, 1))

	b[0] = byte(sid)
	b[1] = byte(0)
	b[2] = byte(tmpid)
	b[3] = byte(tmpid >> 8)

	b[4] = byte(t)
	b[5] = byte(t >> 8)
	b[6] = byte(t >> 16)
	b[7] = byte(t >> 24)

	c, _ := rc4.NewCipher([]byte{0x0c, b[2], b[3], b[0]})
	c.XORKeyStream(b[8:], b[:8])

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:12], b[12:])
}

func CheckUUID(uid string) bool {
	if len(uid) != 36 {
		return false
	}

	b := make([]uint32, 5)

	_, err := fmt.Sscanf(uid, "%x-%x-%x-%x-%x", &b[0], &b[1], &b[2], &b[3], &b[4])
	if err != nil {
		return false
	}

	info1 := make([]byte, 4)
	binary.BigEndian.PutUint32(info1, b[0])

	info2 := make([]byte, 4)
	binary.BigEndian.PutUint16(info2[:2], uint16(b[1]))
	binary.BigEndian.PutUint16(info2[2:], uint16(b[2]))

	c, _ := rc4.NewCipher([]byte{0x0c, info1[2], info1[3], info1[0]})

	tmp := make([]byte, 4)

	c.XORKeyStream(tmp, info1)

	if binary.BigEndian.Uint32(tmp) != b[3] {
		return false
	}

	c.XORKeyStream(tmp, info2)

	if binary.BigEndian.Uint32(tmp) != b[4] {
		return false
	}

	return true
}

func GenMailId() string {
	return "mail-" + GenUUID(0)
}

//是否是同一天
func IsTheSameDay(utime1, utime2 uint32) bool {
	time1 := time.Unix(int64(utime1), 0)
	time2 := time.Unix(int64(utime2), 0)

	return time1.YearDay() == time2.YearDay() && time1.Year() == time2.Year()
}

//是否是同一周
func IsTheSameWeek(utime1, utime2 uint32) bool {
	time1 := time.Unix(int64(utime1), 0)
	time2 := time.Unix(int64(utime2), 0)
	year1, week1 := time1.ISOWeek()
	year2, week2 := time2.ISOWeek()
	return year1 == year2 && week1 == week2
}

//锁定相关
var nid uint32 = 0

func GenLockMessage(sid uint8, tid uint8, value uint8) uint64 {

	tmpid := uint8(atomic.AddUint32(&nid, 1))

	return uint64(time.Now().Unix()) | uint64(tmpid)<<32 | uint64(value)<<40 | uint64(tid)<<48 | uint64(sid)<<56
}

func ParseLockMessage(lid uint64) (sid uint8, tid uint8, value uint8, t uint32, tmpid uint8) {
	return uint8(lid >> 56), uint8(lid >> 48), uint8(lid >> 40), uint32(lid), uint8(lid >> 32)
}
