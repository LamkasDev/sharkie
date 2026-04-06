package pthread

import (
	"unsafe"

	"github.com/gookit/color"
)

type PthreadMutexType uint32
type PthreadMutexProtocol uint32

const (
	PthreadMutexTypeErrorCheck = PthreadMutexType(1)
	PthreadMutexTypeRecursive  = PthreadMutexType(2)
	PthreadMutexTypeNormal     = PthreadMutexType(3)
	PthreadMutexTypeAdaptiveNp = PthreadMutexType(4)
	PthreadMutexTypeMask       = 0xFF
)

var MutexTypeNames = map[PthreadMutexType]string{
	PthreadMutexTypeErrorCheck: "ErrorCheck",
	PthreadMutexTypeRecursive:  "Recursive",
	PthreadMutexTypeNormal:     "Normal",
	PthreadMutexTypeAdaptiveNp: "AdaptiveNp",
	PthreadMutexTypeMask:       "Mask",
}

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
	Name       string
}

const PthreadMutexSize = unsafe.Sizeof(PthreadMutex{})

type PthreadMutexAttr struct {
	Type     PthreadMutexType
	Protocol PthreadMutexProtocol
	Ceiling  int32
}

const PthreadMutexAttrSize = unsafe.Sizeof(PthreadMutexAttr{})

func GetMutexNameText(m *PthreadMutex, addr uintptr) string {
	if true || m.Name == "" {
		return color.Yellow.Sprintf("0x%X", addr)
	}

	return color.Blue.Sprint(m.Name)
}
