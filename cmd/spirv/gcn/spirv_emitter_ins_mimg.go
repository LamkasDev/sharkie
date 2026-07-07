package gcn

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

const (
	DiscoveryMapBufferEntries = 65536
	DiscoveryMapBufferSize    = DiscoveryMapBufferEntries * 4
)

func EmitMIMG(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MimgDetails)
	idStaticTextures := ctx.GetId(BlockContextIdStaticTextures)
	idStaticStorageImages := ctx.GetId(BlockContextIdStaticStorageTextures)
	idC0 := ctx.GetConstId(ConstIdUint0)
	idC0F := ctx.GetConstId(ConstIdFloat0)
	typeInt := ctx.GetId(BlockContextIdTypeInt)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeV2Int := ctx.GetId(BlockContextIdTypeV2Int)
	typeV2Float := ctx.GetId(BlockContextIdTypeV2Float)
	typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
	typeSampledImage := ctx.GetId(BlockContextIdTypeSampledImage)
	typeStorageImage := ctx.GetId(BlockContextIdTypeStorageImage)
	typeImage := ctx.GetId(BlockContextIdTypeImage)
	typePtrUniformSampledImage := ctx.GetId(BlockContextIdPtrUniformSampledImage)
	typePtrUniformStorageImage := ctx.GetId(BlockContextIdPtrUniformStorageImage)
	is1D := isMimgResourceType1D(b, ctx, details)

	switch details.Op {
	case gcnSpec.MimgOpSample, gcnSpec.MimgOpSampleLz:
		bindingIndex := ctx.StaticBindingIndexConst(b, instr.DwordOffset)
		ptr := b.EmitAccessChain(typePtrUniformSampledImage, idStaticTextures, bindingIndex)
		sampledImage := b.EmitLoad(typeSampledImage, ptr)

		// Coordinates from VGPRs (X, Y).
		coordX := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordYRaw := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordY := b.EmitSelect(typeFloat, is1D, idC0F, coordYRaw)
		coordinates := b.EmitCompositeConstruct(typeV2Float, coordX, coordY)

		// Sample.
		var resVec4 SpirvId
		switch {
		case details.Op == gcnSpec.MimgOpSampleLz:
			resVec4 = b.EmitImageSampleExplicitLod(typeV4Float, sampledImage, coordinates, idC0F)
		case ctx.Stage == gcn.GcnShaderStageFragment:
			resVec4 = b.EmitImageSampleImplicitLod(typeV4Float, sampledImage, coordinates)
		default:
			resVec4 = b.EmitImageSampleExplicitLod(typeV4Float, sampledImage, coordinates, idC0F)
		}
		// ctx.EmitDebugPrintfLane(b, 0, "MimgOpSample vk_idx=%d coord=%v2f res=%v4f\n", vulkanIndex, coordinates, resVec4)

		// Write results back to VGPRs based on dmask.
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(typeFloat, resVec4, i)
				ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(typeUint, val))
				vgprOffset++
			}
		}
	case gcnSpec.MimgOpLoadMip:
		bindingIndex := ctx.StaticBindingIndexConst(b, instr.DwordOffset)
		ptr := b.EmitAccessChain(typePtrUniformSampledImage, idStaticTextures, bindingIndex)
		sampledImage := b.EmitLoad(typeSampledImage, ptr)

		// Extract image from sampled image.
		image := b.EmitImage(typeImage, sampledImage)

		// Coordinates from VGPRs (X, Y, Lod).
		coordX := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordYRaw := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordY := b.EmitSelect(typeInt, is1D, b.EmitBitcast(typeInt, idC0), coordYRaw)
		coordinates := b.EmitCompositeConstruct(typeV2Int, coordX, coordY)
		lod := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+2))

		// Image fetch.
		resVec4 := b.EmitImageFetch(typeV4Float, image, coordinates, spec.SpvImageOperandsLodMask, lod)
		// ctx.EmitDebugPrintfLane(b, 0, "MimgOpLoadMip vk_idx=%d coord=%v2d lod=%d res=%v4f\n", vulkanIndex, coordinates, lod, resVec4)

		// Write results back to VGPRs based on dmask.
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(typeFloat, resVec4, i)
				ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(typeUint, val))
				vgprOffset++
			}
		}
	case gcnSpec.MimgOpStore:
		bindingIndex := ctx.StaticBindingIndexConst(b, instr.DwordOffset)
		ptr := b.EmitAccessChain(typePtrUniformStorageImage, idStaticStorageImages, bindingIndex)
		storageImage := b.EmitLoad(typeStorageImage, ptr)

		// Coordinates from VGPRs (X, Y).
		coordX := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordYRaw := b.EmitBitcast(typeInt, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordY := b.EmitSelect(typeInt, is1D, b.EmitBitcast(typeInt, idC0), coordYRaw)
		coordinates := b.EmitCompositeConstruct(typeV2Int, coordX, coordY)

		// Data to store from VGPRs.
		dataParts := make([]SpirvId, 4)
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				dataParts[i] = b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset))
				vgprOffset++
			} else {
				dataParts[i] = idC0F
			}
		}
		texel := b.EmitCompositeConstruct(typeV4Float, dataParts...)

		// Image write.
		b.EmitImageWrite(storageImage, coordinates, texel)
		// ctx.EmitDebugPrintfLane(b, 0, "MimgOpStore vk_idx=%d coord=%v2d texel=%v4f\n", vulkanIndex, coordinates, texel)
	default:
		panic(fmt.Sprintf("unknown mimg op %s", gcnSpec.Mnemotics[gcnSpec.EncMIMG][details.Op]))
	}
}

func isMimgResourceType1D(b *SpvBuilder, ctx *SpirvBlockContext, details *gcnSpec.MimgDetails) SpirvId {
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	dw3 := ctx.GetGcnSgprId(b, details.Srsrc*4+3)
	shifted := b.EmitShiftRightLogical(typeUint, dw3, b.EmitConstantUint(typeUint, 28))
	typ := b.EmitBitwiseAnd(typeUint, shifted, b.EmitConstantUint(typeUint, 0xF))

	return b.EmitIEqual(typeBool, typ, b.EmitConstantUint(typeUint, 8))
}
