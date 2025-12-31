package structs

import (
	"sync"
	"unsafe"
)

var (
	// PSemaphoreRepo maps guest addresses (uintptr) to host semaphores (*sync.Cond).
	PSemaphoreRepo = map[uintptr]*sync.Cond{}

	// PSemaphoreLock protects PSemaphoreRepo, so multiple threads can look up locks safely.
	PSemaphoreLock sync.RWMutex
)

type PSemaphore struct {
	Magic       uint16
	Flags       uint16
	WaitAddress uint32
	Value       int32
	Pshared     int32
	_           [16]byte
}

const PSemaphoreSize = unsafe.Sizeof(PSemaphore{})

// GetPSemaphore retrieves or creates Go sync.Cond corresponding to a guest address.
func GetPSemaphore(guestAddress uintptr) *sync.Cond {
	PSemaphoreLock.RLock()
	semaphore, ok := PSemaphoreRepo[guestAddress]
	PSemaphoreLock.RUnlock()
	if ok {
		return semaphore
	}

	// Create new semaphore.
	PSemaphoreLock.Lock()
	defer PSemaphoreLock.Unlock()
	if semaphore, ok = PSemaphoreRepo[guestAddress]; ok {
		return semaphore
	}

	semaphore = sync.NewCond(&sync.Mutex{})
	PSemaphoreRepo[guestAddress] = semaphore
	return semaphore
}
