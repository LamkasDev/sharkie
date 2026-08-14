package av_player

import "sync"

var GlobalAvPlayerEngine *AvPlayerEngine

type AvPlayerEngine struct {
	Handles    map[uint32]*AvPlayerHandle
	NextHandle uint32
	Lock       sync.RWMutex
}

func NewAvPlayerEngine() *AvPlayerEngine {
	return &AvPlayerEngine{
		Handles:    map[uint32]*AvPlayerHandle{},
		NextHandle: 0x1001,
		Lock:       sync.RWMutex{},
	}
}

func (ape *AvPlayerEngine) CreateHandle() *AvPlayerHandle {
	ape.Lock.Lock()
	defer ape.Lock.Unlock()
	handle := &AvPlayerHandle{
		Id: ape.NextHandle,
	}
	ape.Handles[handle.Id] = handle
	ape.NextHandle++

	return handle
}

func (ape *AvPlayerEngine) GetHandle(id uint32) *AvPlayerHandle {
	ape.Lock.RLock()
	defer ape.Lock.RUnlock()
	return ape.Handles[id]
}

func SetupAvPlayerEngine() {
	GlobalAvPlayerEngine = NewAvPlayerEngine()
}
