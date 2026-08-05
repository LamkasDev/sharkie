package rwlock

import "sync"

// GuestRWLock implements the custom RWLock mechanism
type GuestRWLock struct {
	Mu    sync.RWMutex
	Owner uintptr
}

func NewGuestRWLock() *GuestRWLock {
	return &GuestRWLock{}
}

var (
	// RwlockRepo maps guest addresses (uintptr) to host rwlocks (*GuestRWLock).
	RwlockRepo = map[uintptr]*GuestRWLock{}

	// RwlockLock protects RwlockRepo, so multiple threads can look up locks safely.
	RwlockLock sync.RWMutex
)

// GetRwlock retrieves or creates Go GuestRWLock corresponding to a guest address.
func GetRwlock(guestAddress uintptr) *GuestRWLock {
	RwlockLock.RLock()
	rwlock, ok := RwlockRepo[guestAddress]
	RwlockLock.RUnlock()
	if ok {
		return rwlock
	}

	// Create new rwlock.
	RwlockLock.Lock()
	defer RwlockLock.Unlock()
	if rwlock, ok = RwlockRepo[guestAddress]; ok {
		return rwlock
	}

	rwlock = NewGuestRWLock()
	RwlockRepo[guestAddress] = rwlock
	return rwlock
}
