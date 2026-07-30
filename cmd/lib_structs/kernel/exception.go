package kernel

var (
	ExceptionHandlers = make(map[int]uintptr)
)

type Mcontext struct {
	OnStack  uint64
	Rdi      uint64
	Rsi      uint64
	Rdx      uint64
	Rcx      uint64
	R8       uint64
	R9       uint64
	Rax      uint64
	Rbx      uint64
	Rbp      uint64
	R10      uint64
	R11      uint64
	R12      uint64
	R13      uint64
	R14      uint64
	R15      uint64
	TrapNo   uint32
	Fs       uint16
	Gs       uint16
	Addr     uint64
	Flags    uint32
	Es       uint16
	Ds       uint16
	Err      uint64
	Rip      uint64
	Cs       uint64
	Rflags   uint64
	Rsp      uint64
	Ss       uint64
	Len      uint64
	Fpformat uint64
	Ownedfp  uint64
	Lbrfrom  uint64
	Lbrto    uint64
	Aux1     uint64
	Aux2     uint64
	Fpstate  [104]uint64
	Fsbase   uint64
	Gsbase   uint64
	Spare    [6]uint64
}

type ExStack struct {
	SsSp    uintptr
	SsSize  uint64
	SsFlags int32
	Align   int32
}

type Ucontext struct {
	SigMask  [4]uint32
	Padding  [12]int32
	Mcontext Mcontext
	UcLink   uintptr
	UcStack  ExStack
	UcFlags  int32
	Spare    [4]int32
	Field7   [3]int32
}
