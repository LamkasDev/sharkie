package structs

import "sync"

var (
	// MutexRepo maps guest addresses (uintptr) to host mutexes (*sync.Mutex).
	MutexRepo = map[uintptr]*sync.Mutex{}

	// MutexLock protects MutexRepo, so multiple threads can look up locks safely.
	MutexLock sync.RWMutex
)

// GetMutex retrieves or creates Go sync.Mutex corresponding to a guest address.
func GetMutex(guestAddress uintptr) *sync.Mutex {
	// This doesn't need to be fast, let's just read/write in one pass.
	MutexLock.Lock()
	defer MutexLock.Unlock()
	if mutex, ok := MutexRepo[guestAddress]; ok {
		return mutex
	}

	mutex := &sync.Mutex{}
	MutexRepo[guestAddress] = mutex

	return mutex
}
