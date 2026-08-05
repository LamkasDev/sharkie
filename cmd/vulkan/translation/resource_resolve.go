package translation

import (
	"unsafe"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	spirvGcn "github.com/LamkasDev/sharkie/cmd/spirv/gcn"
	"go101.org/nstd"
)

func applySOP1(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	if details.Dst >= uint32(len(registers)) {
		return
	}
	dst := int(details.Dst - gcnSpec.OpSgpr0)

	switch details.Op {
	case gcnSpec.Sop1OpMovB32:
		registers[dst] = readScalarOperand(details.Src0, instr, registers)
	case gcnSpec.Sop1OpAndSaveexecB64:
		registers[dst] = registers[gcnSpec.OpExecLo]
		registers[dst+1] = registers[gcnSpec.OpExecHi]

		src0Lo := readScalarOperand(details.Src0, instr, registers)
		src0Hi := readScalarOperand(details.Src0+1, instr, registers)
		resLo := src0Lo & registers[gcnSpec.OpExecLo]
		resHi := src0Hi & registers[gcnSpec.OpExecHi]
		registers[gcnSpec.OpExecLo] = resLo
		registers[gcnSpec.OpExecHi] = resHi
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi((resLo | resHi) != 0))
	}
}

func applySOP2(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	if details.Dst >= uint32(len(registers)) {
		return
	}
	dst := int(details.Dst - gcnSpec.OpSgpr0)
	src0 := readScalarOperand(details.Src0, instr, registers)
	src1 := readScalarOperand(details.Src1, instr, registers)

	switch details.Op {
	case gcnSpec.Sop2OpCselectB32:
		registers[dst] = src0
		if registers[gcnSpec.OpScc] != 0 {
			registers[dst] = src1
		}
	case gcnSpec.Sop2OpLshlB32:
		res := src0 << (src1 & 31)
		registers[dst] = res
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi(res != 0))
	case gcnSpec.Sop2OpAndB32:
		res := src0 & src1
		registers[dst] = res
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi(res != 0))
	case gcnSpec.Sop2OpAndB64:
		src0Hi := readScalarOperand(details.Src0+1, instr, registers)
		src1Hi := readScalarOperand(details.Src1+1, instr, registers)
		resLo := src0 & src1
		resHi := src0Hi & src1Hi
		registers[dst] = resLo
		registers[dst+1] = resHi
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi((resLo | resHi) != 0))
	case gcnSpec.Sop2OpBfeU32:
		offset := src1 & 0x7F
		width := (src1 >> 16) & 0x7F
		var res uint32
		if width == 0 {
			res = 0
		} else if width+offset <= 32 {
			// Extract bits [offset : offset+width-1].
			res = (src0 >> offset) & ((1 << width) - 1)
		} else {
			// Shift right by offset.
			res = src0 >> offset
		}
		registers[dst] = res
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi(res != 0))
	}
}

func applySOPC(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	src0 := readScalarOperand(details.Src0, instr, registers)
	src1 := readScalarOperand(details.Src1, instr, registers)

	switch details.Op {
	case gcnSpec.SopcOpCmpEqU32:
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi(src0 == src1))
	}
}

func readScalarOperand(op uint32, instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) uint32 {
	switch {
	case op <= gcnSpec.OpExecHi:
		return registers[op]
	case op >= gcnSpec.OpInt0 && op <= gcnSpec.OpPosInt64:
		return op - gcnSpec.OpInt0
	case op >= gcnSpec.OpNegInt1 && op <= gcnSpec.OpNegInt16:
		panic("unhandled")
	case op >= gcnSpec.OpFloat05 && op <= gcnSpec.OpFloatNeg40:
		panic("unhandled")
	case op >= gcnSpec.OpVccz && op <= gcnSpec.OpScc:
		return registers[op]
	case op == gcnSpec.OpLiteral && instr.HasLiteral:
		return instr.Literal
	case op >= gcnSpec.OpVgpr0:
		panic("unhandled")
	}

	return 0
}

func applySMRD(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	if details.Dst >= uint32(len(registers)) {
		return
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
		dwords = unsafe.Slice((*uint32)(unsafe.Pointer(address+offset)), count)
	default:
		base := details.Base * 2
		address = uintptr(registers[base]) | (uintptr(registers[base+1]) << 32)
		address &= 0xFFFFFFFFFF
		dwords = unsafe.Slice((*uint32)(unsafe.Pointer(address+offset)), count)
	}

	dst := details.Dst
	for i := uint32(0); i < count && int(dst+i) < len(registers); i++ {
		if int(i) < len(dwords) {
			registers[dst+i] = dwords[i]
		} else {
			registers[dst+i] = 0
		}
	}
}
