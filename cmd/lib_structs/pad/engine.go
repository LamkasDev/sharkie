// Package pad contains structs to emulate controller devices.
package pad

import "sync"

var GlobalPadEngine *PadEngine

// PadEngine keeps state of controller devices.
type PadEngine struct {
	Handles    map[uint32]*PadHandle
	NextHandle uint32
	Lock       sync.RWMutex
}

func NewPadEngine() *PadEngine {
	return &PadEngine{
		Handles:    map[uint32]*PadHandle{},
		NextHandle: 0x20000001,
		Lock:       sync.RWMutex{},
	}
}

func (pe *PadEngine) CreateHandle() *PadHandle {
	pe.Lock.Lock()
	defer pe.Lock.Unlock()
	handle := &PadHandle{
		Id: pe.NextHandle,
	}
	pe.Handles[handle.Id] = handle
	pe.NextHandle++

	return handle
}

func (pe *PadEngine) GetHandle(id uint32) *PadHandle {
	pe.Lock.RLock()
	defer pe.Lock.RUnlock()
	return pe.Handles[id]
}

func SetupPadEngine() {
	GlobalPadEngine = NewPadEngine()
}
