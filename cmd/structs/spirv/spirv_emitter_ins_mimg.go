package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMIMG(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*MimgDetails)
	switch details.Op {
	case MimgOpSample:
		idUint := ctx.GetId(BlockContextIdTypeUint)
		idFloat := ctx.GetId(BlockContextIdTypeFloat)
		idV2Float := ctx.GetId(BlockContextIdTypeV2Float)
		idV4Float := ctx.GetId(BlockContextIdTypeV4Float)
		idSampledImage := ctx.GetId(BlockContextIdTypeSampledImage)
		idPtrUniformSampledImage := ctx.GetId(BlockContextIdPtrUniformSampledImage)

		// Get descriptor index from SGPR[srsrc].
		descriptorIndex := ctx.LoadRegisterPointer(b, OpSgpr0+details.Srsrc*4)

		// Access bindless array textures[descriptorIndex].
		texturesVar := ctx.GetId(BlockContextIdBindlessTextures)
		ptr := b.EmitAccessChain(idPtrUniformSampledImage, texturesVar, descriptorIndex)
		sampledImage := b.EmitLoad(idSampledImage, ptr)

		// Coordinates from VGPRs. For 2D we need X, Y.
		coordX := b.EmitBitcast(idFloat, ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr))
		coordY := b.EmitBitcast(idFloat, ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr+1))
		coordinates := b.EmitCompositeConstruct(idV2Float, coordX, coordY)

		// Sample.
		resVec4 := b.EmitImageSampleImplicitLod(idV4Float, sampledImage, coordinates)

		// Write results back to VGPRs based on dmask.
		vgprOffset := uint32(0)
		for i := uint32(0); i < 4; i++ {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(idFloat, resVec4, i)
				ctx.StoreRegisterPointer(b, OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(idUint, val))
				vgprOffset++
			}
		}
	default:
		panic(fmt.Sprintf("unknown mimg op %s", Mnemotics[EncMIMG][details.Op]))
	}
}
