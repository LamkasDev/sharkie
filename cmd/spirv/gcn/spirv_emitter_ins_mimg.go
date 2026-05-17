package gcn

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
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

func EmitMIMG(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MimgDetails)
	idBindlessTextures := ctx.GetId(BlockContextIdBindlessTextures)
	idBindlessStorageImages := ctx.GetId(BlockContextIdBindlessStorageTextures)
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
		vulkanIndex := getBindlessImageIndex(b, ctx, details, true)

		// Access bindless array bindless_textures[vulkanIndex].
		b.EmitDecorate(vulkanIndex, spec.SpvDecorationNonUniform)
		ptr := b.EmitAccessChain(typePtrUniformSampledImage, idBindlessTextures, vulkanIndex)
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
		vulkanIndex := getBindlessImageIndex(b, ctx, details, false)

		// Access bindless array textures[vulkanIndex]
		b.EmitDecorate(vulkanIndex, spec.SpvDecorationNonUniform)
		ptr := b.EmitAccessChain(typePtrUniformSampledImage, idBindlessTextures, vulkanIndex)
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
		vulkanIndex := getBindlessImageIndex(b, ctx, details, false)

		// Access bindless storage array bindless_storage_images[vulkanIndex]
		b.EmitDecorate(vulkanIndex, spec.SpvDecorationNonUniform)
		ptr := b.EmitAccessChain(typePtrUniformStorageImage, idBindlessStorageImages, vulkanIndex)
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

func getBindlessImageIndex(b *SpvBuilder, ctx *SpirvBlockContext, details *gcnSpec.MimgDetails, hasSampler bool) SpirvId {
	idC0 := ctx.GetConstId(ConstIdUint0)
	idC1 := ctx.GetConstId(ConstIdUint1)
	idFFFF := ctx.GetConstId(ConstIdUintFFFF)
	idFFFFFFFF := ctx.GetConstId(ConstIdUintFFFFFFFF)
	idSemanticsRelUniform := ctx.GetConstId(ConstIdUint0 + 0x44)
	idSemanticsAcqUniform := ctx.GetConstId(ConstIdUint0 + 0x42)
	idSemanticsAcqRelUniform := ctx.GetConstId(ConstIdUint0 + 0x48)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
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
	if hasSampler {
		for i := range uint32(4) {
			samplerDesc[i] = ctx.GetGcnSgprId(b, details.Ssamp*4+i)
		}
	} else {
		for i := range uint32(4) {
			samplerDesc[i] = idC0
		}
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
	/* if ctx.Stage == gcn.GcnShaderStageCompute {
		ctx.EmitDebugPrintfLane(b, 0, "getBindlessImageIndex stage=%d rsrc=v%d addr=0x%x,0x%x hashIdx=0x%x resIdx=%d\n",
			ctx.GetConstId(ConstIdUint0+SpirvId(ctx.Stage)),
			ctx.GetConstId(ConstIdUint0+SpirvId(details.Srsrc)),
			imageDesc[1], imageDesc[0], hashIndex, vulkanIndex)
	} */

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
	vulkanIndexAfter := b.EmitPhi(SpirvId(typeUint),
		SpirvId(oldVal), SpirvId(reportMergeLabel), SpirvId(vulkanIndex), SpirvId(foundLabel))
	vulkanIndex = vulkanIndexAfter

	// Check if out of descriptor indices.
	maxBindlessIndex := b.EmitConstantUint(typeUint, 8191)
	isOutOfRange := b.EmitUGreaterThan(ctx.GetId(BlockContextIdTypeBool), vulkanIndex, maxBindlessIndex)

	return b.EmitSelect(typeUint, isOutOfRange, idC0, vulkanIndex)
}
