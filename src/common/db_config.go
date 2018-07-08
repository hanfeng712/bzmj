package common

import (
	"jscfg"
	"logger"
	//	"os"
	"path"
)

type MySQLConfig struct {
	Host        string  `json:"host"`
	Port        uint16  `json:"port"`
	Uname       string  `json:"uname"`
	Pass        string  `json:"pass"`
	Vnode       uint8   `json:"vnode"`
	NodeName    string  `json:"nodename"`
	Dbname      string  `json:"dbname"`
	Charset     string  `json:"charset"`
	PoolSize    uint16  `json:"pool"`
	IdleTimeOut float64 `json:"idle"`
	MaxRetry    uint8   `json:"retry"`
}

type CacheConfig struct {
	Host        string  `json:"host"`
	Port        uint16  `json:"port"`
	Index       uint8   `json:"index"`
	Vnode       uint8   `json:"vnode"`
	NodeName    string  `json:"nodename"`
	PoolSize    uint16  `json:"pool"`
	IdleTimeOut float64 `json:"idle"`
	MaxRetry    uint8   `json:"retry"`
}

type TableConfig struct {
	DBProfile    string `json:"db-profile"`
	CacheProfile string `json:"cache-profile"`
	DeleteExpiry uint64 `json:"expiry"`
}

type DBConfig struct {
	DBHost        string
	DebugHost     string
	GcTime        uint8
	DBProfiles    map[string][]MySQLConfig `json:"database"`
	CacheProfiles map[string][]CacheConfig `json:"cache"`
	Tables        map[string]TableConfig   `json:"tables"`
}

//读取配置表
func ReadDbConfig(file string, cfg *DBConfig) error {
	//cfgpath, _ := os.Getwd()
	cfgpath := "/home/hanfeng/golang/src/bzmj/bin"
	if err := jscfg.ReadJson(path.Join(cfgpath, file), cfg); err != nil {
		logger.Fatal("read center config failed, %v", err)
		return err
	}

	return nil
}
