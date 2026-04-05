package gpu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

const LiverpoolConstRamSize = 0x8000

type LiverpoolCommandRing struct {
	Pending []PM4IndirectBuffer
}

const LiverpoolCommandRingSize = unsafe.Sizeof(LiverpoolCommandRing{})

// LiverpoolRegisters mirrors register banks on the Liverpool GPU.
type LiverpoolRegisters struct {
	Config     [gcn.GcnRegBankSize]uint32
	Shader     [gcn.GcnRegBankSize]uint32
	Context    [gcn.GcnRegBankSize]uint32
	UserConfig [gcn.GcnRegBankSize]uint32
}
