package common

import (
	"database/sql"
	"fmt"
	"logger"
	"pools"
	"time"

	_ "github.com/code.google.com/p/go-mysql-driver/mysql"
)

type DbPool struct {
	*pools.RoundRobin
}

type CreateDBFunc func() (*sql.DB, string, error)

func NewDBPool(cfg MySQLConfig) (pool *DbPool) {
	pool = &DbPool{pools.NewRoundRobin(int(cfg.PoolSize), time.Duration(cfg.IdleTimeOut*1e9))}
	pool.Open(DBCreator(cfg))
	return pool
}

func (self *DbPool) Open(dbFactory CreateDBFunc) {
	if dbFactory == nil {
		return
	}
	f := func() (pools.Resource, error) {
		db, dns, err := dbFactory()
		if err != nil {
			return nil, err
		}
		return &dbInstance{db, self, false, dns}, nil
	}
	self.RoundRobin.Open(f)
}

func (self *DbPool) Get() *dbInstance {
	r, err := self.RoundRobin.Get()
	if err != nil {
		panic(err)
	}

	return r.(*dbInstance)
}

type dbInstance struct {
	sqldb    *sql.DB
	pool     *DbPool
	isClosed bool
	dns      string
}

func (self *dbInstance) Exec(query string, args ...interface{}) (sql.Result, error) {
	return self.sqldb.Exec(query, args...)
}

func (self *dbInstance) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return self.sqldb.Query(query, args...)
}

func (self *dbInstance) Close() {
	self.sqldb.Close()
	self.isClosed = true
}

func (self *dbInstance) IsClosed() bool {
	return self.isClosed
}

func (self *dbInstance) Recycle() {
	self.pool.Put(self)
}

func DBCreator(cfg MySQLConfig) CreateDBFunc {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", cfg.Uname, cfg.Pass, cfg.Host, cfg.Port, cfg.Dbname, cfg.Charset)
	logger.Info("MySqlDNS: %s", dns)
	return func() (db *sql.DB, dnsInfo string, err error) {
		dnsInfo = dns
		var retry uint8 = 0

		for {
			db, err = sql.Open("mysql", dns)
			if err == nil {
				break
			}

			logger.Error("Error on Create db: %s; try: %d/%d", err.Error(), retry, cfg.MaxRetry)

			if retry >= cfg.MaxRetry {
				return
			}

			retry++
		}

		return
	}
}
