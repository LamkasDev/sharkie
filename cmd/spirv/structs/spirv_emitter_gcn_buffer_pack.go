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

	translatedAddress := ctx.TranslateAddress(b, absoluteAddress)

	// Extract components.
	f0 := b.EmitCompositeExtract(typeFloat, storeVec4, 0)
	f1 := b.EmitCompositeExtract(typeFloat, storeVec4, 1)
	f2 := b.EmitCompositeExtract(typeFloat, storeVec4, 2)
	f3 := b.EmitCompositeExtract(typeFloat, storeVec4, 3)

	isTypedStore := false
	switch op {
	case gcnSpec.MubufOpStoreFormatX, gcnSpec.MubufOpStoreFormatXy, gcnSpec.MubufOpStoreFormatXyz, gcnSpec.MubufOpStoreFormatXyzw:
		isTypedStore = true
	}

	is8 := ctx.GetId(BlockContextIdFalse)
	is16 := ctx.GetId(BlockContextIdFalse)
	is88 := ctx.GetId(BlockContextIdFalse)
	is32 := ctx.GetId(BlockContextIdFalse)
	is1616 := ctx.GetId(BlockContextIdFalse)
	is101111 := ctx.GetId(BlockContextIdFalse)
	is1010102 := ctx.GetId(BlockContextIdFalse)
	is2101010 := ctx.GetId(BlockContextIdFalse)
	is8888 := ctx.GetId(BlockContextIdFalse)
	is3232 := ctx.GetId(BlockContextIdFalse)
	is16161616 := ctx.GetId(BlockContextIdFalse)
	is323232 := ctx.GetId(BlockContextIdFalse)
	is32323232 := ctx.GetId(BlockContextIdFalse)

	is16BitSize := ctx.GetId(BlockContextIdFalse)
	is32BitSize := ctx.GetId(BlockContextIdTrue)
	isPacked10Format := ctx.GetId(BlockContextIdFalse)

	isUnorm := ctx.GetId(BlockContextIdFalse)
	isSnorm := ctx.GetId(BlockContextIdFalse)
	isUscaled := ctx.GetId(BlockContextIdFalse)
	isSscaled := ctx.GetId(BlockContextIdFalse)
	isUint := ctx.GetId(BlockContextIdFalse)
	isSint := ctx.GetId(BlockContextIdFalse)
	isSnormOgl := ctx.GetId(BlockContextIdFalse)
	isFloat := ctx.GetId(BlockContextIdFalse)
	isAnySnorm := ctx.GetId(BlockContextIdFalse)

	// Format resolution.
	if isTypedStore {
		isUnorm = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint0))
		isSnorm = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint1))
		isUscaled = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint2))
		isSscaled = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint3))
		isUint = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint4))
		isSint = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint5))
		isSnormOgl = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint6))
		isFloat = b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint7))
		isAnySnorm = b.EmitLogicalOr(typeBool, isSnorm, isSnormOgl)

		isAnyNorm := b.EmitLogicalOr(typeBool, isUnorm, isAnySnorm)
		isAnyIntOrNorm := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAnyNorm, isUint), b.EmitLogicalOr(typeBool, isUscaled, isSscaled))
		isAnyIntOrNorm = b.EmitLogicalOr(typeBool, isAnyIntOrNorm, b.EmitLogicalOr(typeBool, isSint, isFloat))

		is8 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint1)), isAnyIntOrNorm)
		is16 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint2)), isAnyIntOrNorm)
		is88 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint3)), isAnyIntOrNorm)
		is32 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint4)), isAnyIntOrNorm)
		is1616 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint5)), isAnyIntOrNorm)
		is101111 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint6)), isAnyIntOrNorm)
		is1010102 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint8)), isAnyIntOrNorm)
		is2101010 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint9)), isAnyIntOrNorm)
		is8888 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint10)), isAnyIntOrNorm)
		is3232 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint11)), isAnyIntOrNorm)
		is16161616 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint12)), isAnyIntOrNorm)
		is323232 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint13)), isAnyIntOrNorm)
		is32323232 = b.EmitLogicalAnd(typeBool, b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint14)), isAnyIntOrNorm)

		isPacked10Format = b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is101111, is1010102), is2101010)
		is16BitSize = b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is16, is1616), is16161616)
		is32BitSize = b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is32, is3232), b.EmitLogicalOr(typeBool, is323232, is32323232))
	}

	// Float multipliers & divisors.
	divisor := b.EmitSelect(typeFloat, is32BitSize, ctx.GetConstId(ConstIdFloat1), b.EmitSelect(typeFloat, is16BitSize, ctx.GetConstId(ConstIdFloat65535), ctx.GetConstId(ConstIdFloat255)))

	div_2_10_10_10_x := b.EmitConstantFloat(typeFloat, 1023.0)
	div_2_10_10_10_w := b.EmitConstantFloat(typeFloat, 3.0)
	div0Packed := b.EmitSelect(typeFloat, is2101010, div_2_10_10_10_x, b.EmitSelect(typeFloat, is1010102, div_2_10_10_10_w, div_2_10_10_10_x))
	div1Packed := b.EmitSelect(typeFloat, is2101010, div_2_10_10_10_x, b.EmitSelect(typeFloat, is1010102, div_2_10_10_10_x, div_2_10_10_10_x))
	div2Packed := b.EmitSelect(typeFloat, is2101010, div_2_10_10_10_x, b.EmitSelect(typeFloat, is1010102, div_2_10_10_10_x, div_2_10_10_10_x))
	div3Packed := b.EmitSelect(typeFloat, is2101010, div_2_10_10_10_w, b.EmitSelect(typeFloat, is1010102, div_2_10_10_10_x, ctx.GetConstId(ConstIdFloat1)))

	div0 := b.EmitSelect(typeFloat, isPacked10Format, div0Packed, divisor)
	div1 := b.EmitSelect(typeFloat, isPacked10Format, div1Packed, divisor)
	div2 := b.EmitSelect(typeFloat, isPacked10Format, div2Packed, divisor)
	div3 := b.EmitSelect(typeFloat, isPacked10Format, div3Packed, divisor)

	// Component scaling.
	comp0Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f0, div0))
	comp1Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f1, div1))
	comp2Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f2, div2))
	comp3Unorm := b.EmitConvertFToU(typeUint, b.EmitFMul(typeFloat, f3, div3))

	comp0Uint := b.EmitBitcast(typeUint, f0)
	comp1Uint := b.EmitBitcast(typeUint, f1)
	comp2Uint := b.EmitBitcast(typeUint, f2)
	comp3Uint := b.EmitBitcast(typeUint, f3)

	comp0Sint := b.EmitBitcast(typeUint, f0)
	comp1Sint := b.EmitBitcast(typeUint, f1)
	comp2Sint := b.EmitBitcast(typeUint, f2)
	comp3Sint := b.EmitBitcast(typeUint, f3)

	sdiv_2_10_10_10_x := b.EmitConstantFloat(typeFloat, 511.0)
	sdiv_2_10_10_10_w := b.EmitConstantFloat(typeFloat, 1.0)
	sdiv0Packed := b.EmitSelect(typeFloat, is2101010, sdiv_2_10_10_10_x, b.EmitSelect(typeFloat, is1010102, sdiv_2_10_10_10_w, sdiv_2_10_10_10_x))
	sdiv1Packed := b.EmitSelect(typeFloat, is2101010, sdiv_2_10_10_10_x, b.EmitSelect(typeFloat, is1010102, sdiv_2_10_10_10_x, sdiv_2_10_10_10_x))
	sdiv2Packed := b.EmitSelect(typeFloat, is2101010, sdiv_2_10_10_10_x, b.EmitSelect(typeFloat, is1010102, sdiv_2_10_10_10_x, sdiv_2_10_10_10_x))
	sdiv3Packed := b.EmitSelect(typeFloat, is2101010, sdiv_2_10_10_10_w, b.EmitSelect(typeFloat, is1010102, sdiv_2_10_10_10_x, ctx.GetConstId(ConstIdFloat1)))

	sdivStandard := b.EmitSelect(typeFloat, is32BitSize, b.EmitConstantFloat(typeFloat, 2147483647.0), b.EmitSelect(typeFloat, is16BitSize, b.EmitConstantFloat(typeFloat, 32767.0), b.EmitConstantFloat(typeFloat, 127.0)))
	sdiv0 := b.EmitSelect(typeFloat, isPacked10Format, sdiv0Packed, sdivStandard)
	sdiv1 := b.EmitSelect(typeFloat, isPacked10Format, sdiv1Packed, sdivStandard)
	sdiv2 := b.EmitSelect(typeFloat, isPacked10Format, sdiv2Packed, sdivStandard)
	sdiv3 := b.EmitSelect(typeFloat, isPacked10Format, sdiv3Packed, sdivStandard)

	// Snorm stores.
	comp0SnormF := b.EmitFMul(typeFloat, f0, sdiv0)
	comp1SnormF := b.EmitFMul(typeFloat, f1, sdiv1)
	comp2SnormF := b.EmitFMul(typeFloat, f2, sdiv2)
	comp3SnormF := b.EmitFMul(typeFloat, f3, sdiv3)

	comp0Snorm := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, comp0SnormF))
	comp1Snorm := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, comp1SnormF))
	comp2Snorm := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, comp2SnormF))
	comp3Snorm := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, comp3SnormF))

	comp0Uscaled := b.EmitConvertFToU(typeUint, f0)
	comp1Uscaled := b.EmitConvertFToU(typeUint, f1)
	comp2Uscaled := b.EmitConvertFToU(typeUint, f2)
	comp3Uscaled := b.EmitConvertFToU(typeUint, f3)

	comp0Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f0))
	comp1Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f1))
	comp2Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f2))
	comp3Sscaled := b.EmitBitcast(typeUint, b.EmitConvertFToS(typeInt, f3))

	packHalf := func(floatA, floatB SpirvId) SpirvId {
		vec2 := b.EmitCompositeConstruct(typeVec2, floatA, floatB)
		return b.EmitExtInst(typeUint, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpPackHalf2x16, vec2)
	}

	f32ToUfloat11 := func(floatVal SpirvId) SpirvId {
		f32Int := b.EmitBitcast(typeUint, floatVal)
		shifted := b.EmitShiftRightLogical(typeUint, f32Int, b.EmitConstantUint(typeUint, 17))
		return b.EmitBitwiseAnd(typeUint, shifted, b.EmitConstantUint(typeUint, 0x7FF))
	}
	f32ToUfloat10 := func(floatVal SpirvId) SpirvId {
		f32Int := b.EmitBitcast(typeUint, floatVal)
		shifted := b.EmitShiftRightLogical(typeUint, f32Int, b.EmitConstantUint(typeUint, 18))
		return b.EmitBitwiseAnd(typeUint, shifted, b.EmitConstantUint(typeUint, 0x3FF))
	}

	comp0Float101111 := f32ToUfloat11(f0)
	comp1Float101111 := f32ToUfloat11(f1)
	comp2Float101111 := f32ToUfloat10(f2)

	comp0Float16 := packHalf(f0, ctx.GetConstId(ConstIdFloat0))
	comp1Float16 := packHalf(f1, ctx.GetConstId(ConstIdFloat0))
	comp2Float16 := packHalf(f2, ctx.GetConstId(ConstIdFloat0))
	comp3Float16 := packHalf(f3, ctx.GetConstId(ConstIdFloat0))

	comp0Float := b.EmitSelect(typeUint, is101111, comp0Float101111, b.EmitSelect(typeUint, is16BitSize, comp0Float16, comp0Uint))
	comp1Float := b.EmitSelect(typeUint, is101111, comp1Float101111, b.EmitSelect(typeUint, is16BitSize, comp1Float16, comp1Uint))
	comp2Float := b.EmitSelect(typeUint, is101111, comp2Float101111, b.EmitSelect(typeUint, is16BitSize, comp2Float16, comp2Uint))
	comp3Float := b.EmitSelect(typeUint, is16BitSize, comp3Float16, comp3Uint)

	selectRaw := func(unorm, snorm, sint, float, uscaled, sscaled, uintVal SpirvId) SpirvId {
		res := b.EmitSelect(typeUint, isUnorm, unorm, uintVal)
		res = b.EmitSelect(typeUint, isAnySnorm, snorm, res)
		res = b.EmitSelect(typeUint, isSint, sint, res)
		res = b.EmitSelect(typeUint, isFloat, float, res)
		res = b.EmitSelect(typeUint, isUscaled, uscaled, res)
		res = b.EmitSelect(typeUint, isSscaled, sscaled, res)
		return res
	}
	raw0 := selectRaw(comp0Unorm, comp0Snorm, comp0Sint, comp0Float, comp0Uscaled, comp0Sscaled, comp0Uint)
	raw1 := selectRaw(comp1Unorm, comp1Snorm, comp1Sint, comp1Float, comp1Uscaled, comp1Sscaled, comp1Uint)
	raw2 := selectRaw(comp2Unorm, comp2Snorm, comp2Sint, comp2Float, comp2Uscaled, comp2Sscaled, comp2Uint)
	raw3 := selectRaw(comp3Unorm, comp3Snorm, comp3Sint, comp3Float, comp3Uscaled, comp3Sscaled, comp3Uint)

	raw0_8 := b.EmitBitwiseAnd(typeUint, raw0, cFF)
	raw1_8 := b.EmitBitwiseAnd(typeUint, raw1, cFF)
	raw2_8 := b.EmitBitwiseAnd(typeUint, raw2, cFF)
	raw3_8 := b.EmitBitwiseAnd(typeUint, raw3, cFF)

	raw0_16 := b.EmitBitwiseAnd(typeUint, raw0, cFFFF)
	raw1_16 := b.EmitBitwiseAnd(typeUint, raw1, cFFFF)
	raw2_16 := b.EmitBitwiseAnd(typeUint, raw2, cFFFF)
	raw3_16 := b.EmitBitwiseAnd(typeUint, raw3, cFFFF)

	// Bitfield packing & selection.
	// Packed format insertion.
	word0_2_10_10_10 := b.EmitBitFieldInsert(typeUint, raw0, raw1, b.EmitConstantUint(typeUint, 10), b.EmitConstantUint(typeUint, 10))
	word0_2_10_10_10 = b.EmitBitFieldInsert(typeUint, word0_2_10_10_10, raw2, b.EmitConstantUint(typeUint, 20), b.EmitConstantUint(typeUint, 10))
	word0_2_10_10_10 = b.EmitBitFieldInsert(typeUint, word0_2_10_10_10, raw3, b.EmitConstantUint(typeUint, 30), b.EmitConstantUint(typeUint, 2))

	word0_10_10_10_2 := b.EmitBitFieldInsert(typeUint, raw0, raw1, b.EmitConstantUint(typeUint, 2), b.EmitConstantUint(typeUint, 10))
	word0_10_10_10_2 = b.EmitBitFieldInsert(typeUint, word0_10_10_10_2, raw2, b.EmitConstantUint(typeUint, 12), b.EmitConstantUint(typeUint, 10))
	word0_10_10_10_2 = b.EmitBitFieldInsert(typeUint, word0_10_10_10_2, raw3, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))

	word0_10_11_11 := b.EmitBitFieldInsert(typeUint, raw0, raw1, b.EmitConstantUint(typeUint, 11), b.EmitConstantUint(typeUint, 11))
	word0_10_11_11 = b.EmitBitFieldInsert(typeUint, word0_10_11_11, raw2, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))

	word0Packed := b.EmitSelect(typeUint, is2101010, word0_2_10_10_10, b.EmitSelect(typeUint, is1010102, word0_10_10_10_2, word0_10_11_11))

	word0_88 := b.EmitBitwiseOr(typeUint, raw0_8, b.EmitShiftLeftLogical(typeUint, raw1_8, ctx.GetConstId(ConstIdUint8)))
	word0_1616 := b.EmitBitwiseOr(typeUint, raw0_16, b.EmitShiftLeftLogical(typeUint, raw1_16, ctx.GetConstId(ConstIdUint16)))
	word0_8888_lo := b.EmitBitwiseOr(typeUint, raw0_8, b.EmitShiftLeftLogical(typeUint, raw1_8, ctx.GetConstId(ConstIdUint8)))
	word0_8888_hi := b.EmitBitwiseOr(typeUint, b.EmitShiftLeftLogical(typeUint, raw2_8, ctx.GetConstId(ConstIdUint16)), b.EmitShiftLeftLogical(typeUint, raw3_8, ctx.GetConstId(ConstIdUint24)))
	word0_8888 := b.EmitBitwiseOr(typeUint, word0_8888_lo, word0_8888_hi)

	word0 := b.EmitSelect(typeUint, isPacked10Format, word0Packed, b.EmitSelect(typeUint, is32BitSize, raw0,
		b.EmitSelect(typeUint, is16BitSize, b.EmitSelect(typeUint, is1616, word0_1616, b.EmitSelect(typeUint, is16161616, word0_1616, raw0_16)),
			b.EmitSelect(typeUint, is88, word0_88, b.EmitSelect(typeUint, is8888, word0_8888, raw0_8)))))

	word1_16161616 := b.EmitBitwiseOr(typeUint, raw2_16, b.EmitShiftLeftLogical(typeUint, raw3_16, ctx.GetConstId(ConstIdUint16)))
	word1 := b.EmitSelect(typeUint, is16161616, word1_16161616, raw1)
	word2 := raw2
	word3 := raw3

	// Store memory values.
	storeWord0 := func(val SpirvId) {
		is8BitMem := is8
		is16BitMem := b.EmitLogicalOr(typeBool, is16, is88)

		label8Bit := b.AllocId()
		labelCheck16 := b.AllocId()
		label16Bit := b.AllocId()
		label32Bit := b.AllocId()
		labelMerge16 := b.AllocId()
		labelMergeTotal := b.AllocId()

		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(ConstId64Uint0))

		b.EmitSelectionMerge(labelMergeTotal, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(is8BitMem, label8Bit, labelCheck16)

		b.EmitLabel(label8Bit)
		ptr8 := b.EmitConvertUToPtr(typePtrPsbUint8, addr)
		val8 := b.EmitUConvert(typeUint8, val)
		b.EmitStore(ptr8, val8, spec.SpvMemoryAccessAligned, 1)
		b.EmitBranch(labelMergeTotal)

		b.EmitLabel(labelCheck16)
		b.EmitSelectionMerge(labelMerge16, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(is16BitMem, label16Bit, label32Bit)

		b.EmitLabel(label16Bit)
		ptr16 := b.EmitConvertUToPtr(typePtrPsbUint16, addr)
		val16 := b.EmitUConvert(typeUint16, val)
		b.EmitStore(ptr16, val16, spec.SpvMemoryAccessAligned, 2)
		b.EmitBranch(labelMerge16)

		b.EmitLabel(label32Bit)
		ptr32 := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		b.EmitStore(ptr32, val, spec.SpvMemoryAccessAligned, 4)
		b.EmitBranch(labelMerge16)

		b.EmitLabel(labelMerge16)
		b.EmitBranch(labelMergeTotal)
		b.EmitLabel(labelMergeTotal)
	}

	storeWord := func(offset SpirvId, val SpirvId) {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(offset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		b.EmitStore(ptr, val, spec.SpvMemoryAccessAligned, 4)
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
