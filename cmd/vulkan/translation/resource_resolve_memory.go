package translation

import (
	"unsafe"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	spirvGcn "github.com/LamkasDev/sharkie/cmd/spirv/gcn"
)

type ResolvedMemoryAccess struct {
	Instruction *gcnSpec.Instruction
	BaseAddress uintptr
}

func (t *GpuTranslator) applyAndResolveSMRD(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) ResolvedMemoryAccess {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	if details.Dst >= uint32(len(registers)) {
		return ResolvedMemoryAccess{Instruction: instr}
	}
	count := spirvGcn.SmrdLoadDwordCount(details.Op)

	var offset uintptr
	if details.ImmOff {
		if instr.HasLiteral {
			offset = uintptr(instr.Literal)
		} else {
			offset = uintptr(details.Offset * 4)
		}
	} else {
		if details.Offset < uint32(len(registers)) {
			offset = uintptr(registers[details.Offset])
		}
	}

	var address uintptr
	var dwords []uint32
	switch {
	case details.Op >= gcnSpec.SmrdOpBufferLoadDword && details.Op <= gcnSpec.SmrdOpBufferLoadDwordx16:
		base := details.Base * 2
		address = uintptr(registers[base]) | (uintptr(registers[base+1]&0xFFFF) << 32)
		address &= 0xFFFFFFFFFF
		translated := t.TranslateToHostAddress(address)
		if translated == 0 {
			return ResolvedMemoryAccess{Instruction: instr, BaseAddress: address}
		}
		dwords = unsafe.Slice((*uint32)(unsafe.Pointer(translated+offset)), count)
	default:
		base := details.Base * 2
		address = uintptr(registers[base]) | (uintptr(registers[base+1]) << 32)
		address &= 0xFFFFFFFFFF
		translated := t.TranslateToHostAddress(address)
		if translated == 0 {
			return ResolvedMemoryAccess{Instruction: instr, BaseAddress: address}
		}
		dwords = unsafe.Slice((*uint32)(unsafe.Pointer(translated+offset)), count)
	}

	dst := details.Dst
	for i := uint32(0); i < count && int(dst+i) < len(registers); i++ {
		if int(i) < len(dwords) {
			registers[dst+i] = dwords[i]
		} else {
			registers[dst+i] = 0
		}
	}

	return ResolvedMemoryAccess{Instruction: instr, BaseAddress: address}
}
