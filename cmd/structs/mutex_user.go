package structs

import "sync"

var (
	// UserMutexRepo maps guest addresses (uintptr) to host user mutex (*sync.Cond).
	UserMutexRepo = map[uintptr]*sync.Cond{}

	// UserMutexLock protects UserMutexRepo, so multiple threads can look up locks safely.
	UserMutexLock sync.RWMutex
)

// GetUserMutex retrieves or creates Go sync.Cond corresponding to a guest address.
func GetUserMutex(guestAddress uintptr) *sync.Cond {
	UserMutexLock.RLock()
	mutex, ok := UserMutexRepo[guestAddress]
	UserMutexLock.RUnlock()
	if ok {
		return mutex
	}

	// Create new user mutex.
	UserMutexLock.Lock()
	defer UserMutexLock.Unlock()
	if mutex, ok = UserMutexRepo[guestAddress]; ok {
		return mutex
	}

	mutex = sync.NewCond(&sync.Mutex{})
	UserMutexRepo[guestAddress] = mutex
	return mutex
}
