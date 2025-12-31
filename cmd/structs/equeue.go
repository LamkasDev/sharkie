package structs

import "sync"

var (
	// EqueueRepo maps handles to host equeues (*Equeue).
	EqueueRepo = map[uintptr]*Equeue{}

	// EqueueLock protects EqueueRepo, so multiple threads can look up equeues safely.
	EqueueLock sync.RWMutex

	NextEqueueId = uintptr(1)
)

type Equeue struct {
	Handle     uintptr
	Name       string
	Events     chan Kevent
	UserEvents map[uintptr]bool
	Lock       sync.Mutex
}

func CreateEqueue(name string) *Equeue {
	EqueueLock.Lock()
	defer EqueueLock.Unlock()

	equeue := &Equeue{
		Handle:     NextEqueueId,
		Name:       name,
		Events:     make(chan Kevent, 256),
		UserEvents: map[uintptr]bool{},
		Lock:       sync.Mutex{},
	}
	EqueueRepo[equeue.Handle] = equeue
	NextEqueueId++
	return equeue
}

func GetEqueue(handle uintptr) *Equeue {
	EqueueLock.RLock()
	defer EqueueLock.RUnlock()
	return EqueueRepo[handle]
}
