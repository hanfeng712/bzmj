package logserver

import (
	"database/sql"
	"fmt"
	"net"
	"rpcplus"

	_ "github.com/code.google.com/p/go-mysql-driver/mysql"
	//"time"
	"log"
	"os"
	//	"rpc/proto"
)

type MysqlDb struct {
	ConnectInfo string
	Db          *sql.DB
	StmtMap     map[string]*sql.Stmt
	bTrans      bool
	SqlMap      map[string]string
}

func (self *MysqlDb) Exec(stmt string, args ...interface{}) (sql.Result, error) {
	pStmt := self.StmtMap[stmt]
	result, err := pStmt.Exec(args...)

	if err != nil {
		fmt.Printf("stmt % execute failed : %s \n", err)
	}
	// 这里需要根据错误类型来重新执行
	return result, err
}

func InitMysqlServices(info string, trans bool) *MysqlDb {
	ms := &MysqlDb{ConnectInfo: info,
		Db:      nil,
		StmtMap: make(map[string]*sql.Stmt),
		SqlMap:  make(map[string]string),
		bTrans:  trans}

	ms.SqlMap = sqlMap

	Db, err := sql.Open("mysql", ms.ConnectInfo)

	if err != nil {
		fmt.Printf("can not connect to mysql %s \n", ms.ConnectInfo)
		return nil
	}

	Db.SetMaxOpenConns(100)

	fmt.Printf("connect mysql success!! %s \n", ms.ConnectInfo)

	for key, stmtString := range ms.SqlMap {
		stmt, err := Db.Prepare(stmtString)

		if err != nil {
			ms.StmtMap = nil
			fmt.Printf("Prepare failed : %s \n", err)
			Db.Close()
			return nil
		}
		ms.StmtMap[key] = stmt
	}

	ms.Db = Db
	return ms
}

func (self *MysqlDb) ReInitMysqlServices() error {
	Db, err := sql.Open("mysql", self.ConnectInfo)

	if err != nil {
		return nil
	}

	for key, stmtString := range self.SqlMap {
		stmt, err := Db.Prepare(stmtString)

		if err != nil {
			self.StmtMap = nil
			return nil
		}
		self.StmtMap[key] = stmt
	}

	self.Db = Db
	return nil
}

type LogServices struct {
	mysqlInfo string
	db        *MysqlDb
}

func (self *LogServices) getDb() *MysqlDb {
	return self.db
}

var pLogServices *LogServices

func CreateServices(listener net.Listener, host string, port uint16, user string, pass string, database string, charset string) *LogServices {

	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", user, pass, host, port, database, charset)
	db := InitMysqlServices(dns, false)

	if nil == db {
		os.Exit(0)
	}

	pLogServices = &LogServices{mysqlInfo: dns, db: db}

	self := pLogServices
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pLogServices)

	for {
		newConn, err := listener.Accept()

		if err != nil {
			log.Println("StartServices %s \n", err.Error())
			break
		}

		//开始对cns的RPC服务
		go func(conn net.Conn) {
			rpcServer.ServeConn(conn)
		}(newConn)
	}

	return self
}

/*
func (self *LogServices) LogPlayerLoginLogoutGame(msg *proto.LogPlayerLoginLogout, result *proto.LogPlayerLoginLogoutResult) error {
	result = &proto.LogPlayerLoginLogoutResult{}

	var Info string
	if msg.Logout {
		Info = fmt.Sprintf("%s;", "1")

	} else {
		Info = fmt.Sprintf("%s;%s", "0", msg.Ip)
	}
	self.getDb().Exec("PlayerLoginLogoutGame", msg.ChannelId, msg.Playerid, msg.Time, Info)

	return nil
}

func (self *LogServices) LogResources(msg *proto.LogResources, result *proto.LogResourcesResult) error {
	result = &proto.LogResourcesResult{}

	if msg.Gain {
		self.getDb().Exec("GainResources", msg.ChannelId, msg.Uid, msg.Time, msg.ResType, msg.ResNum, msg.ResWay)
	} else {
		self.getDb().Exec("LoseResources", msg.ChannelId, msg.Uid, msg.Time, msg.ResType, msg.ResNum, msg.ResWay)
	}
	return nil
}
*/
