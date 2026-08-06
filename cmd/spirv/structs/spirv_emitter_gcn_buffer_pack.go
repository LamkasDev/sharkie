package structs

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitFormatPackHelper(b *SpvBuilder, ctx *SpirvBlockContext, absoluteAddress, dataFormat, numFormat SpirvId, op uint32, storeVec4 SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint8 := ctx.GetId(BlockContextIdTypeUint8)
	typeUint16 := ctx.GetId(BlockContextIdTypeUint16)
	typeInt := ctx.GetId(BlockContextIdTypeInt)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec2 := ctx.GetId(BlockContextIdTypeV2Float)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typePtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	typePtrPsbUint8 := ctx.GetId(BlockContextIdPtrPsbUint8)
	typePtrPsbUint16 := ctx.GetId(BlockContextIdPtrPsbUint16)
	cFF := ctx.GetConstId(ConstIdUint255)
	cFFFF := ctx.GetConstId(ConstIdUintFFFF)

	// Calculate final aligned address.
	cFFFFFFFFC_64 := b.EmitConstantUint64(typeUint64, 0xFFFFFFFFFFFFFFFC)
	alignedAddress := b.EmitBitwiseAnd(typeUint64, absoluteAddress, cFFFFFFFFC_64)
	translatedAddress := ctx.TranslateAddress(b, alignedAddress)

	// Shift if unaligned access.
	c3_64 := b.EmitConstantUint64(typeUint64, 3)
	mod4_64 := b.EmitBitwiseAnd(typeUint64, absoluteAddress, c3_64)
	mod4 := b.EmitUConvert(typeUint, mod4_64)
	bitShift := b.EmitIMul(typeUint, mod4, b.EmitConstantUint(typeUint, 8))

	// Extract all 4 components from the input vec4.
	f0 := b.EmitCompositeExtract(typeFloat, storeVec4, 0)
	f1 := b.EmitCompositeExtract(typeFloat, storeVec4, 1)
	f2 := b.EmitCompositeExtract(typeFloat, storeVec4, 2)
	f3 := b.EmitCompositeExtract(typeFloat, storeVec4, 3)

	// Resolve number formats.
	isUnorm := b.EmitLogicalOr(typeBool,
		b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint0)),
		b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint1))) // Treat snorm as unorm
	isUscaled := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint2))
	isSscaled := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint3))
	isUint := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint4))
	isSint := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint5))
	isFloat := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint7))
	isAnyIntOrNorm := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isUnorm, isUint), b.EmitLogicalOr(typeBool, isUscaled, isSscaled))
	isAnyIntOrNorm = b.EmitLogicalOr(typeBool, isAnyIntOrNorm, b.EmitLogicalOr(typeBool, isSint, isFloat))

	// Resolve data formats.
	is8 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint1)), isAnyIntOrNorm)
	is16 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint2)), isAnyIntOrNorm)
	is88 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint3)), isAnyIntOrNorm)
	is32 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint4)), isAnyIntOrNorm)
	is1616 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint5)), isAnyIntOrNorm)
	is8888 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint10)), isAnyIntOrNorm)
	is3232 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint11)), isAnyIntOrNorm)
	is16161616 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint12)), isAnyIntOrNorm)
	is323232 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint13)), isAnyIntOrNorm)
	is32323232 := b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint14)), isAnyIntOrNorm)

	// Select raw values based on format.
	is16BitSize := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is16, is1616), is16161616)
	is32BitSize := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is32, is3232), b.EmitLogicalOr(typeBool, is323232, is32323232))

	// Divisor for UNORM.
	divisor := b.EmitSelect(typeFloat, is32BitSize, ctx.GetConstId(ConstIdFloat1), b.EmitSelect(typeFloat, is16BitSize, ctx.GetConstId(ConstIdFloat65535), ctx.GetConstId(ConstIdFloat255)))

	// Unorm values.
	comp0Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f0, divisor))
	comp1Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f1, divisor))
	comp2Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f2, divisor))
	comp3Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f3, divisor))

	// Uint values.
	comp0Uint := b.EmitBitcast(typeUint, f0)
	comp1Uint := b.EmitBitcast(typeUint, f1)
	comp2Uint := b.EmitBitcast(typeUint, f2)
	comp3Uint := b.EmitBitcast(typeUint, f3)

	// Sint values.
	comp0Sint := b.EmitBitcast(typeUint, f0)
	comp1Sint := b.EmitBitcast(typeUint, f1)
	comp2Sint := b.EmitBitcast(typeUint, f2)
	comp3Sint := b.EmitBitcast(typeUint, f3)

	// Uscaled values.
	comp0Uscaled := b.EmitConvertFToU(typeUint, f0)
	comp1Uscaled := b.EmitConvertFToU(typeUint, f1)
	comp2Uscaled := b.EmitConvertFToU(typeUint, f2)
	comp3Uscaled := b.EmitConvertFToU(typeUint, f3)

	// Sscaled values.
	comp0Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f0))
	comp1Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f1))
	comp2Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f2))
	comp3Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f3))

	// If it's a 32-bit float, raw0 is already a 32-bit float. We only pack 16-bit.
	packHalf := func(floatA, floatB SpirvId) SpirvId {
		vec2 := b.EmitCompositeConstruct(typeVec2, floatA, floatB)
		return b.EmitExtInst(typeUint, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpPackHalf2x16, vec2)
	}
	comp0Float16 := packHalf(f0, ctx.GetConstId(ConstIdFloat0))
	comp1Float16 := packHalf(f1, ctx.GetConstId(ConstIdFloat0))
	comp2Float16 := packHalf(f2, ctx.GetConstId(ConstIdFloat0))
	comp3Float16 := packHalf(f3, ctx.GetConstId(ConstIdFloat0))

	// Float values.
	comp0Float := b.EmitSelect(typeUint, is16BitSize, comp0Float16, comp0Uint)
	comp1Float := b.EmitSelect(typeUint, is16BitSize, comp1Float16, comp1Uint)
	comp2Float := b.EmitSelect(typeUint, is16BitSize, comp2Float16, comp2Uint)
	comp3Float := b.EmitSelect(typeUint, is16BitSize, comp3Float16, comp3Uint)

	// Selection chain.
	selectRaw := func(unorm, sint, float, uscaled, sscaled, uintVal SpirvId) SpirvId {
		res := b.EmitSelect(typeUint, isUnorm, unorm, uintVal)
		res = b.EmitSelect(typeUint, isSint, sint, res)
		res = b.EmitSelect(typeUint, isFloat, float, res)
		res = b.EmitSelect(typeUint, isUscaled, uscaled, res)
		res = b.EmitSelect(typeUint, isSscaled, sscaled, res)
		return res
	}
	raw0 := selectRaw(comp0Unorm, comp0Sint, comp0Float, comp0Uscaled, comp0Sscaled, comp0Uint)
	raw1 := selectRaw(comp1Unorm, comp1Sint, comp1Float, comp1Uscaled, comp1Sscaled, comp1Uint)
	raw2 := selectRaw(comp2Unorm, comp2Sint, comp2Float, comp2Uscaled, comp2Sscaled, comp2Uint)
	raw3 := selectRaw(comp3Unorm, comp3Sint, comp3Float, comp3Uscaled, comp3Sscaled, comp3Uint)

	// Mask values to clean bounds.
	raw0_8 := b.EmitBitwiseAnd(typeUint, raw0, cFF)
	raw1_8 := b.EmitBitwiseAnd(typeUint, raw1, cFF)
	raw2_8 := b.EmitBitwiseAnd(typeUint, raw2, cFF)
	raw3_8 := b.EmitBitwiseAnd(typeUint, raw3, cFF)

	raw0_16 := b.EmitBitwiseAnd(typeUint, raw0, cFFFF)
	raw1_16 := b.EmitBitwiseAnd(typeUint, raw1, cFFFF)
	raw2_16 := b.EmitBitwiseAnd(typeUint, raw2, cFFFF)
	raw3_16 := b.EmitBitwiseAnd(typeUint, raw3, cFFFF)

	// Build words.
	word0_88 := b.EmitBitwiseOr(typeUint, raw0_8, b.EmitShiftLeftLogical(typeUint, raw1_8, ctx.GetConstId(ConstIdUint8)))
	word0_1616 := b.EmitBitwiseOr(typeUint, raw0_16, b.EmitShiftLeftLogical(typeUint, raw1_16, ctx.GetConstId(ConstIdUint16)))
	word0_8888_lo := b.EmitBitwiseOr(typeUint, raw0_8, b.EmitShiftLeftLogical(typeUint, raw1_8, ctx.GetConstId(ConstIdUint8)))
	word0_8888_hi := b.EmitBitwiseOr(typeUint, b.EmitShiftLeftLogical(typeUint, raw2_8, ctx.GetConstId(ConstIdUint16)), b.EmitShiftLeftLogical(typeUint, raw3_8, ctx.GetConstId(ConstIdUint24)))
	word0_8888 := b.EmitBitwiseOr(typeUint, word0_8888_lo, word0_8888_hi)
	word0 := b.EmitSelect(typeUint, is32BitSize, raw0,
		b.EmitSelect(typeUint, is16BitSize, b.EmitSelect(typeUint, is1616, word0_1616, b.EmitSelect(typeUint, is16161616, word0_1616, raw0_16)),
			b.EmitSelect(typeUint, is88, word0_88, b.EmitSelect(typeUint, is8888, word0_8888, raw0_8))))
	word1_16161616 := b.EmitBitwiseOr(typeUint, raw2_16, b.EmitShiftLeftLogical(typeUint, raw3_16, ctx.GetConstId(ConstIdUint16)))
	word1 := b.EmitSelect(typeUint, is16161616, word1_16161616, raw1)
	word2 := raw2
	word3 := raw3

	// Dispatch stores based on instruction.
	storeWord0 := func(val SpirvId) {
		// Determine the byte size we need to store.
		is8BitMem := is8
		is16BitMem := b.EmitLogicalOr(typeBool, is16, is88)

		// Allocate labels.
		label8Bit := b.AllocId()
		labelCheck16 := b.AllocId()
		label16Bit := b.AllocId()
		label32Bit := b.AllocId()
		labelMerge16 := b.AllocId()
		labelMergeTotal := b.AllocId()

		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(ConstId64Uint0))
		shiftedVal := b.EmitShiftLeftLogical(typeUint, val, bitShift)

		// Is 8-bit branch.
		b.EmitSelectionMerge(labelMergeTotal, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(is8BitMem, label8Bit, labelCheck16)

		// 8-bit store block.
		b.EmitLabel(label8Bit)
		addr8 := b.EmitIAdd(typeUint64, addr, b.EmitUConvert(typeUint64, mod4))
		ptr8 := b.EmitConvertUToPtr(typePtrPsbUint8, addr8)
		val8 := b.EmitUConvert(typeUint8, val)
		b.EmitStore(ptr8, val8, spec.SpvMemoryAccessAligned, 1)
		b.EmitBranch(labelMergeTotal)

		// Is 16-bit branch.
		b.EmitLabel(labelCheck16)
		b.EmitSelectionMerge(labelMerge16, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(is16BitMem, label16Bit, label32Bit)

		// 16-bit store block.
		b.EmitLabel(label16Bit)
		addr16 := b.EmitIAdd(typeUint64, addr, b.EmitUConvert(typeUint64, mod4))
		ptr16 := b.EmitConvertUToPtr(typePtrPsbUint16, addr16)
		val16 := b.EmitUConvert(typeUint16, val)
		b.EmitStore(ptr16, val16, spec.SpvMemoryAccessAligned, 2)
		b.EmitBranch(labelMerge16)

		// 32-bit store block.
		b.EmitLabel(label32Bit)
		ptr32 := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		b.EmitStore(ptr32, shiftedVal, spec.SpvMemoryAccessAligned, 4)
		b.EmitBranch(labelMerge16)

		// Merge.
		b.EmitLabel(labelMerge16)
		b.EmitBranch(labelMergeTotal)
		b.EmitLabel(labelMergeTotal)
	}
	storeWord := func(offset SpirvId, val SpirvId) {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(offset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		shiftedVal := b.EmitShiftLeftLogical(typeUint, val, bitShift)
		b.EmitStore(ptr, shiftedVal, spec.SpvMemoryAccessAligned, 4)
	}
	storeWord0(word0)
	switch op {
	case gcnSpec.MubufOpStoreDwordx2, gcnSpec.MubufOpStoreFormatXy:
		storeWord(ConstId64Uint4, word1)
	case gcnSpec.MubufOpStoreDwordx3, gcnSpec.MubufOpStoreFormatXyz:
		storeWord(ConstId64Uint4, word1)
		storeWord(ConstId64Uint8, word2)
	case gcnSpec.MubufOpStoreDwordx4, gcnSpec.MubufOpStoreFormatXyzw:
		storeWord(ConstId64Uint4, word1)
		storeWord(ConstId64Uint8, word2)
		storeWord(ConstId64Uint12, word3)
	}
}
