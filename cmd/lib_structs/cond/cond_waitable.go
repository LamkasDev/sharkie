package cond

import (
	"sync"
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
	cond.Broadcast()
}

func (cond *CondWaitable) WaitChan() <-chan struct{} {
	return cond.wait
}
