package posix

import (
	"unsafe"

	"github.com/gookit/color"
)

type PthreadMutexType uint32

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
}

type PthreadMutexProtocol uint32

const (
	PthreadMutexProtocolNone    = PthreadMutexProtocol(0)
	PthreadMutexProtocolInherit = PthreadMutexProtocol(1)
	PthreadMutexProtocolProtect = PthreadMutexProtocol(2)
)

var MutexProtocolNames = map[PthreadMutexProtocol]string{
	PthreadMutexProtocolNone:    "None",
	PthreadMutexProtocolInherit: "Inherit",
	PthreadMutexProtocolProtect: "Protect",
}

const (
	ThrMutexInitializer         = 0
	ThrAdaptiveMutexInitializer = 1
	ThrMutexDestroyed           = 2
)

type PthreadMutex struct {
	Flags      uint32
	Owner      uintptr
	Count      int32
	SpinLoops  int32
	YieldLoops int32
	Protocol   PthreadMutexProtocol
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
	return color.Blue.Sprint(m.Name)
}
