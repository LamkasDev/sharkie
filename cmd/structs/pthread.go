package structs

type PthreadSchedulingPolicy uint32
type PthreadAttrFlags uint32

const (
	PthreadSchedulingPolicyFifo       = PthreadSchedulingPolicy(1)
	PthreadSchedulingPolicyOther      = PthreadSchedulingPolicy(2)
	PthreadSchedulingPolicyRoundRobin = PthreadSchedulingPolicy(3)
)

const (
	PthreadAttrFlagsDetached = PthreadAttrFlags(1)
	PthreadFlagsScopeSystem  = PthreadAttrFlags(2)
	PthreadFlagsInheritSched = PthreadAttrFlags(4)
	PthreadFlagsNoFloat      = PthreadAttrFlags(8)
	PthreadFlagsStackUser    = PthreadAttrFlags(0x100)
)

type Pthread struct {
	ThreadId            int32
	Flags               uint32
	_                   [4]byte // Padding.
	ReturnValue         uintptr
	Error               int32
	_                   [4]byte   // More padding yippee!
	_                   [452]byte // Biggg padding uwu!
	CleanupHandlerStack uintptr
	Name                [32]byte
}

type PthreadAttr struct {
	SchedulingPolicy  PthreadSchedulingPolicy
	SchedulingInherit int32
	Priority          int32
	Suspend           int32
	_                 [4]byte // Padding yay!
	Flags             PthreadAttrFlags
	_                 [4]byte // More padding yippee!
	StackAddress      uintptr
	StackSize         uintptr
	GuardSize         uintptr
	CpuSetSize        uintptr
	CpuSet            uintptr
}
