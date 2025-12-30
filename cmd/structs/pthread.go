package structs

import "unsafe"

type PthreadSchedulingPolicy uint32
type PthreadInheritScheduling uint32
type PthreadDetachState uint32
type PthreadScope uint32
type PthreadAttrFlags uint32

const PthreadMagic = uint32(0xD09BA115)

const (
	PthreadSchedulingPolicyFifo       = PthreadSchedulingPolicy(0)
	PthreadSchedulingPolicyOther      = PthreadSchedulingPolicy(2)
	PthreadSchedulingPolicyRoundRobin = PthreadSchedulingPolicy(3)
)

var SchedulingPolicyNames = map[PthreadSchedulingPolicy]string{
	PthreadSchedulingPolicyFifo:       "Fifo",
	PthreadSchedulingPolicyOther:      "Other",
	PthreadSchedulingPolicyRoundRobin: "RoundRobin",
}

const (
	PthreadInheritSchedulingExplicit = PthreadInheritScheduling(0)
	PthreadInheritSchedulingInherit  = PthreadInheritScheduling(4)
)

var InheritSchedulingNames = map[PthreadInheritScheduling]string{
	PthreadInheritSchedulingInherit:  "Inherit",
	PthreadInheritSchedulingExplicit: "Explicit",
}

const (
	PthreadDetachStateJoinable = PthreadDetachState(0)
	PthreadDetachStateDetached = PthreadDetachState(1)
)

var DetachStateNames = map[PthreadDetachState]string{
	PthreadDetachStateJoinable: "Joinable",
	PthreadDetachStateDetached: "Detached",
}

const (
	PthreadScopeProcess = PthreadScope(0)
	PthreadScopeSystem  = PthreadScope(2)
)

var ScopeNames = map[PthreadScope]string{
	PthreadScopeProcess: "Process",
	PthreadScopeSystem:  "System",
}

const (
	PthreadAttrFlagsDetached     = PthreadAttrFlags(1)
	PthreadAttrFlagsScopeSystem  = PthreadAttrFlags(2)
	PthreadAttrFlagsInheritSched = PthreadAttrFlags(4)
	PthreadAttrFlagsNoFloat      = PthreadAttrFlags(8)
	PthreadAttrFlagsStackUser    = PthreadAttrFlags(0x100)
)

type Pthread struct {
	Self         uintptr
	_            uintptr
	TcbSelf      uintptr
	_            [104]byte // Padding yippee!
	StartFunc    uintptr
	Arg          uintptr
	Attr         PthreadAttr
	_            [240]byte // Biggg padding uwu!
	ReturnValue  uintptr
	_            [24]byte
	NamePtr      uintptr
	CleanupStack uintptr
	_            [44]byte
	Magic        uint32
	_            [480]byte
}

const PthreadSize = unsafe.Sizeof(Pthread{})

type PthreadCleanupEntry struct {
	Next       uintptr
	Handler    uintptr
	Arg        uintptr
	ShouldFree int32
	_          [4]byte
}

const PthreadCleanupEntrySize = unsafe.Sizeof(PthreadCleanupEntry{})

type PthreadAttr struct {
	SchedulingPolicy  PthreadSchedulingPolicy
	InheritScheduling PthreadInheritScheduling
	Priority          int32
	Suspend           int32
	Flags             PthreadAttrFlags
	_                 [4]byte // Padding yippee!
	StackAddress      uintptr
	StackSize         uintptr
	GuardSize         uintptr
	CpuSetSize        uintptr
	CpuSet            uintptr
}

const PthreadAttrSize = unsafe.Sizeof(PthreadAttr{})
