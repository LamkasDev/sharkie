package cond

import (
	"sync"
)

var (
	// CondRepo maps guest addresses (uintptr) to host conds (*CondWaitable).
	CondRepo = map[uintptr]*CondWaitable{}

	// CondLock protects CondRepo, so multiple threads can look up conds safely.
	CondLock sync.RWMutex
)

// GetCond retrieves or creates CondWaitable corresponding to a guest address.
func GetCond(guestAddress uintptr) *CondWaitable {
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

	cond = NewCondWaitable()
	CondRepo[guestAddress] = cond
	return cond
}
