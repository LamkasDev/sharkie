package np_trophy

import (
	"sync"
)

var GlobalTrophyManager *TrophyManager

type TrophyManager struct {
	Contexts      map[uint32]*TrophyContext
	NextContextId uint32
	Handles       map[uint32]*TrophyHandle
	NextHandleId  uint32
	Lock          sync.RWMutex
}

func NewTrophyManager() *TrophyManager {
	return &TrophyManager{
		Contexts:      map[uint32]*TrophyContext{},
		NextContextId: 0x1001,
		Handles:       map[uint32]*TrophyHandle{},
		NextHandleId:  0x1001,
		Lock:          sync.RWMutex{},
	}
}

func (tm *TrophyManager) CreateContext() *TrophyContext {
	tm.Lock.Lock()
	defer tm.Lock.Unlock()
	context := &TrophyContext{
		Id: tm.NextContextId,
	}
	tm.Contexts[context.Id] = context
	tm.NextContextId++

	return context
}

func (tm *TrophyManager) GetContext(id uint32) *TrophyContext {
	tm.Lock.RLock()
	defer tm.Lock.RUnlock()
	return tm.Contexts[id]
}

func (tm *TrophyManager) CreateHandle() *TrophyHandle {
	tm.Lock.Lock()
	defer tm.Lock.Unlock()
	handle := &TrophyHandle{
		Id: tm.NextHandleId,
	}
	tm.Handles[handle.Id] = handle
	tm.NextHandleId++

	return handle
}

func (tm *TrophyManager) GetHandle(id uint32) *TrophyHandle {
	tm.Lock.RLock()
	defer tm.Lock.RUnlock()
	return tm.Handles[id]
}

func SetupTrophyManager() {
	GlobalTrophyManager = NewTrophyManager()
}
