package spirv

import (
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
)

type SgprSource struct {
	UserDataOffset int32 // -1 if unknown
}

type SpirvShaderResource struct {
	InstructionOffset uintptr
	FixedSlot         int32 // -1 if dynamic
	RsrcUserData      int32 // UserData offset for T# base
	SampUserData      int32 // UserData offset for S# base
}

func AnalyzeResources(shader *GcnShader) []SpirvShaderResource {
	// Simple analysis: track SGPRs from UserData.
	sgprs := make([]SgprSource, 104)
	for i := range sgprs {
		sgprs[i].UserDataOffset = -1
	}

	// Initialize s0-s15 from UserData.
	userDataOffset := int32(gpu.GcnStageToUserDataOffset[shader.Stage])
	for i := 0; i < 16; i++ {
		sgprs[i].UserDataOffset = userDataOffset + int32(i)
	}

	var resources []SpirvShaderResource

	// Walk all blocks in RPO.
	rpo := shader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.Cfg.Blocks[blockId]
		for _, instr := range block.Instructions {
			switch instr.Encoding {
			case gcnSpec.EncSOP1:
				details := instr.Details.(*gcnSpec.ScalarDetails)
				if details.Op == gcnSpec.Sop1OpMovB32 {
					// sdst = ssrc0
					if details.Src0 >= gcnSpec.OpSgpr0 && details.Src0 <= gcnSpec.OpSgpr103 &&
						details.Dst >= gcnSpec.OpSgpr0 && details.Dst <= gcnSpec.OpSgpr103 {
						sgprs[details.Dst].UserDataOffset = sgprs[details.Src0].UserDataOffset
					}
				} else {
					// Any other SOP1 clears destination.
					if details.Dst >= gcnSpec.OpSgpr0 && details.Dst <= gcnSpec.OpSgpr103 {
						sgprs[details.Dst].UserDataOffset = -1
					}
				}
			case gcnSpec.EncSMRD:
				details := instr.Details.(*gcnSpec.SmrdDetails)
				// s_load_dword*
				// For now, we don't track loads from memory as "Fixed" since they might change.
				// Phase 2 only tracks direct UserData registers.
				for i := range uint32(16) { // max load size
					if details.Dst+i <= 103 {
						sgprs[details.Dst+i].UserDataOffset = -1
					}
				}
			case gcnSpec.EncMIMG:
				details := instr.Details.(*gcnSpec.MimgDetails)
				if details.Op == gcnSpec.MimgOpSample {
					res := SpirvShaderResource{
						InstructionOffset: instr.DwordOffset,
						FixedSlot:         -1,
						RsrcUserData:      sgprs[details.Srsrc*4].UserDataOffset,
						SampUserData:      sgprs[details.Ssamp*4].UserDataOffset,
					}

					// If both are from UserData, we can assign a fixed slot.
					if res.RsrcUserData != -1 && res.SampUserData != -1 {
						res.FixedSlot = int32(len(resources)) + 1024 // Offset fixed slots to avoid collision with dynamic ones?
						// Actually, dynamic ones use indices assigned by discoveryNextIndex.
						// Let's use a safe range for fixed slots, e.g. 1024-2047.
					}
					resources = append(resources, res)
				}
			default:
				// TODO: handle other instructions that might modify SGPRs.
			}
		}
	}

	return resources
}
