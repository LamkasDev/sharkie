package spirv

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func AnalyzeResources(shader *GcnShader) []SpirvShaderResource {
	var resources []SpirvShaderResource
	rpo := shader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.Cfg.Blocks[blockId]
		for _, instr := range block.Instructions {
			switch instr.Encoding {
			case gcnSpec.EncMIMG:
				details := instr.Details.(*gcnSpec.MimgDetails)
				kind := MimgAccessKind(details.Op)
				if kind == ImageAccessUnknown {
					panic("unknown image access")
				}
				res := SpirvShaderResource{
					InstructionOffset: instr.DwordOffset,
					Kind:              kind,
				}
				resources = append(resources, res)
			}
		}
	}

	return resources
}

func MimgAccessKind(op uint32) ImageAccessKind {
	switch op {
	case gcnSpec.MimgOpLoad, gcnSpec.MimgOpLoadMip:
		return ImageAccessLoad
	case gcnSpec.MimgOpStore, gcnSpec.MimgOpStoreMip:
		return ImageAccessStore
	case gcnSpec.MimgOpSample, gcnSpec.MimgOpSampleLz:
		return ImageAccessSample
	default:
		return ImageAccessUnknown
	}
}

func MubufAccessKind(op uint32) BufferAccessKind {
	switch op {
	case gcnSpec.MubufOpLoadFormatX, gcnSpec.MubufOpLoadDword,
		gcnSpec.MubufOpLoadFormatXy, gcnSpec.MubufOpLoadDwordx2,
		gcnSpec.MubufOpLoadFormatXyz, gcnSpec.MubufOpLoadDwordx3,
		gcnSpec.MubufOpLoadFormatXyzw, gcnSpec.MubufOpLoadDwordx4:
		return BufferAccessLoad
	case gcnSpec.MubufOpStoreFormatX, gcnSpec.MubufOpStoreDword,
		gcnSpec.MubufOpStoreFormatXy, gcnSpec.MubufOpStoreDwordx2,
		gcnSpec.MubufOpStoreFormatXyz, gcnSpec.MubufOpStoreDwordx3,
		gcnSpec.MubufOpStoreFormatXyzw, gcnSpec.MubufOpStoreDwordx4:
		return BufferAccessStore
	default:
		return BufferAccessUnknown
	}
}

func MtbufAccessKind(op uint32) BufferAccessKind {
	switch op {
	case gcnSpec.MtbufOpTbufferLoadFormatX, gcnSpec.MtbufOpTbufferLoadFormatXy,
		gcnSpec.MtbufOpTbufferLoadFormatXyz, gcnSpec.MtbufOpTbufferLoadFormatXyzw:
		return BufferAccessLoad
	case gcnSpec.MtbufOpTbufferStoreFormatX, gcnSpec.MtbufOpTbufferStoreFormatXy,
		gcnSpec.MtbufOpTbufferStoreFormatXyz, gcnSpec.MtbufOpTbufferStoreFormatXyzw:
		return BufferAccessStore
	default:
		return BufferAccessUnknown
	}
}
