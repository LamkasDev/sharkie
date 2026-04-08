package structs

import (
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/structs/cond"
)

const PSemaphoreMagic = 0x736D

var (
	// PSemaphoreRepo maps guest addresses (uintptr) to host semaphores (*CondWaitable).
	PSemaphoreRepo = map[uintptr]*cond.CondWaitable{}

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

// GetPSemaphore retrieves or creates CondWaitable corresponding to a guest address.
func GetPSemaphore(guestAddress uintptr) *cond.CondWaitable {
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

	semaphore = cond.NewCondWaitable()
	PSemaphoreRepo[guestAddress] = semaphore
	return semaphore
}
