package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	spirvCommon "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

func ParseFetchShaderInstructions(shader *gcn.GcnShader, userData []uint32) []*gcnSpec.Instruction {
	var instructions []*gcnSpec.Instruction
	stageBase := spirvStructs.GcnStageToUserDataOffset[shader.Stage]
	registers := gcnSpec.GcnRegisters{}
	for i := range uint32(16) {
		offset := int(stageBase) + int(i)
		if offset < len(userData) {
			registers[i] = userData[offset]
		}
	}

	for _, blockId := range shader.Cfg.ReversePostOrder() {
		block := &shader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instruction := &block.Instructions[i]
			instructions = append(instructions, instruction)
			switch instruction.Encoding {
			case gcnSpec.EncSOP1:
				details := instruction.Details.(*gcnSpec.ScalarDetails)
				if details.Op == gcnSpec.Sop1OpSetpcB64 {
					return instructions
				}
			}
		}
	}

	return instructions
}

func resolveMUBUFQuick(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) *spirvCommon.FetchAttribute {
	details := instr.Details.(*gcnSpec.MubufDetails)

	numElements := uint32(1)
	if details.Op == gcnSpec.MubufOpLoadFormatXy || details.Op == gcnSpec.MubufOpLoadDwordx2 {
		numElements = 2
	} else if details.Op == gcnSpec.MubufOpLoadFormatXyz || details.Op == gcnSpec.MubufOpLoadDwordx3 {
		numElements = 3
	} else if details.Op == gcnSpec.MubufOpLoadFormatXyzw || details.Op == gcnSpec.MubufOpLoadDwordx4 {
		numElements = 4
	}

	if details.Op == gcnSpec.MubufOpLoadFormatX || details.Op == gcnSpec.MubufOpLoadFormatXy ||
		details.Op == gcnSpec.MubufOpLoadFormatXyz || details.Op == gcnSpec.MubufOpLoadFormatXyzw {
		return &spirvCommon.FetchAttribute{
			DestVgpr:    uint32(details.Vdata),
			BufferIndex: uint32(details.Srsrc) / 4,
			NumElements: numElements,
		}
	}
	return nil
}

func GetFetchShaderPC(shader *gcn.GcnShader, userData []uint32) uintptr {
	if shader == nil {
		return 0
	}
	stageBase := spirvStructs.GcnStageToUserDataOffset[shader.Stage]
	registers := gcnSpec.GcnRegisters{}
	for i := range uint32(16) {
		offset := int(stageBase) + int(i)
		if offset < len(userData) {
			registers[i] = userData[offset]
		}
	}

	for _, blockId := range shader.Cfg.ReversePostOrder() {
		block := &shader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instr := &block.Instructions[i]
			switch instr.Encoding {
			case gcnSpec.EncSOP1:
				applySOP1(instr, &registers)
				details := instr.Details.(*gcnSpec.ScalarDetails)
				if details.Op == gcnSpec.Sop1OpSwappcB64 {
					fetchPCLo := registers[details.Src0]
					fetchPCHi := registers[details.Src0+1]
					return uintptr(fetchPCLo) | (uintptr(fetchPCHi) << 32)
				}
			case gcnSpec.EncSOP2:
				applySOP2(instr, &registers)
			case gcnSpec.EncSOPC:
				applySOPC(instr, &registers)
			case gcnSpec.EncSMRD:
				applySMRD(instr, &registers)
			}
		}
	}
	return 0
}

// ResolvedBufferAccess is a T# resolved at a specific MUBUF instruction for a user-data snapshot.
type ResolvedBufferAccess struct {
	DestVgpr   uint32
	Descriptor spirvStructs.BufferDescriptor
}

// ResolveBufferResources simulates scalar SGPR updates, then resolves T# descriptors at
// precomputed MUBUF sites based on the quick layout.
func ResolveBufferResources(shader *gcn.GcnShader, userData []uint32) []ResolvedBufferAccess {
	stageBase := spirvStructs.GcnStageToUserDataOffset[shader.Stage]
	registers := gcnSpec.GcnRegisters{}
	for i := range 16 {
		offset := int(stageBase) + i
		if offset < len(userData) {
			registers[i] = userData[offset]
		}
	}

	var accesses []ResolvedBufferAccess
	rpo := shader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instr := &block.Instructions[i]
			switch instr.Encoding {
			case gcnSpec.EncSOP1:
				applySOP1(instr, &registers)
			case gcnSpec.EncSOP2:
				applySOP2(instr, &registers)
			case gcnSpec.EncSOPC:
				applySOPC(instr, &registers)
			case gcnSpec.EncSMRD:
				applySMRD(instr, &registers)
			case gcnSpec.EncMUBUF:
				accesses = append(accesses, resolveMUBUF(instr, &registers))
			}
		}
	}
	return accesses
}

func resolveMUBUF(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) ResolvedBufferAccess {
	details := instr.Details.(*gcnSpec.MubufDetails)
	baseSgpr := int(details.Srsrc * 4)
	var dwords [4]uint32
	for j := range 4 {
		if baseSgpr+j < len(registers) {
			dwords[j] = registers[baseSgpr+j]
		}
	}
	desc := spirvStructs.NewBufferDescriptor(dwords[:])
	return ResolvedBufferAccess{
		DestVgpr:   details.Vdata,
		Descriptor: desc,
	}
}
