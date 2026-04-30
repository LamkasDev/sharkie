package spirv

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func emitMIMG(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MimgDetails)
	switch details.Op {
	case gcnSpec.MimgOpSample:
		idUint := ctx.GetId(BlockContextIdTypeUint)
		idFloat := ctx.GetId(BlockContextIdTypeFloat)
		idV2Float := ctx.GetId(BlockContextIdTypeV2Float)
		idV4Float := ctx.GetId(BlockContextIdTypeV4Float)
		idSampledImage := ctx.GetId(BlockContextIdTypeSampledImage)
		idPtrUniformSampledImage := ctx.GetId(BlockContextIdPtrUniformSampledImage)

		// Get descriptor index from SGPR[srsrc].
		descriptorIndex := ctx.LoadRegisterPointer(b, gcnSpec.OpSgpr0+details.Srsrc*4)

		// Access bindless array textures[descriptorIndex].
		texturesVar := ctx.GetId(BlockContextIdBindlessTextures)
		ptr := b.EmitAccessChain(idPtrUniformSampledImage, texturesVar, descriptorIndex)
		sampledImage := b.EmitLoad(idSampledImage, ptr)

		// Coordinates from VGPRs. For 2D we need X, Y.
		coordX := b.EmitBitcast(idFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordY := b.EmitBitcast(idFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordinates := b.EmitCompositeConstruct(idV2Float, coordX, coordY)

		// Sample.
		resVec4 := b.EmitImageSampleImplicitLod(idV4Float, sampledImage, coordinates)

		// Write results back to VGPRs based on dmask.
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(idFloat, resVec4, i)
				ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(idUint, val))
				vgprOffset++
			}
		}
	default:
		panic(fmt.Sprintf("unknown mimg op %s", gcnSpec.Mnemotics[gcnSpec.EncMIMG][details.Op]))
	}
}
