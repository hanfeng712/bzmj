package timer

import (
	"sync"
	"time"
)

type typeAction int

const (
	STOP typeAction = iota
	RESET
	TRIGGER
)

type Timer struct {
	interval time.Duration

	mu      sync.Mutex
	running bool

	msg chan typeAction
}

func NewTimer(interval time.Duration) *Timer {
	return &Timer{
		interval: interval,
		msg:      make(chan typeAction, 1),
	}
}

func (tm *Timer) Start(keephouse func()) {
	tm.mu.Lock()
	if tm.running {
		tm.mu.Unlock()
		return
	}
	tm.running = true
	go tm.run(keephouse)
	tm.mu.Unlock()
}

func (tm *Timer) run(keephouse func()) {
	for {
		var ch <-chan time.Time
		if tm.interval <= 0 {
			ch = nil
		} else {
			ch = time.After(tm.interval)
		}
		select {
		case action := <-tm.msg:
			switch action {
			case STOP:
				return
			case RESET:
				continue
			}
		case <-ch:
		}
		keephouse()
	}
	panic("unreachable")
}

func (tm *Timer) SetInterval(ns time.Duration) {
	tm.interval = ns
	tm.mu.Lock()
	if tm.running {
		tm.msg <- RESET
	}
	tm.mu.Unlock()
}

func (tm *Timer) Trigger() {
	tm.mu.Lock()
	if tm.running {
		tm.msg <- TRIGGER
	}
	tm.mu.Unlock()
}

func (tm *Timer) TriggerAfter(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		tm.Trigger()
	}()
}

func (tm *Timer) Stop() {
	tm.mu.Lock()
	if tm.running {
		tm.msg <- STOP
		tm.running = false
	}
	tm.mu.Unlock()
}
