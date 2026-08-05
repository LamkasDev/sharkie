package patcher

var GlobalPatcherRuntime = NewPatcherRuntime()

// PatcherRuntime keeps track of runtime patching state.
type PatcherRuntime struct {
	FailedPatchAddresses map[uint64]bool
}

// NewPatcherRuntime creates a new PatcherRuntime.
func NewPatcherRuntime() *PatcherRuntime {
	return &PatcherRuntime{
		FailedPatchAddresses: make(map[uint64]bool),
	}
}

// IsTcbAccess checks if the instruction at the given instruction pointer is a flagged TCB access.
func (pr *PatcherRuntime) IsTcbAccess(rip uint64) bool {
	return pr.FailedPatchAddresses[rip]
}
