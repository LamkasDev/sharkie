package spirv

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func AnalyzeResources(shader *GcnShader, ctx SpirvShaderContext) []SpirvShaderResource {
	var resources []SpirvShaderResource
	rpo := shader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instr := &block.Instructions[i]
			if res := analyzeInstruction(instr); res != nil {
				resources = append(resources, *res)
			}
		}
	}

	if vsCtx, ok := ctx.(SpirvVertexShaderContext); ok {
		for _, instr := range vsCtx.FetchShaderInstructions {
			if res := analyzeInstruction(instr); res != nil {
				resources = append(resources, *res)
			}
		}
	}

	return resources
}

func analyzeInstruction(instr *gcnSpec.Instruction) *SpirvShaderResource {
	switch instr.Encoding {
	case gcnSpec.EncMIMG:
		details := instr.Details.(*gcnSpec.MimgDetails)
		kind := MimgAccessKind(details.Op)
		if kind == ImageAccessUnknown {
			panic("unknown image access")
		}
		return &SpirvShaderResource{
			Instruction: instr,
			Type:        SpirvShaderResourceTypeImage,
			Kind:        kind,
		}
	case gcnSpec.EncMUBUF:
		details := instr.Details.(*gcnSpec.MubufDetails)
		kind := MubufAccessKind(details.Op)
		if kind == BufferAccessUnknown {
			panic("unknown buffer access")
		}
		return &SpirvShaderResource{
			Instruction: instr,
			Type:        SpirvShaderResourceTypeBuffer,
			Kind:        kind,
		}
	case gcnSpec.EncMTBUF:
		details := instr.Details.(*gcnSpec.MtbufDetails)
		kind := MtbufAccessKind(details.Op)
		if kind == BufferAccessUnknown {
			panic("unknown buffer access")
		}
		return &SpirvShaderResource{
			Instruction: instr,
			Type:        SpirvShaderResourceTypeBuffer,
			Kind:        kind,
		}
	}

	return nil
}

func MimgAccessKind(op uint32) ImageAccessKind {
	switch op {
	case gcnSpec.MimgOpLoad, gcnSpec.MimgOpLoadMip:
		return ImageAccessLoad
	case gcnSpec.MimgOpStore, gcnSpec.MimgOpStoreMip:
		return ImageAccessStore
	case gcnSpec.MimgOpSample, gcnSpec.MimgOpSampleLz, gcnSpec.MimgOpSampleB, gcnSpec.MimgOpSampleLzO:
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
