package translation

import (
	"math"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	spirvCommon "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"go101.org/nstd"
)

// ResolveResources simulates scalar SGPR updates and resolves T# descriptors.
func (t *GpuTranslator) ResolveResources(shader *gcn.GcnShader, userData []uint32) ([]ResolvedImageAccess, []ResolvedBufferAccess, []ResolvedMemoryAccess) {
	stageBase := spirvStructs.GcnStageToUserDataOffset[shader.Stage]
	registers := gcnSpec.GcnRegisters{}
	for i := range 16 {
		offset := int(stageBase) + i
		registers[i] = userData[offset]
	}

	var imageAccesses []ResolvedImageAccess
	var bufferAccesses []ResolvedBufferAccess
	var memoryAccesses []ResolvedMemoryAccess
	rpo := shader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instr := &block.Instructions[i]
			imageAccesses, bufferAccesses, memoryAccesses = t.resolveResourcesIns(shader, instr, &registers, imageAccesses, bufferAccesses, memoryAccesses)
		}
	}

	return imageAccesses, bufferAccesses, memoryAccesses
}

func (t *GpuTranslator) GetMubufFormats(shader *gcn.GcnShader, userData []uint32) []spirvCommon.SpirvMubufFormat {
	_, bufferAccesses, _ := t.ResolveResources(shader, userData)
	var formats []spirvCommon.SpirvMubufFormat
	for _, access := range bufferAccesses {
		switch access.Instruction.Encoding {
		case gcnSpec.EncMUBUF, gcnSpec.EncMTBUF:
			formats = append(formats, spirvCommon.SpirvMubufFormat{
				PC:         uint32(access.Instruction.DwordOffset),
				DataFormat: access.Descriptor.DataFormat,
				NumFormat:  access.Descriptor.NumFormat,
			})
		}
	}

	return formats
}

func (t *GpuTranslator) resolveResourcesIns(shader *gcn.GcnShader, instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters, imageAccesses []ResolvedImageAccess, bufferAccesses []ResolvedBufferAccess, memoryAccesses []ResolvedMemoryAccess) ([]ResolvedImageAccess, []ResolvedBufferAccess, []ResolvedMemoryAccess) {
	switch instr.Encoding {
	case gcnSpec.EncSOP1:
		details := instr.Details.(*gcnSpec.ScalarDetails)
		if details.Op == gcnSpec.Sop1OpSwappcB64 {
			fetchPCLo := registers[details.Src0]
			fetchPCHi := registers[details.Src0+1]
			fetchShaderAddress := uintptr(fetchPCLo) | (uintptr(fetchPCHi) << 32)
			if fetchShaderAddress != 0 {
				fetchShader := gpu.GlobalLiverpool.GetShader(gcn.GcnShaderStageFetch, fetchShaderAddress)
				if fetchShader != nil {
					fetchInstructions := ParseFetchShaderInstructions(fetchShader)
					for _, fetchInstr := range fetchInstructions {
						imageAccesses, bufferAccesses, memoryAccesses = t.resolveResourcesIns(shader, fetchInstr, registers, imageAccesses, bufferAccesses, memoryAccesses)
					}
				}
			}
		}
		applySOP1(shader, instr, registers)
	case gcnSpec.EncSOP2:
		applySOP2(instr, registers)
	case gcnSpec.EncSOPC:
		applySOPC(instr, registers)
	case gcnSpec.EncSOPK:
		applySOPK(instr, registers)
	case gcnSpec.EncSMRD:
		access := t.applyAndResolveSMRD(instr, registers)
		memoryAccesses = append(memoryAccesses, access)
	case gcnSpec.EncMIMG:
		imageAccesses = append(imageAccesses, resolveMIMG(instr, registers))
	case gcnSpec.EncMUBUF:
		bufferAccesses = append(bufferAccesses, resolveMUBUF(instr, registers))
	case gcnSpec.EncMTBUF:
		bufferAccesses = append(bufferAccesses, resolveMTBUF(instr, registers))
	}

	return imageAccesses, bufferAccesses, memoryAccesses
}

func applySOP1(shader *gcn.GcnShader, instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
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

func applySOPK(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	if details.Dst >= uint32(len(registers)) {
		return
	}
	dst := int(details.Dst - gcnSpec.OpSgpr0)

	switch details.Op {
	case gcnSpec.SopkOpMovkI32:
		registers[dst] = uint32(int32(int16(instr.Literal)))
	}
}

func readScalarOperand(op uint32, instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) uint32 {
	switch {
	case op <= gcnSpec.OpExecHi:
		return registers[op]
	case op >= gcnSpec.OpInt0 && op <= gcnSpec.OpPosInt64:
		return op - gcnSpec.OpInt0
	case op >= gcnSpec.OpNegInt1 && op <= gcnSpec.OpNegInt16:
		return uint32(int32(-1 - int32(op-gcnSpec.OpNegInt1)))
	case op >= gcnSpec.OpFloat05 && op <= gcnSpec.OpFloatNeg40:
		switch op {
		case gcnSpec.OpFloat05:
			return math.Float32bits(0.5)
		case gcnSpec.OpFloatNeg05:
			return math.Float32bits(-0.5)
		case gcnSpec.OpFloat10:
			return math.Float32bits(1.0)
		case gcnSpec.OpFloatNeg10:
			return math.Float32bits(-1.0)
		case gcnSpec.OpFloat20:
			return math.Float32bits(2.0)
		case gcnSpec.OpFloatNeg20:
			return math.Float32bits(-2.0)
		case gcnSpec.OpFloat40:
			return math.Float32bits(4.0)
		case gcnSpec.OpFloatNeg40:
			return math.Float32bits(-4.0)
		default:
			panic("unhandled")
		}
	case op >= gcnSpec.OpVccz && op <= gcnSpec.OpScc:
		return registers[op]
	case op == gcnSpec.OpLiteral && instr.HasLiteral:
		return instr.Literal
	case op >= gcnSpec.OpVgpr0:
		panic("unhandled")
	}

	return 0
}
