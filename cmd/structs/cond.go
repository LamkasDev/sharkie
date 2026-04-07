package structs

import (
	"context"
	"runtime/pprof"
	"sync"
	"time"
)

var (
	// CondRepo maps guest addresses (uintptr) to host conds (*sync.Cond).
	CondRepo = map[uintptr]*sync.Cond{}

	// CondLock protects CondRepo, so multiple threads can look up conds safely.
	CondLock sync.RWMutex

	// GlobalCondMutex is a global mutex for all condition variables.
	GlobalCondMutex = &sync.Mutex{}
)

// GetCond retrieves or creates Go sync.Cond corresponding to a guest address.
func GetCond(guestAddress uintptr) *sync.Cond {
	CondLock.RLock()
	cond, ok := CondRepo[guestAddress]
	CondLock.RUnlock()
	if ok {
		return cond
	}

	// Create new cond.
	CondLock.Lock()
	defer CondLock.Unlock()
	if cond, ok = CondRepo[guestAddress]; ok {
		return cond
	}

	cond = sync.NewCond(GlobalCondMutex)
	CondRepo[guestAddress] = cond
	return cond
}

func CondWaitTimeout(cond *sync.Cond, timeout time.Duration) bool {
	done := make(chan struct{})
	go pprof.Do(context.Background(), pprof.Labels("name", "CondWaitTimeout"), func(ctx context.Context) {
		cond.Wait()
		cond.L.Unlock()
		close(done)
	})

	select {
	case <-done:
		cond.L.Lock()
		return true
	case <-time.After(timeout):
		cond.L.Lock()
		return false
	}
}
