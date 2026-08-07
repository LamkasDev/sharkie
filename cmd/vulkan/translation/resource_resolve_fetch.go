package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

func ParseFetchShaderInstructions(shader *gcn.GcnShader) []*gcnSpec.Instruction {
	var instructions []*gcnSpec.Instruction
	for _, blockId := range shader.Cfg.ReversePostOrder() {
		block := &shader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instruction := &block.Instructions[i]
			instructions = append(instructions, instruction)
		}
	}

	return instructions
}

func (t *GpuTranslator) GetFetchShaderPC(shader *gcn.GcnShader, userData []uint32) uintptr {
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
				applySOP1(shader, instr, &registers)
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
				t.applySMRD(instr, &registers)
			}
		}
	}
	return 0
}
