package posix

import (
	"unsafe"
)

type PthreadAttrFlags uint32

const (
	PthreadAttrFlagsDetached     = PthreadAttrFlags(1)
	PthreadAttrFlagsScopeSystem  = PthreadAttrFlags(2)
	PthreadAttrFlagsInheritSched = PthreadAttrFlags(4)
	PthreadAttrFlagsNoFloat      = PthreadAttrFlags(8)
	PthreadAttrFlagsStackUser    = PthreadAttrFlags(0x100)
)

type PthreadAttr struct {
	SchedulingPolicy  PthreadSchedulingPolicy
	InheritScheduling PthreadInheritScheduling
	Priority          int32
	Suspend           int32
	Flags             PthreadAttrFlags
	_                 [4]byte // Padding yippee!
	StackAddress      uintptr
	StackSize         uint64
	GuardSize         uint64
	CpuSetSize        uint64
	CpuSet            uintptr
}

const PthreadAttrSize = unsafe.Sizeof(PthreadAttr{})
