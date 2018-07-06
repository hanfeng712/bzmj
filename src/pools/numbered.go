package pools

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Numbered 允许你通过id来管理资源，被管理的资源不需要实现任何接口
type Numbered struct {
	mu        sync.Mutex
	empty     *sync.Cond // 当池子变成空的时候广播
	resources map[int64]*numberedWrapper
}

type numberedWrapper struct {
	val         interface{}
	inUse       bool
	timeCreated time.Time
}

func NewNumbered() *Numbered {
	n := &Numbered{resources: make(map[int64]*numberedWrapper)}
	n.empty = sync.NewCond(&n.mu)
	return n
}

// 通过id注册资源，不会锁住对象，如果id已存在会返回一个错误
func (self *Numbered) Register(id int64, val interface{}) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if _, ok := self.resources[id]; ok {
		return errors.New("already present")
	}
	self.resources[id] = &numberedWrapper{val, false, time.Now()}
	return nil
}

// 反注册指定资源，如果资源不存在则忽略
func (self *Numbered) Unregister(id int64) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.resources, id)
	if len(self.resources) == 0 {
		self.empty.Broadcast()
	}
}

// 获取并锁住资源。如果没有发现或者正在被使用则返回错误
func (self *Numbered) Get(id int64) (val interface{}, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	nw, ok := self.resources[id]
	if !ok {
		return nil, errors.New("not found")
	}
	if nw.inUse {
		return nil, errors.New("in use")
	}
	nw.inUse = true
	return nw.val, nil
}

// 解锁资源
func (self *Numbered) Put(id int64) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if nw, ok := self.resources[id]; ok {
		nw.inUse = false
	}
}

// 获取超时的资源， 并且锁住它们。 （已被占用的资源不会被返回）
func (self *Numbered) GetTimedout(timeout time.Duration) (vals []interface{}) {
	self.mu.Lock()
	defer self.mu.Unlock()
	now := time.Now()
	for _, nw := range self.resources {
		if nw.inUse {
			continue
		}
		if nw.timeCreated.Add(timeout).Sub(now) <= 0 {
			nw.inUse = true
			vals = append(vals, nw.val)
		}
	}
	return vals
}

// 当池子变空之后， 返回
func (self *Numbered) WaitForEmpty() {
	self.mu.Lock()
	defer self.mu.Unlock()
	for len(self.resources) != 0 {
		self.empty.Wait()
	}
}

func (self *Numbered) StatsJSON() string {
	s := self.Stats()
	return fmt.Sprintf("{\"Size\": %v}", s)
}

func (self *Numbered) Stats() (size int) {
	self.mu.Lock()
	defer self.mu.Unlock()
	return len(self.resources)
}
