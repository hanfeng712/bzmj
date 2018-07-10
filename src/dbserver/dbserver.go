package dbserver

import (
	"bytes"
	"common"
	"fmt"
	"logger"
	"net"
	"net/http"
	proto "rpc/proto"
	rpc "rpcplus"
)

type dbGroup map[uint32]*common.DbPool
type cacheGroup map[uint32]*common.CachePool

type DBServer struct {
	dbGroups      map[string]dbGroup
	dbNodes       map[string][]uint32
	dbVirNodes    map[string]map[uint32]uint32
	cacheGroups   map[string]cacheGroup
	cacheNodes    map[string][]uint32
	cacheVirNodes map[string]map[uint32]uint32
	tables        map[string]*table
	exit          chan bool
}

func StartServices(self *DBServer, listener net.Listener) {
	rpcServer := rpc.NewServer()
	rpcServer.Register(self)

	rpcServer.HandleHTTP("/dbserver/rpc", "/debug/rpc")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("StartServices %s", err.Error())
			break
		}
		go func() {
			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}
}

func WaitForExit(self *DBServer) {
	<-self.exit
	close(self.exit)
}

func NewDBServer(cfg common.DBConfig) (server *DBServer) {
	server = &DBServer{
		dbGroups:    map[string]dbGroup{},
		cacheGroups: map[string]cacheGroup{},
		tables:      map[string]*table{},
		exit:        make(chan bool),
	}

	http.Handle("/debug/state", debugHTTP{server})

	//初始化所有的db
	for key, pools := range cfg.DBProfiles {
		logger.Info("Init DB Profile %s", key)

		server.dbGroups = make(map[string]dbGroup)
		server.dbVirNodes = make(map[string]map[uint32]uint32)
		server.dbNodes = make(map[string][]uint32)

		temDbs := make(map[uint32]uint32)
		temGroups := make(dbGroup)
		temDbInt := []uint32{}
		for _, poolCfg := range pools {
			logger.Info("Init DB %v", poolCfg)
			leng := makeHash(poolCfg.NodeName)
			temGroups[leng] = common.NewDBPool(poolCfg)
			if poolCfg.Vnode <= 0 {
				poolCfg.Vnode = 1
			}

			var i uint8
			for i = 0; i < poolCfg.Vnode; i++ {
				keys := makeHash(fmt.Sprintf("%s#%d", poolCfg.NodeName, i))
				temDbs[keys] = leng
				temDbInt = append(temDbInt, keys)
			}

		}
		server.dbGroups[key] = temGroups
		bubbleSort(temDbInt) //排序节点
		server.dbVirNodes[key] = temDbs
		server.dbNodes[key] = temDbInt
	}

	//初始化所有的cache
	for key, pools := range cfg.CacheProfiles {
		logger.Info("Init Cache Profile %s", key)

		server.cacheGroups = make(map[string]cacheGroup)
		server.cacheVirNodes = make(map[string]map[uint32]uint32)
		server.cacheNodes = make(map[string][]uint32)

		temDbs := make(map[uint32]uint32)
		temGroups := make(cacheGroup)
		temDbInt := []uint32{}
		for _, poolCfg := range pools {
			logger.Info("Init Cache %v", poolCfg)
			leng := makeHash(poolCfg.NodeName)
			temGroups[leng] = common.NewCachePool(poolCfg)

			if poolCfg.Vnode <= 0 {
				poolCfg.Vnode = 1
			}

			//leng := len(server.cacheGroups[key]) - 1
			var i uint8
			for i = 0; i < poolCfg.Vnode; i++ {
				keys := makeHash(fmt.Sprintf("%s#%d", poolCfg.NodeName, i))
				temDbs[keys] = leng
				temDbInt = append(temDbInt, keys)
			}

		}
		server.cacheGroups[key] = temGroups
		bubbleSort(temDbInt) //排序节点
		server.cacheVirNodes[key] = temDbs
		server.cacheNodes[key] = temDbInt

	}

	//初始化table
	for key, table := range cfg.Tables {
		logger.Info("Init Table: %s %v", key, table)

		server.tables[key] = NewTable(key, table, server)
	}

	return server
}

func (self *DBServer) Write(args *proto.DBWrite, reply *proto.DBWriteResult) error {
	if table, exist := self.tables[args.Table]; exist {
		err := table.write(args.Key, args.Value)
		if err != nil {
			return err
		}
		reply.Code = proto.Ok
	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *DBServer) Query(args *proto.DBQuery, reply *proto.DBQueryResult) error {

	if table, exist := self.tables[args.Table]; exist {
		rst, err := table.get(args.Key)
		if err != nil {
			return err
		}
		if rst != nil {
			reply.Value = rst
			reply.Code = proto.Ok
		} else {
			reply.Code = proto.NoExist
		}

	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *DBServer) Delete(args *proto.DBDel, reply *proto.DBDelResult) error {
	if table, exist := self.tables[args.Table]; exist {
		err := table.del(args.Key)
		if err != nil {
			return err
		}
		reply.Code = proto.Ok
	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *DBServer) Quit(args *int32, reply *int32) error {
	self.exit <- true
	return nil
}

func (self *DBServer) statsJSON() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	fmt.Fprintf(buf, "{")
	for k, v := range self.tables {

		fmt.Fprintf(buf, "\n \"Table\": {")

		fmt.Fprintf(buf, "\n   \"Name\": \"%v\",", k)
		fmt.Fprintf(buf, "\n   \"States\": %v,", v.tableStats.String())
		fmt.Fprintf(buf, "\n   \"Rates\": %v,", v.qpsRates.String())

		fmt.Fprintf(buf, "\n }")
	}

	fmt.Fprintf(buf, "\n}")
	return buf.String()
}

func bubbleSort(values []uint32) {
	flag := true
	for i := 0; i < len(values)-1; i++ {
		flag = true

		for j := 0; j < len(values)-i-1; j++ {
			if values[j] > values[j+1] {
				values[j], values[j+1] = values[j+1], values[j]
				flag = false
			}
		}
		if flag == true {
			break
		}
	}
}
