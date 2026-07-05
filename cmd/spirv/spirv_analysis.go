package spirv

import (
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func AnalyzeResources(shader *GcnShader) []SpirvShaderResource {
	sgprs := make([]SgprSource, 104)
	for i := range sgprs {
		sgprs[i].UserDataOffset = -1
	}

	userDataOffset := int32(structs.GcnStageToUserDataOffset[shader.Stage])
	for i := 0; i < 16; i++ {
		sgprs[i].UserDataOffset = userDataOffset + int32(i)
	}

	var resources []SpirvShaderResource

	rpo := shader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.Cfg.Blocks[blockId]
		for _, instr := range block.Instructions {
			switch instr.Encoding {
			case gcnSpec.EncSOP1:
				details := instr.Details.(*gcnSpec.ScalarDetails)
				if details.Op == gcnSpec.Sop1OpMovB32 {
					if details.Src0 <= gcnSpec.OpSgpr103 && details.Dst <= gcnSpec.OpSgpr103 {
						sgprs[details.Dst].UserDataOffset = sgprs[details.Src0].UserDataOffset
					}
				} else if details.Dst <= gcnSpec.OpSgpr103 {
					sgprs[details.Dst].UserDataOffset = -1
				}
			case gcnSpec.EncSMRD:
				details := instr.Details.(*gcnSpec.SmrdDetails)
				for i := range uint32(16) {
					if details.Dst+i <= 103 {
						sgprs[details.Dst+i].UserDataOffset = -1
					}
				}
			case gcnSpec.EncMIMG:
				details := instr.Details.(*gcnSpec.MimgDetails)
				kind, ok := mimgAccessKind(details.Op)
				if !ok {
					continue
				}
				res := SpirvShaderResource{
					InstructionOffset: instr.DwordOffset,
					Kind:              kind,
					RsrcUserData:      sgprs[details.Srsrc*4].UserDataOffset,
					SampUserData:      -1,
				}
				if kind == ImageAccessSample {
					res.SampUserData = sgprs[details.Ssamp*4].UserDataOffset
				}
				resources = append(resources, res)
			}
		}
	}

	return resources
}

func mimgAccessKind(op uint32) (ImageAccessKind, bool) {
	switch op {
	case gcnSpec.MimgOpLoad:
		return ImageAccessLoad, true
	case gcnSpec.MimgOpLoadMip:
		return ImageAccessLoadMip, true
	case gcnSpec.MimgOpStore:
		return ImageAccessStore, true
	case gcnSpec.MimgOpStoreMip:
		return ImageAccessStoreMip, true
	case gcnSpec.MimgOpSample:
		return ImageAccessSample, true
	default:
		return 0, false
	}
}
