package structs

import "sync"

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
	Lock         sync.Mutex
	Cond         *sync.Cond
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
		Lock:         sync.Mutex{},
	}
	semaphore.Cond = sync.NewCond(&semaphore.Lock)
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
	CreateSemaphore("SceLncSuspendBlock00000000", 0, 0, 255)
	CreateSemaphore("SceNpTpip-1", 0, 0, 255)
}
