package structs

import "sync"

var (
	EventFlagRepo   = map[int32]*EventFlag{}
	EventFlagLock   sync.RWMutex
	EventFlagNextId = int32(1)
)

const (
	EVF_ATTR_TH_DEFAULT = 0x00
	EVF_ATTR_TH_FIFO    = 0x01
	EVF_ATTR_TH_PRIO    = 0x02
	EVF_ATTR_SINGLE     = 0x10
	EVF_ATTR_MULTI      = 0x20

	EVF_NAME_MAX = 32
)

type EventFlag struct {
	Id             int32
	Name           string
	Attributes     uint32
	CurrentPattern uint64
	InitialPattern uint64

	Mutex sync.Mutex
}

func AddEventFlag(ef *EventFlag) int32 {
	EventFlagLock.Lock()
	defer EventFlagLock.Unlock()

	id := EventFlagNextId
	ef.Id = id
	EventFlagRepo[id] = ef
	EventFlagNextId++
	return id
}

func GetEventFlag(id int32) *EventFlag {
	EventFlagLock.RLock()
	defer EventFlagLock.RUnlock()
	return EventFlagRepo[id]
}
