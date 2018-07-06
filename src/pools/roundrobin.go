//	管理可服用资源， 比如连接池。
package pools

import (
	"fmt"
	"sync"
	"time"
)

// 用于创建资源的工厂
type Factory func() (Resource, error)

// 每一个资源必须支持Resource接口。 Close()和IsClosed()之间的调用是同步的
// 由调用者来保证
type Resource interface {
	Close()
	IsClosed() bool
}

// RoundRobin 允许你用轮叫调度（round robin）的方式使用资源池
type RoundRobin struct {
	mu          sync.Mutex
	available   *sync.Cond
	resources   chan fifoWrapper
	size        int64
	factory     Factory
	idleTimeout time.Duration

	// 状态
	waitCount int64
	waitTime  time.Duration
}

type fifoWrapper struct {
	resource Resource
	timeUsed time.Time
}


// NewRoundRobin 创建新的RoundRobin池
// capacity：池内最大资源容量（pool会调用factory创建资源）
// idleTimeout：长时间未被使用的资源，会被释放。
func NewRoundRobin(capacity int, idleTimeout time.Duration) *RoundRobin {
	r := &RoundRobin{
		resources:   make(chan fifoWrapper, capacity),
		size:        0,
		idleTimeout: idleTimeout,
	}
	r.available = sync.NewCond(&r.mu)
	return r
}

// 设置工厂函数，开始允许分配资源
func (self *RoundRobin) Open(factory Factory) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.factory = factory
}

// 清空资源池（会等待所有资源被放回Put到池内）
func (self *RoundRobin) Close() {
	self.mu.Lock()
	defer self.mu.Unlock()
	for self.size > 0 {
		select {
		case fw := <-self.resources:
			go fw.resource.Close()
			self.size--
		default:
			self.available.Wait()
		}
	}
	self.factory = nil
}

func (self *RoundRobin) IsClosed() bool {
	return self.factory == nil
}

// 返回一个可用资源，如果没有空闲的，但是没达到容量上限，则用工厂创建一个。 
// 否则的话，卡起等
func (self *RoundRobin) Get() (resource Resource, err error) {
	return self.get(true)
}

// 返回一个可用资源，如果没有空闲的，但是没达到容量上限，则用工厂创建一个。 
// 否则的话，返回nil
func (self *RoundRobin) TryGet() (resource Resource, err error) {
	return self.get(false)
}

func (self *RoundRobin) get(wait bool) (resource Resource, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	// 此循环里曼的任何等待都会先释放掉锁，然后在返回的时候重新申请
	for {
		select {
		case fw := <-self.resources:
			// 从channel里取出一个空闲资源
			if self.idleTimeout > 0 && fw.timeUsed.Add(self.idleTimeout).Sub(time.Now()) < 0 {
				// 如果空闲太久了， 那么先吧它移除
				go fw.resource.Close()
				self.size--
				// 这时候应该没有人在等待， 但无论如何先发一个信号再说
				self.available.Signal()
				continue
			}
			return fw.resource, nil
		default:
			// 检查池子是否已满
			if self.size >= int64(cap(self.resources)) {
				// 已满的话，等待
				if wait {
					start := time.Now()
					self.available.Wait()
					self.recordWait(start)
					continue
				}
				return nil, nil
			}
			// 未满的话创建一个资源
			if resource, err = self.waitForCreate(); err != nil {
				// 创建如果失败了， 让出位置。其他人可能再等待
				self.available.Signal()
				return nil, err
			}

			// 创建成功， 增加大小
			self.size++
			return resource, err
		}
	}
	panic("unreachable")
}

func (self *RoundRobin) recordWait(start time.Time) {
	self.waitCount++
	self.waitTime += time.Now().Sub(start)
}

func (self *RoundRobin) waitForCreate() (resource Resource, err error) {
	// 预先占一个坑， 在创建完毕后再释放坑。 这样不会导致过渡创建
	self.size++
	self.mu.Unlock()
	defer func() {
		self.mu.Lock()
		self.size--
	}()
	return self.factory()
}

// 将资源放回到池子里面。你必须吧所有资源返回到池子里面， 即使这个资源已经关闭了
// 如果资源已经关闭了， Put会忽略它。 Close和IsClose的调用是线程同步的
func (self *RoundRobin) Put(resource Resource) {
	self.mu.Lock()
	defer self.available.Signal()
	defer self.mu.Unlock()

	if self.size > int64(cap(self.resources)) {
		go resource.Close()
		self.size--
	} else if resource.IsClosed() {
		self.size--
	} else {
		if len(self.resources) == cap(self.resources) {
			panic("unexpected")
		}
		self.resources <- fifoWrapper{resource, time.Now()}
	}
}

// 设置池子容量， 可以扩充和缩减
func (self *RoundRobin) SetCapacity(capacity int) {
	self.mu.Lock()
	defer self.available.Broadcast()
	defer self.mu.Unlock()

	nr := make(chan fifoWrapper, capacity)
	// 将资源从老的channel复制到新的channel
	// 额外的资源会被抛弃。
	for {
		select {
		case fw := <-self.resources:
			if len(nr) < cap(nr) {
				nr <- fw
			} else {
				go fw.resource.Close()
				self.size--
			}
			continue
		default:
		}
		break
	}
	self.resources = nr
}

func (self *RoundRobin) SetIdleTimeout(idleTimeout time.Duration) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.idleTimeout = idleTimeout
}

func (self *RoundRobin) StatsJSON() string {
	s, c, a, wc, wt, it := self.Stats()
	return fmt.Sprintf("{\"Size\": %v, \"Capacity\": %v, \"Available\": %v, \"WaitCount\": %v, \"WaitTime\": %v, \"IdleTimeout\": %v}", s, c, a, wc, int64(wt), int64(it))
}

func (self *RoundRobin) Stats() (size, capacity, available, waitCount int64, waitTime, idleTimeout time.Duration) {
	self.mu.Lock()
	defer self.mu.Unlock()
	return self.size, int64(cap(self.resources)), int64(len(self.resources)), self.waitCount, self.waitTime, self.idleTimeout
}
