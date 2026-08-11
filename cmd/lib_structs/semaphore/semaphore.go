package semaphore

import (
	"fmt"
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/cond"
)

var (
	// SemaphoreRepo maps handles to host semaphores (*Semaphore).
	SemaphoreRepo = map[uint32]*Semaphore{}

	// SemaphoreLock protects SemaphoreRepo, so multiple threads can look up semaphores safely.
	SemaphoreLock sync.RWMutex

	NextSemaphoreId = uint32(1)
)

type Semaphore struct {
	Handle           uint32
	Name             string
	Attributes       uint32
	InitCount        int32
	CurrentCount     int32
	MaxCount         int32
	CancelGeneration int32

	Cond *CondWaitable
}

func CreateSemaphore(name string, attributes uint32, currentCount, maxCount int32) *Semaphore {
	SemaphoreLock.Lock()
	defer SemaphoreLock.Unlock()

	semaphore := &Semaphore{
		Handle:       NextSemaphoreId,
		Name:         name,
		Attributes:   attributes,
		InitCount:    currentCount,
		CurrentCount: currentCount,
		MaxCount:     maxCount,
		Cond:         NewCondWaitable(),
	}
	SemaphoreRepo[semaphore.Handle] = semaphore
	NextSemaphoreId++
	return semaphore
}

func DeleteSemaphore(handle uint32) {
	SemaphoreLock.Lock()
	defer SemaphoreLock.Unlock()
	delete(SemaphoreRepo, handle)
}

func GetSemaphore(handle uint32) *Semaphore {
	SemaphoreLock.RLock()
	defer SemaphoreLock.RUnlock()
	return SemaphoreRepo[handle]
}

func SetupSemaphores() {
	CreateSemaphore(fmt.Sprintf("SceLncSuspendBlock%08x", 1), 0, 0, 255)
	CreateSemaphore("SceNpTpip-1", 0, 0, 255)
}
