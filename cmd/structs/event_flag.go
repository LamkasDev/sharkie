package structs

import (
	"fmt"
	"sync"

	"github.com/gookit/color"
)

var (
	// EventFlagRepo maps handles to host event flags (*EventFlag).
	EventFlagRepo = map[uintptr]*EventFlag{}

	// EventFlagLock protects EventFlagRepo, so multiple threads can look up event flags safely.
	EventFlagLock sync.RWMutex

	NextEventFlagId = uintptr(1)
)

const (
	EVF_ATTR_TH_FIFO = 0x01
	EVF_ATTR_TH_PRIO = 0x02
	EVF_ATTR_SINGLE  = 0x10
	EVF_ATTR_MULTI   = 0x20

	EVF_WAITMODE_AND       = 0x01
	EVF_WAITMODE_OR        = 0x02
	EVF_WAITMODE_CLEAR_ALL = 0x10
	EVF_WAITMODE_CLEAR_PAT = 0x20

	EVF_NAME_MAX = 32
)

type EventFlag struct {
	Handle         uintptr
	Name           string
	Attributes     uint32
	CurrentPattern uint64
	InitialPattern uint64

	Lock sync.Mutex
	Cond *sync.Cond
}

func NewEventFlag(name string, attributes uint32, currentPattern, initialPattern uint64) *EventFlag {
	eventFlag := &EventFlag{
		Handle:         NextEventFlagId,
		Name:           name,
		Attributes:     attributes,
		CurrentPattern: currentPattern,
		InitialPattern: initialPattern,
		Lock:           sync.Mutex{},
	}
	eventFlag.Cond = sync.NewCond(&eventFlag.Lock)

	return eventFlag
}

func CreateEventFlag(name string, attributes uint32, currentPattern, initialPattern uint64) *EventFlag {
	EventFlagLock.Lock()
	defer EventFlagLock.Unlock()

	eventFlag := NewEventFlag(name, attributes, currentPattern, initialPattern)
	EventFlagRepo[eventFlag.Handle] = eventFlag
	NextEventFlagId++
	return eventFlag
}

func GetEventFlag(handle uintptr) *EventFlag {
	EventFlagLock.RLock()
	defer EventFlagLock.RUnlock()
	return EventFlagRepo[handle]
}

func CheckEventFlagCondition(current, wait uint64, mode uint32) bool {
	if (mode & EVF_WAITMODE_OR) != 0 {
		return (current & wait) != 0
	}

	return (current & wait) == wait
}

func GetEventFlagName(eventFlag *EventFlag) string {
	if eventFlag.Name == "" {
		return color.Yellow.Sprintf("0x%X", eventFlag.Handle)
	}

	return color.Blue.Sprint(eventFlag.Name)
}

func CreateDefaultEventFlags(names []string) {
	for _, name := range names {
		CreateEventFlag(name, EVF_ATTR_TH_FIFO|EVF_ATTR_MULTI, 0, 0)
	}
}

func SetupEventFlags() {
	CreateDefaultEventFlags([]string{AudioInEventFlagName})
	CreateDefaultEventFlags([]string{
		"SceBootStatusFlags",
		"SceSystemStateMgrInfo",
		"SceSystemStateMgrStatus",
		fmt.Sprintf("SceNpTusIpc_%08x", 1001),
		fmt.Sprintf("SceNpScoreIpc_%08x", 1001),
	})
}
