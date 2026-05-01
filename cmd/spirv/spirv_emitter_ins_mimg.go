package spirv

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func emitMIMG(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MimgDetails)
	switch details.Op {
	case gcnSpec.MimgOpSample, gcnSpec.MimgOpSampleLz:
		idC0 := ctx.GetConstId(ConstIdUint0)
		idC1 := ctx.GetConstId(ConstIdUint1)
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeV2Float := ctx.GetId(BlockContextIdTypeV2Float)
		typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
		typeSampledImage := ctx.GetId(BlockContextIdTypeSampledImage)
		typePtrUniformSampledImage := ctx.GetId(BlockContextIdPtrUniformSampledImage)
		typePtrStorageUint := ctx.GetId(BlockContextIdPtrStorageUint)

		// Check if we have a fixed slot from analysis.
		var fixedSlot int32 = -1
		for _, res := range ctx.Resources {
			if res.InstructionOffset == instr.DwordOffset {
				fixedSlot = res.FixedSlot
				break
			}
		}

		var vulkanIndex SpirvId
		if false && fixedSlot != -1 {
			vulkanIndex = b.EmitConstantUint(typeUint, uint32(fixedSlot))
		} else {
			// Load the descriptors.
			rsrcLength := uint32(8)
			if details.R128 {
				rsrcLength = 4
			}
			rsrc := make([]SpirvId, rsrcLength)
			for i := range rsrcLength {
				rsrc[i] = ctx.GetGcnSgprId(b, details.Srsrc*4+i)
			}
			samp := make([]SpirvId, 4)
			for i := range uint32(4) {
				samp[i] = ctx.GetGcnSgprId(b, details.Ssamp*4+i)
			}

			// Hash the combined descriptors.
			hash := idC0
			for i := range rsrcLength {
				hash = b.EmitBitwiseXor(typeUint, hash, rsrc[i])
			}
			for i := range uint32(4) {
				hash = b.EmitBitwiseXor(typeUint, hash, samp[i])
			}
			hashIndex := b.EmitBitwiseAnd(typeUint, hash, b.EmitConstantUint(typeUint, 0xFFFF))

			// Lookup in GlobalDescriptorMap.
			idDescriptorMap := ctx.GetId(BlockContextIdGlobalDescriptorMap)
			ptrMapEntry := b.EmitAccessChain(typePtrStorageUint, idDescriptorMap, ctx.GetConstId(ConstIdUint0), hashIndex)
			vulkanIndex = b.EmitLoad(typeUint, ptrMapEntry)

			// Fallback if missing (vulkanIndex == 0).
			isMissing := b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), vulkanIndex, idC0)
			missingLabel := b.AllocId()
			foundLabel := b.AllocId()
			mergeLabel := b.AllocId()
			b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
			b.EmitBranchConditional(isMissing, missingLabel, foundLabel)

			// Missing case.
			b.EmitLabel(missingLabel)
			idMissingBuffer := ctx.GetId(BlockContextIdMissingResourceBuffer)
			ptrCount := b.EmitAccessChain(typePtrStorageUint, idMissingBuffer, idC0)

			// Atomic increment count (scope=device (1), semantics=relaxed (0)).
			offset := b.EmitAtomicIAdd(typeUint, ptrCount, idC1, idC0, idC1)
			b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1, b.EmitString("incremented counter.\n"))

			// Store descriptors to MissingResourceBuffer.
			for i := range rsrcLength {
				ptrDword := b.EmitAccessChain(typePtrStorageUint, idMissingBuffer, idC1, offset, ctx.GetConstId(ConstIdUint0+SpirvId(i)))
				b.EmitStore(ptrDword, rsrc[i])
			}
			for i := range uint32(4) {
				ptrDword := b.EmitAccessChain(typePtrStorageUint, idMissingBuffer, idC1, offset, ctx.GetConstId(ConstIdUint0+SpirvId(rsrcLength+i)))
				b.EmitStore(ptrDword, samp[i])
			}
			b.EmitBranch(mergeLabel)

			// Found case.
			b.EmitLabel(foundLabel)
			b.EmitBranch(mergeLabel)

			// Merge.
			b.EmitLabel(mergeLabel)
		}

		// Access bindless array textures[vulkanIndex].
		texturesVar := ctx.GetId(BlockContextIdBindlessTextures)
		ptr := b.EmitAccessChain(typePtrUniformSampledImage, texturesVar, vulkanIndex)
		sampledImage := b.EmitLoad(typeSampledImage, ptr)

		// Coordinates from VGPRs. For 2D we need X, Y.
		coordX := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordY := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordinates := b.EmitCompositeConstruct(typeV2Float, coordX, coordY)

		// Sample.
		var resVec4 SpirvId
		if ctx.Stage == gcn.GcnShaderStageFragment {
			resVec4 = b.EmitImageSampleImplicitLod(typeV4Float, sampledImage, coordinates)
		} else {
			resVec4 = b.EmitImageSampleExplicitLod(typeV4Float, sampledImage, coordinates, ctx.GetConstId(ConstIdFloat0))
		}

		// Write results back to VGPRs based on dmask.
		vgprOffset := uint32(0)
		for i := range uint32(4) {
			if (details.Dmask>>i)&1 == 1 {
				val := b.EmitCompositeExtract(typeFloat, resVec4, i)
				ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+vgprOffset, b.EmitBitcast(typeUint, val))
				vgprOffset++
			}
		}
	default:
		panic(fmt.Sprintf("unknown mimg op %s", gcnSpec.Mnemotics[gcnSpec.EncMIMG][details.Op]))
	}
}
