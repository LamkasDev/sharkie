package structs

import (
	"fmt"
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/structs/cond"
)

var (
	// SemaphoreRepo maps handles to host semaphores (*Semaphore).
	SemaphoreRepo = map[uintptr]*Semaphore{}

	// SemaphoreLock protects SemaphoreRepo, so multiple threads can look up semaphores safely.
	SemaphoreLock sync.RWMutex

	NextSemaphoreId = uintptr(1)
)

type Semaphore struct {
	Handle       uintptr
	Name         string
	Attributes   uint32
	CurrentCount int32
	MaxCount     int32

	Cond *CondWaitable
}

func CreateSemaphore(name string, attributes uint32, currentCount, maxCount int32) *Semaphore {
	SemaphoreLock.Lock()
	defer SemaphoreLock.Unlock()

	semaphore := &Semaphore{
		Handle:       NextSemaphoreId,
		Name:         name,
		Attributes:   attributes,
		CurrentCount: currentCount,
		MaxCount:     maxCount,
		Cond:         NewCondWaitable(),
	}
	SemaphoreRepo[semaphore.Handle] = semaphore
	NextSemaphoreId++
	return semaphore
}

func DeleteSemaphore(handle uintptr) {
	SemaphoreLock.Lock()
	defer SemaphoreLock.Unlock()
	delete(SemaphoreRepo, handle)
}

func GetSemaphore(handle uintptr) *Semaphore {
	SemaphoreLock.RLock()
	defer SemaphoreLock.RUnlock()
	return SemaphoreRepo[handle]
}

func SetupSemaphores() {
	CreateSemaphore(fmt.Sprintf("SceLncSuspendBlock%08x", GlobalAppInfo.AppId), 0, 0, 255)
	CreateSemaphore("SceNpTpip-1", 0, 0, 255)
}
