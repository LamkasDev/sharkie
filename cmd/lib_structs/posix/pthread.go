package posix

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

const PthreadMagic = uint32(0xD09BA115)

type PthreadSchedulingPolicy uint32

const (
	PthreadSchedulingPolicyFifo       = PthreadSchedulingPolicy(1)
	PthreadSchedulingPolicyOther      = PthreadSchedulingPolicy(2)
	PthreadSchedulingPolicyRoundRobin = PthreadSchedulingPolicy(3)
)

var SchedulingPolicyNames = map[PthreadSchedulingPolicy]string{
	PthreadSchedulingPolicyFifo:       "Fifo",
	PthreadSchedulingPolicyOther:      "Other",
	PthreadSchedulingPolicyRoundRobin: "RoundRobin",
}

type PthreadInheritScheduling uint32

const (
	PthreadInheritSchedulingExplicit = PthreadInheritScheduling(0)
	PthreadInheritSchedulingInherit  = PthreadInheritScheduling(4)
)

var InheritSchedulingNames = map[PthreadInheritScheduling]string{
	PthreadInheritSchedulingInherit:  "Inherit",
	PthreadInheritSchedulingExplicit: "Explicit",
}

type PthreadDetachState uint32

const (
	PthreadDetachStateJoinable = PthreadDetachState(0)
	PthreadDetachStateDetached = PthreadDetachState(1)
)

var DetachStateNames = map[PthreadDetachState]string{
	PthreadDetachStateJoinable: "Joinable",
	PthreadDetachStateDetached: "Detached",
}

type PthreadScope uint32

const (
	PthreadScopeProcess = PthreadScope(0)
	PthreadScopeSystem  = PthreadScope(2)
)

var ScopeNames = map[PthreadScope]string{
	PthreadScopeProcess: "Process",
	PthreadScopeSystem:  "System",
}

type PthreadOnceState uint32

const (
	PthreadOnceStateNeverDone  = PthreadOnceState(0)
	PthreadOnceStateDone       = PthreadOnceState(1)
	PthreadOnceStateInProgress = PthreadOnceState(2)
	PthreadOnceStateWait       = PthreadOnceState(3)
)

type PthreadOnce struct {
	State uint32
}

const PthreadOnceSize = unsafe.Sizeof(PthreadOnce{})

type Pthread struct {
	Self         uintptr
	Lock         uint32
	Flags        uint32
	TcbSelf      uintptr
	_            [104]byte // Padding yippee!
	StartFunc    uintptr
	Arg          uintptr
	Attr         PthreadAttr
	_            [240]byte // Biggg padding uwu!
	ReturnValue  uintptr
	_            [24]byte
	NamePtr      Cstring
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
