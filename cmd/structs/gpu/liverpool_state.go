package gpu

import "github.com/LamkasDev/sharkie/cmd/structs/gcn"

// LiverpoolRegisters mirrors register banks on the Liverpool GPU.
type LiverpoolRegisters struct {
	Config     [gcn.GcnRegBankSize]uint32
	Shader     [gcn.GcnRegBankSize]uint32
	Context    [gcn.GcnRegBankSize]uint32
	UserConfig [gcn.GcnRegBankSize]uint32
}

// LiverpoolDrawState holds state derived from SET_* packets that is needed to decode draw calls.
// It is reset when Walk() clears the rings.
type LiverpoolDrawState struct {
	IndexType uint32  // 0 = 16-bit, 1 = 32-bit
	IndexBase uintptr // CPU-visible address of the current index buffer
}
