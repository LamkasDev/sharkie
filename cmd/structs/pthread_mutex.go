package structs

import "github.com/gookit/color"

type PthreadMutexType uint32
type PthreadMutexProtocol uint32

const (
	PthreadMutexTypeErrorCheck = PthreadMutexType(1)
	PthreadMutexTypeRecursive  = PthreadMutexType(2)
	PthreadMutexTypeNormal     = PthreadMutexType(3)
	PthreadMutexTypeAdaptiveNp = PthreadMutexType(4)
	PthreadMutexTypeMask       = 0xFF
)

const (
	PthreadMutexProtocolNone    = PthreadMutexProtocol(0)
	PthreadMutexProtocolInherit = PthreadMutexProtocol(1)
	PthreadMutexProtocolProtect = PthreadMutexProtocol(2)
)

const (
	ThrMutexInitializer         = 0
	ThrAdaptiveMutexInitializer = 1
	ThrMutexDestroyed           = 2
)

type PthreadMutex struct {
	Lock       uintptr // TODO: TimedMutex
	Flags      uint32
	_          uint32  // Padding.
	Owner      uintptr // TODO: *Pthread
	Count      int32
	SpinLoops  int32
	YieldLoops int32
	Protocol   PthreadMutexProtocol
	_          [20]byte // More padding yay!
	NamedObjId uint32
	NamePtr    uintptr
}

type PthreadMutexAttr struct {
	Type     PthreadMutexType
	Protocol PthreadMutexProtocol
	Ceiling  int32
}

func GetMutexNameText(m *PthreadMutex, addr uintptr) string {
	if m.NamePtr == 0 {
		return color.Yellow.Sprintf("0x%X", addr)
	}

	name := ReadCString(m.NamePtr)
	return color.Blue.Sprint(name)
}
