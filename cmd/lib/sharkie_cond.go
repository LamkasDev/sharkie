package lib

import "sync"

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
	// This doesn't need to be fast, let's just read/write in one pass.
	CondLock.Lock()
	defer CondLock.Unlock()
	if cond, ok := CondRepo[guestAddress]; ok {
		return cond
	}

	cond := sync.NewCond(GlobalCondMutex)
	CondRepo[guestAddress] = cond

	return cond
}
