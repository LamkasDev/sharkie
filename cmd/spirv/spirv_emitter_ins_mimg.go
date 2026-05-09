package spirv

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

const (
	DiscoveryMapBufferSize = 65536 * 4

	MissingResourceBufferHeader    = 16
	MissingResourceBufferEntrySize = 12 * 4
	MissingResourceBufferEntries   = 1024
	MissingResourceBufferSize      = MissingResourceBufferHeader + MissingResourceBufferEntries*MissingResourceBufferEntrySize
)

func emitMIMG(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MimgDetails)
	switch details.Op {
	case gcnSpec.MimgOpSample, gcnSpec.MimgOpSampleLz:
		idC0 := ctx.GetConstId(ConstIdUint0)
		idC1 := ctx.GetConstId(ConstIdUint1)
		idFFFF := ctx.GetConstId(ConstIdUintFFFF)
		idFFFFFFFF := ctx.GetConstId(ConstIdUintFFFFFFFF)
		idSemanticsRelUniform := ctx.GetConstId(ConstIdUint0 + 0x44)
		idSemanticsAcqUniform := ctx.GetConstId(ConstIdUint0 + 0x42)
		idSemanticsAcqRelUniform := ctx.GetConstId(ConstIdUint0 + 0x48)
		idBindlessTextures := ctx.GetId(BlockContextIdBindlessTextures)

		typeBool := ctx.GetId(BlockContextIdTypeBool)
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeV2Float := ctx.GetId(BlockContextIdTypeV2Float)
		typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
		typeSampledImage := ctx.GetId(BlockContextIdTypeSampledImage)
		typePtrUniformSampledImage := ctx.GetId(BlockContextIdPtrUniformSampledImage)
		typePtrStorageUint := ctx.GetId(BlockContextIdPtrStorageUint)

		// Load the descriptors and hash them.
		imageDescLength := uint32(8)
		if details.R128 {
			imageDescLength = 4
		}
		imageDesc := make([]SpirvId, 8)
		for i := range imageDescLength {
			imageDesc[i] = ctx.GetGcnSgprId(b, details.Srsrc*4+i)
		}
		for i := imageDescLength; i < 8; i++ {
			imageDesc[i] = idC0
		}
		samplerDesc := make([]SpirvId, 4)
		for i := range uint32(4) {
			samplerDesc[i] = ctx.GetGcnSgprId(b, details.Ssamp*4+i)
		}

		// Hash all 12 dwords.
		hash := idC0
		for i := range uint32(8) {
			hash = b.EmitBitwiseXor(typeUint, hash, imageDesc[i])
		}
		for i := range uint32(4) {
			hash = b.EmitBitwiseXor(typeUint, hash, samplerDesc[i])
		}
		hashIndex := b.EmitBitwiseAnd(typeUint, hash, idFFFF)

		// Lookup in global descriptor map.
		idDescriptorMap := ctx.GetId(BlockContextIdGlobalDescriptorMap)
		ptrMapEntry := b.EmitAccessChain(typePtrStorageUint, idDescriptorMap, idC0, hashIndex)

		// Initial check using AtomicLoad (Device scope, Acquire | UniformMemory semantics (0x2 | 0x40 = 0x42)).
		vulkanIndex := b.EmitAtomicLoad(typeUint, ptrMapEntry, idC1, idSemanticsAcqUniform)

		// 1. Check if index is 0, then branch.
		isZero := b.EmitIEqual(typeBool, vulkanIndex, idC0)
		missingLabel := b.AllocId()
		foundLabel := b.AllocId()
		mergeLabel := b.AllocId()
		b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(isZero, missingLabel, foundLabel)

		// 1a. Zero index case - attempt to claim reporting rights by atomically swapping 0 to 0xFFFFFFFF.
		// Use AcquireRelease | UniformMemory (0x8 | 0x40 = 0x48) for success, Acquire | UniformMemory (0x2 | 0x40 = 0x42) for failure.
		b.EmitLabel(missingLabel)
		oldVal := b.EmitAtomicCompareExchange(typeUint, ptrMapEntry, idC1, idSemanticsAcqRelUniform, idSemanticsAcqUniform, idFFFFFFFF, idC0)

		// 2. Check if successfully swapped 0 to 0xFFFFFFFF, then branch.
		shouldReport := b.EmitIEqual(typeBool, oldVal, idC0)
		reportLabel := b.AllocId()
		skipReportLabel := b.AllocId()
		reportMergeLabel := b.AllocId()
		b.EmitSelectionMerge(reportMergeLabel, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(shouldReport, reportLabel, skipReportLabel)

		// 2a. Swapped case, store the descriptors.
		b.EmitLabel(reportLabel)
		idMissingBuffer := ctx.GetId(BlockContextIdMissingResourceBuffer)
		ptrCount := b.EmitAccessChain(typePtrStorageUint, idMissingBuffer, idC0)

		// Atomic increment count.
		offset := b.EmitAtomicIAdd(typeUint, ptrCount, idC1, idSemanticsAcqRelUniform, idC1)

		// 3. Check if missing resource buffer is full, then branch.
		idMissingResourceEntries := b.EmitConstantUint(typeUint, MissingResourceBufferEntries)
		inRange := b.EmitULessThan(typeBool, offset, idMissingResourceEntries)
		storeLabel := b.AllocId()
		unclaimedLabel := b.AllocId()
		storeMergeLabel := b.AllocId()
		b.EmitSelectionMerge(storeMergeLabel, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(inRange, storeLabel, unclaimedLabel)

		// 3a. In bounds case - store descriptors to missing resource buffer.
		b.EmitLabel(storeLabel)
		for i := range uint32(8) {
			ptrDword := b.EmitAccessChain(typePtrStorageUint, idMissingBuffer, idC1, offset, ctx.GetConstId(ConstIdUint0+SpirvId(i)))
			b.EmitStore(ptrDword, imageDesc[i])
		}
		for i := range uint32(4) {
			ptrDword := b.EmitAccessChain(typePtrStorageUint, idMissingBuffer, idC1, offset, ctx.GetConstId(ConstIdUint0+SpirvId(8+i)))
			b.EmitStore(ptrDword, samplerDesc[i])
		}
		b.EmitBranch(storeMergeLabel)
		b.EmitLabel(unclaimedLabel)

		// 3b. Out of bounds case - release reporting rights so someone else can try next frame.
		// Use Release | UniformMemory (0x4 | 0x40 = 0x44).
		b.EmitAtomicStore(ptrMapEntry, idC1, idSemanticsRelUniform, idC0)
		b.EmitBranch(storeMergeLabel)

		// Merge 3.
		b.EmitLabel(storeMergeLabel)
		b.EmitBranch(reportMergeLabel)

		// 2b. Failed swap case.
		b.EmitLabel(skipReportLabel)
		b.EmitBranch(reportMergeLabel)

		// Merge 2.
		b.EmitLabel(reportMergeLabel)
		b.EmitBranch(mergeLabel)

		// 1b. Found index case.
		b.EmitLabel(foundLabel)
		b.EmitBranch(mergeLabel)

		// Merge 1.
		b.EmitLabel(mergeLabel)

		// Pick the most recent index using OpPhi.
		vulkanIndexAfter := b.AllocId()
		b.instr(&b.code, spec.SpvOpPhi, uint32(typeUint), uint32(vulkanIndexAfter),
			uint32(oldVal), uint32(reportMergeLabel),
			uint32(vulkanIndex), uint32(foundLabel))
		vulkanIndex = vulkanIndexAfter

		// Check if out of descriptor indices.
		maxBindlessIndex := b.EmitConstantUint(typeUint, 8191)
		isOutOfRange := b.EmitUGreaterThan(ctx.GetId(BlockContextIdTypeBool), vulkanIndex, maxBindlessIndex)
		vulkanIndex = b.EmitSelect(typeUint, isOutOfRange, idC0, vulkanIndex)

		// Access bindless array textures[vulkanIndex] (use non-uniform decoration on index for correct behavior across subgroup).
		b.EmitDecorate(vulkanIndex, spec.SpvDecorationNonUniform)
		ptr := b.EmitAccessChain(typePtrUniformSampledImage, idBindlessTextures, vulkanIndex)
		sampledImage := b.EmitLoad(typeSampledImage, ptr)

		// Coordinates from VGPRs. For 2D we need X, Y.
		coordX := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr))
		coordY := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr+1))
		coordinates := b.EmitCompositeConstruct(typeV2Float, coordX, coordY)

		// Sample.
		var resVec4 SpirvId
		switch {
		case details.Op == gcnSpec.MimgOpSampleLz:
			resVec4 = b.EmitImageSampleExplicitLod(typeV4Float, sampledImage, coordinates, ctx.GetConstId(ConstIdFloat0))
		case ctx.Stage == gcn.GcnShaderStageFragment:
			resVec4 = b.EmitImageSampleImplicitLod(typeV4Float, sampledImage, coordinates)
		default:
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
