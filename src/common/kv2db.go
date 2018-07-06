package common

import (
	"fmt"
	"logger"
	"rpc/proto"
	"rpcplus"

	//gp "code.google.com/p/goprotobuf/proto"
	"github.com/code.google.com/p/snappy-go/snappy"
	gp "github.com/golang/protobuf/proto"
)

func KVQuery(db *rpcplus.Client, table, uid string, value gp.Message) (exist bool, err error) {
	//ts("KVQuery", table, uid)
	//defer te("KVQuery", table, uid)

	var reply proto.DBQueryResult

	err = db.Call("DBServer.Query", proto.DBQuery{table, uid}, &reply)

	if err != nil {
		logger.Error("KVQuery Error On Query %s : %s (%s)", table, uid, err.Error())
		return
	}

	switch reply.Code {
	case proto.Ok:

		var dst []byte

		dst, err = snappy.Decode(nil, reply.Value)

		if err != nil {
			logger.Error("KVQuery Unmarshal Error On snappy.Decode %s : %s (%s)", table, uid, err.Error())
			return
		}

		err = gp.Unmarshal(dst, value)

		if err != nil {
			logger.Error("KVQuery Unmarshal Error On Query %s : %s (%s)", table, uid, err.Error())
			return
		}

		exist = true
		return

	case proto.NoExist:
		return
	}

	logger.Error("KVQuery Unknow DBReturn %d", reply.Code)

	return false, fmt.Errorf("KVQuery Unknow DBReturn %d", reply.Code)
}

func KVWrite(db *rpcplus.Client, table, uid string, value gp.Message) (result bool, err error) {
	//ts("KVWrite", table, uid)
	//defer te("KVWrite", table, uid)

	buf, err := gp.Marshal(value)

	if err != nil {
		logger.Error("KVWrite Error On Marshal %s : %s (%s)", table, uid, err.Error())
		return
	}

	dst, err := snappy.Encode(nil, buf)

	if err != nil {
		logger.Error("KVWrite Error On snappy.Encode %s : %s (%s)", table, uid, err.Error())
		return
	}

	var reply proto.DBWriteResult
	err = db.Call("DBServer.Write", proto.DBWrite{table, uid, dst}, &reply)

	if err != nil {
		logger.Error("KVWrite Error On Create %s: %s (%s)", table, uid, err.Error())
		return
	}

	if reply.Code != proto.Ok {
		logger.Error("KVWrite Error On Create %s: %s code (%d)", table, uid, reply.Code)
		return
	}

	result = true
	return
}

func KVDelete(db *rpcplus.Client, table, uid string) (result bool, err error) {
	//ts("KVDelete", table, uid)
	//defer te("KVDelete", table, uid)

	var reply proto.DBDelResult
	err = db.Call("DBServer.Delete", proto.DBDel{table, uid}, &reply)

	if err != nil {
		logger.Error("KVDelete Error On %s: %s (%s)", table, uid, err.Error())
		return
	}

	if reply.Code != proto.Ok {
		logger.Error("KVDelete Error On %s: %s code (%d)", table, uid, reply.Code)
		return
	}

	result = true
	return
}
