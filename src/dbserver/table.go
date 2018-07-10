package dbserver

import (
	"common"
	"database/sql"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"hash/crc32"
	"io"
	"logger"
	"stats"
	"time"
)

const (
	keylen = 64
)

type table struct {
	name         string
	caches       cacheGroup
	dbs          dbGroup
	deleteExpiry uint64
	tableStats   *stats.Timings
	qpsRates     *stats.Rates
	cacheNode    []uint32
	dbNode       []uint32
	dbVirNode    map[uint32]uint32
	cacheVirNode map[uint32]uint32
}

func NewTable(name string, cfg common.TableConfig, db *DBServer) (t *table) {
	var (
		caches       cacheGroup
		cacheNode    []uint32
		dbs          dbGroup
		dbNode       []uint32
		dbVirNode    map[uint32]uint32
		cacheVirNode map[uint32]uint32
	)
	if cfg.CacheProfile != "" {
		var exist bool
		if caches, exist = db.cacheGroups[cfg.CacheProfile]; !exist {
			logger.Fatal("NewTable: table cache profile not found: %s", cfg.CacheProfile)
		}
		cacheNode, _ = db.cacheNodes[cfg.CacheProfile]
		cacheVirNode, _ = db.cacheVirNodes[cfg.CacheProfile]
	}

	if cfg.DBProfile != "" {
		var exist bool
		if dbs, exist = db.dbGroups[cfg.DBProfile]; !exist {
			logger.Fatal("NewTable: table db profile not found: %s", cfg.DBProfile)
		}
		dbNode, _ = db.dbNodes[cfg.DBProfile]
		dbVirNode, _ = db.dbVirNodes[cfg.DBProfile]
		for _, dbpool := range dbs {
			db := dbpool.Get()
			defer db.Recycle()
			/*
				query := fmt.Sprintf(`
						CREATE TABLE IF NOT EXISTS %s (
					    added_id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
					    id BINARY(32) NOT NULL,
					    body MEDIUMBLOB,
					    updated TIMESTAMP NOT NULL,
					    UNIQUE KEY (id),
					    KEY (updated)
					) ENGINE=InnoDB;
					`, name)
			*/

			query := fmt.Sprintf(`
					CREATE TABLE IF NOT EXISTS %s (
				    id BINARY(64) NOT NULL PRIMARY KEY,
					auto_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				    body MEDIUMBLOB,
				    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				    KEY (updated),
					key (auto_id)
				) ENGINE=InnoDB;
				`, name)

			logger.Info("CreateQuery :%s", query)
			rst, err := db.Exec(
				query,
			)

			if err != nil {
				logger.Fatal("NewTable: db %v create table %s faild! %s", dbpool, name, err.Error())
			}

			logger.Info("NewTable: db %v init %s: %v", dbpool, name, rst)

		}
	}

	if caches == nil && dbs == nil {
		logger.Fatal("NewTable: table %s need a save func", name)
	}

	queryStats := stats.NewTimings("")
	qpsRates := stats.NewRates("", queryStats, 20, 10e9)
	return &table{
		name, caches, dbs,
		cfg.DeleteExpiry,
		queryStats,
		qpsRates,
		cacheNode,
		dbNode,
		dbVirNode,
		cacheVirNode,
	}
}

func (self *table) write(key string, value []byte) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	defer self.tableStats.Record("write", time.Now())
	hid := makeHash(key)

	if self.caches != nil {
		cidx := self.getCacheNode(hid)
		cache := cidx.Get()
		defer cache.Recycle()

		_, err = cache.Conn.Do("SET", self.name+"_"+key, value)
		//fmt.Print("SET ", self.name+"_"+key, " : ",value, "\n")

		if err != nil {
			logger.Fatal("write error: %s (%s, %v)", err.Error(), key, value)
		}
	}

	if self.dbs != nil {
		didx := self.getDbNode(hid)
		db := didx.Get()
		defer db.Recycle()

		_, err = db.Exec("INSERT INTO "+self.name+" (id, body) values(?, ?) ON DUPLICATE KEY UPDATE body=?;", key, value, value)
		//_, err = db.Exec("REPLACE INTO "+self.name+" (id, body) values(?, ?);", key, value)

		if err != nil {
			logger.Fatal("write error: %s (%s, %v)", err.Error(), key, value)
		}
	}

	return
}

func (self *table) get(key string) (ret []byte, err error) {
	if len(key) > keylen {
		return nil, fmt.Errorf("key (%s) len must <= 64", key)
	}

	defer self.tableStats.Record("get", time.Now())
	hid := makeHash(key)

	if self.caches != nil {
		cidx := self.getCacheNode(hid)
		cache := cidx.Get()
		defer cache.Recycle()

		ret, err = redis.Bytes(cache.Conn.Do("GET", self.name+"_"+key))
		//fmt.Print("GET ", self.name+"_"+key, "\n")
		if err != nil {

			if err != redis.ErrNil {
				logger.Fatal("get error: %s (%s, %v)", err.Error(), key, ret)
			} else {
				err = nil
			}
		}

		if ret != nil {
			return
		}
	}

	if self.dbs != nil {
		didx := self.getDbNode(hid)
		db := didx.Get()
		defer db.Recycle()

		var rows *sql.Rows
		rows, err = db.Query("SELECT body from "+self.name+" where id = CAST(? as BINARY(64)) LIMIT 1;", key)

		if err != nil {
			logger.Fatal("get error: %s (%s, %v)", err.Error(), key, rows)
		}

		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&ret)
			if err != nil {
				logger.Fatal("get scan error %s (%s)", err.Error(), key)
			}
			return
		}
	}

	return nil, nil
}

func (self *table) del(key string) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	defer self.tableStats.Record("del", time.Now())
	hid := makeHash(key)

	if self.caches != nil {
		cidx := self.getCacheNode(hid)
		cache := cidx.Get()
		defer cache.Recycle()

		_, err = cache.Conn.Do("DEL", self.name+"_"+key)

		if err != nil {
			if err != redis.ErrNil {
				logger.Fatal("del error: %s (%s)", err.Error(), key)
			} else {
				err = nil
			}
		}
	}

	if self.dbs != nil {
		didx := self.getDbNode(hid)
		db := didx.Get()
		defer db.Recycle()

		_, err = db.Query("DELETE from "+self.name+" where id = CAST(? as BINARY(64));", key)

		if err != nil {
			logger.Fatal("delete error: %s (%s)", err.Error(), key)
		}
	}

	return
}

func (self *table) getDbNode(key uint32) *common.DbPool {

	var index = 0
	for k, v := range self.dbNode {
		if key < v {
			index = k
			break
		}
	}

	node, ok := self.dbs[self.dbVirNode[self.dbNode[index]]]
	if !ok {
		logger.Fatal("getDbNode error: no find  (%d)", key)
	}
	return node
}

func (self *table) getCacheNode(key uint32) *common.CachePool {

	var index = 0
	for k, v := range self.cacheNode {
		if key < v {
			index = k
			break
		}
	}

	node, ok := self.caches[self.cacheVirNode[self.cacheNode[index]]]
	if !ok {
		logger.Fatal("getCacheNode error: no find (%d)", key)
	}
	return node
}

func makeHash(key string) uint32 {
	ieee := crc32.NewIEEE()
	io.WriteString(ieee, key)
	return ieee.Sum32()
}
