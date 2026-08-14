package np_web_api

import (
	"sync"
)

var GlobalWebApiManager *WebApiManager

type WebApiManager struct {
	Contexts      map[uint32]*WebApiContext
	NextContextId uint32
	Lock          sync.RWMutex
}

func NewWebApiManager() *WebApiManager {
	return &WebApiManager{
		Contexts:      map[uint32]*WebApiContext{},
		NextContextId: 0x1001,
		Lock:          sync.RWMutex{},
	}
}

func (wam *WebApiManager) CreateContext() *WebApiContext {
	wam.Lock.Lock()
	defer wam.Lock.Unlock()
	context := &WebApiContext{
		Id: wam.NextContextId,
	}
	wam.Contexts[context.Id] = context
	wam.NextContextId++

	return context
}

func (wam *WebApiManager) GetContext(id uint32) *WebApiContext {
	wam.Lock.RLock()
	defer wam.Lock.RUnlock()
	return wam.Contexts[id]
}

func SetupWebApiManager() {
	GlobalWebApiManager = NewWebApiManager()
}
