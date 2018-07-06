package common

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"logger"
	"pools"
	"time"
)

type CachePool struct {
	*pools.RoundRobin
}

type CreateCacheFunc func() (redis.Conn, error)

func NewCachePool(cfg CacheConfig) (pool *CachePool) {
	pool = &CachePool{pools.NewRoundRobin(int(cfg.PoolSize), time.Duration(cfg.IdleTimeOut*1e9))}
	pool.Open(CacheCreator(cfg))
	tmp := make([]*cacheInstance, cfg.PoolSize)
	for i := uint16(0); i < cfg.PoolSize; i++ {
		tmp[i] = pool.Get()
	}
	for i := uint16(0); i < cfg.PoolSize; i++ {
		tmp[i].Recycle()
	}
	return pool
}

func (self *CachePool) Open(cacheFactory CreateCacheFunc) {
	if cacheFactory == nil {
		return
	}
	f := func() (pools.Resource, error) {
		c, err := cacheFactory()
		if err != nil {
			return nil, err
		}
		return &cacheInstance{c, self, false}, nil
	}
	self.RoundRobin.Open(f)
}

func (self *CachePool) Get() *cacheInstance {
	r, err := self.RoundRobin.Get()
	if err != nil {
		panic(err)
	}

	return r.(*cacheInstance)
}

type cacheInstance struct {
	redis.Conn
	pool     *CachePool
	isClosed bool
}

func (self *cacheInstance) Close() {
	self.Conn.Close()
	self.isClosed = true
}

func (self *cacheInstance) IsClosed() bool {
	return self.isClosed
}

func (self *cacheInstance) Recycle() {
	self.pool.Put(self)
}

func CacheCreator(cfg CacheConfig) CreateCacheFunc {
	dns := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return func() (c redis.Conn, err error) {
		var retry uint8 = 0

		for {
			c, err = redis.Dial("tcp", dns)
			if err == nil {
				break
			}

			logger.Error("Error on Create redis: %s; try: %d/%d", err.Error(), retry, cfg.MaxRetry)

			if retry >= cfg.MaxRetry {
				return
			}

			retry++
		}

		_, err = c.Do("SELECT", cfg.Index)

		return
	}
}
