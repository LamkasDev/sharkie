package cond

import (
	"sync"
	"time"
)

type CondWaitable struct {
	Mutex sync.Mutex
	wait  chan struct{}
}

func NewCondWaitable() *CondWaitable {
	return &CondWaitable{
		wait: make(chan struct{}),
	}
}

func (cond *CondWaitable) Broadcast() {
	cond.Mutex.Lock()
	defer cond.Mutex.Unlock()

	close(cond.wait)
	cond.wait = make(chan struct{})
}

func (cond *CondWaitable) Signal() {
	cond.Mutex.Lock()
	defer cond.Mutex.Unlock()

	select {
	case cond.wait <- struct{}{}:
	default:
	}
}

func (cond *CondWaitable) Wait() {
	cond.Mutex.Lock()
	w := cond.wait
	cond.Mutex.Unlock()

	<-w
}

func (cond *CondWaitable) WaitTimeout(timeout time.Duration) bool {
	cond.Mutex.Lock()
	w := cond.wait
	cond.Mutex.Unlock()

	select {
	case <-w:
		return true
	case <-time.After(timeout):
		return false
	}
}
