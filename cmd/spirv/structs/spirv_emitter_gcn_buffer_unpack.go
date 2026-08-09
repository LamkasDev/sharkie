package structs

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitFormatUnpackHelper(b *SpvBuilder, ctx *SpirvBlockContext, absoluteAddress, dataFormat, numFormat, outOfRange SpirvId, op uint32) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint8 := ctx.GetId(BlockContextIdTypeUint8)
	typeUint16 := ctx.GetId(BlockContextIdTypeUint16)
	typeInt := ctx.GetId(BlockContextIdTypeInt)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec4 := ctx.GetId(BlockContextIdTypeV4Float)
	typeVec2 := ctx.GetId(BlockContextIdTypeV2Float)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typePtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	typePtrPsbUint8 := ctx.GetId(BlockContextIdPtrPsbUint8)
	typePtrPsbUint16 := ctx.GetId(BlockContextIdPtrPsbUint16)
	cFF := ctx.GetConstId(ConstIdUint255)
	cFFFF := ctx.GetConstId(ConstIdUintFFFF)

	translatedAddress := ctx.TranslateAddress(b, absoluteAddress)
	inRange := b.EmitLogicalNot(typeBool, outOfRange)

	isTypedLoad := false
	switch op {
	case gcnSpec.MubufOpLoadFormatX, gcnSpec.MubufOpLoadFormatXy, gcnSpec.MubufOpLoadFormatXyz, gcnSpec.MubufOpLoadFormatXyzw:
		isTypedLoad = true
	}

	needsWord1 := ctx.GetId(BlockContextIdFalse)
	needsWord2 := ctx.GetId(BlockContextIdFalse)
	needsWord3 := ctx.GetId(BlockContextIdFalse)

	is8BitMem := ctx.GetId(BlockContextIdFalse)
	is16BitMem := ctx.GetId(BlockContextIdFalse)
	is32BitMem := ctx.GetId(BlockContextIdTrue)

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

	if isTypedLoad {
		is8 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint1))
		is16 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint2))
		is88 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint3))
		is32 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint4))
		is1616 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint5))
		is101111 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint6))
		is1010102 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint8))
		is2101010 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint9))
		is8888 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint10))
		is3232 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint11))
		is16161616 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint12))
		is323232 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint13))
		is32323232 = b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint14))

		is8BitMem = is8
		is16BitMem = b.EmitLogicalOr(typeBool, is16, is88)
		is32BitMem = b.EmitLogicalNot(typeBool, b.EmitLogicalOr(typeBool, is8BitMem, is16BitMem))

		needsWord1 = b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is3232, is16161616), b.EmitLogicalOr(typeBool, is323232, is32323232))
		needsWord2 = b.EmitLogicalOr(typeBool, is323232, is32323232)
		needsWord3 = is32323232
	} else {
		switch op {
		case gcnSpec.MubufOpLoadDwordx2:
			needsWord1 = ctx.GetId(BlockContextIdTrue)
		case gcnSpec.MubufOpLoadDwordx3:
			needsWord1 = ctx.GetId(BlockContextIdTrue)
			needsWord2 = ctx.GetId(BlockContextIdTrue)
		case gcnSpec.MubufOpLoadDwordx4:
			needsWord1 = ctx.GetId(BlockContextIdTrue)
			needsWord2 = ctx.GetId(BlockContextIdTrue)
			needsWord3 = ctx.GetId(BlockContextIdTrue)
		}
	}

	// Load memory values.
	const0_8 := b.EmitConstantUint(typeUint8, 0)
	const0_16 := b.EmitConstantUint(typeUint16, 0)

	load8 := func(offset SpirvId, cond SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(offset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint8, addr)
		val8 := b.EmitLoadConditional(typeUint8, ptr, cond, const0_8, spec.SpvMemoryAccessAligned, 1)
		return b.EmitUConvert(typeUint, val8)
	}
	load16 := func(offset SpirvId, cond SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(offset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint16, addr)
		val16 := b.EmitLoadConditional(typeUint16, ptr, cond, const0_16, spec.SpvMemoryAccessAligned, 2)
		return b.EmitUConvert(typeUint, val16)
	}
	load32 := func(offset SpirvId, cond SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(offset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		return b.EmitLoadConditional(typeUint, ptr, cond, ctx.GetConstId(ConstIdUint0), spec.SpvMemoryAccessAligned, 4)
	}

	cond8 := b.EmitLogicalAnd(typeBool, inRange, is8BitMem)
	cond16 := b.EmitLogicalAnd(typeBool, inRange, is16BitMem)
	cond32 := b.EmitLogicalAnd(typeBool, inRange, is32BitMem)

	val8 := load8(ConstId64Uint0, cond8)
	val16 := load16(ConstId64Uint0, cond16)
	val32 := load32(ConstId64Uint0, cond32)

	word0 := b.EmitSelect(typeUint, is32BitMem, val32, b.EmitSelect(typeUint, is16BitMem, val16, val8))
	word1 := load32(ConstId64Uint4, b.EmitLogicalAnd(typeBool, inRange, needsWord1))
	word2 := load32(ConstId64Uint8, b.EmitLogicalAnd(typeBool, inRange, needsWord2))
	word3 := load32(ConstId64Uint12, b.EmitLogicalAnd(typeBool, inRange, needsWord3))

	f0 := b.EmitBitcast(typeFloat, word0)
	f1 := b.EmitBitcast(typeFloat, word1)
	f2 := b.EmitBitcast(typeFloat, word2)
	f3 := b.EmitBitcast(typeFloat, word3)

	if isTypedLoad {
		// Format conversions & unpacking.
		isUnorm := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint0))
		isSnorm := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint1))
		isUscaled := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint2))
		isSscaled := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint3))
		isUint := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint4))
		isSint := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint5))
		isSnormOgl := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint6))
		isFloat := b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint7))
		isAnyNorm := b.EmitLogicalOr(typeBool, isUnorm, b.EmitLogicalOr(typeBool, isSnorm, isSnormOgl))
		isAnyIntOrNorm := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAnyNorm, isUint), b.EmitLogicalOr(typeBool, isUscaled, isSscaled))
		isAnyIntOrNorm = b.EmitLogicalOr(typeBool, isAnyIntOrNorm, b.EmitLogicalOr(typeBool, isSint, isFloat))

		is8Format := b.EmitLogicalAnd(typeBool, is8, isAnyIntOrNorm)
		is16Format := b.EmitLogicalAnd(typeBool, is16, isAnyIntOrNorm)
		is88Format := b.EmitLogicalAnd(typeBool, is88, isAnyIntOrNorm)
		is32Format := b.EmitLogicalAnd(typeBool, is32, isAnyIntOrNorm)
		is1616Format := b.EmitLogicalAnd(typeBool, is1616, isAnyIntOrNorm)
		is8888Format := b.EmitLogicalAnd(typeBool, is8888, isAnyIntOrNorm)
		is3232Format := b.EmitLogicalAnd(typeBool, is3232, isAnyIntOrNorm)
		is16161616Format := b.EmitLogicalAnd(typeBool, is16161616, isAnyIntOrNorm)
		is323232Format := b.EmitLogicalAnd(typeBool, is323232, isAnyIntOrNorm)
		is32323232Format := b.EmitLogicalAnd(typeBool, is32323232, isAnyIntOrNorm)
		isPacked10Format := b.EmitLogicalAnd(typeBool, b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is101111, is1010102), is2101010), isAnyIntOrNorm)

		is16BitSize := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is16Format, is1616Format), is16161616Format)
		is32BitSize := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is32Format, is3232Format), b.EmitLogicalOr(typeBool, is323232Format, is32323232Format))

		// Standard byte and short extraction.
		comp0Byte := b.EmitBitwiseAnd(typeUint, word0, cFF)
		comp1Byte := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint8)), cFF)
		comp2Byte := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint16)), cFF)
		comp3Byte := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint24)), cFF)

		comp0Short := b.EmitBitwiseAnd(typeUint, word0, cFFFF)
		comp1Short := b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint16))
		comp2Short := b.EmitBitwiseAnd(typeUint, word1, cFFFF)
		comp3Short := b.EmitShiftRightLogical(typeUint, word1, ctx.GetConstId(ConstIdUint16))

		// Bitfield extraction for packed formats.
		// 2_10_10_10; X(10), Y(10), Z(10), W(2).
		comp0_2_10_10_10 := b.EmitBitFieldUExtract(typeUint, word0, ctx.GetConstId(ConstIdUint0), b.EmitConstantUint(typeUint, 10))
		comp1_2_10_10_10 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 10), b.EmitConstantUint(typeUint, 10))
		comp2_2_10_10_10 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 20), b.EmitConstantUint(typeUint, 10))
		comp3_2_10_10_10 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 30), b.EmitConstantUint(typeUint, 2))

		// 10_10_10_2; X(2), Y(10), Z(10), W(10).
		comp0_10_10_10_2 := b.EmitBitFieldUExtract(typeUint, word0, ctx.GetConstId(ConstIdUint0), b.EmitConstantUint(typeUint, 2))
		comp1_10_10_10_2 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 2), b.EmitConstantUint(typeUint, 10))
		comp2_10_10_10_2 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 12), b.EmitConstantUint(typeUint, 10))
		comp3_10_10_10_2 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))

		// 10_11_11; X(11), Y(11), Z(10).
		comp0_10_11_11 := b.EmitBitFieldUExtract(typeUint, word0, ctx.GetConstId(ConstIdUint0), b.EmitConstantUint(typeUint, 11))
		comp1_10_11_11 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 11), b.EmitConstantUint(typeUint, 11))
		comp2_10_11_11 := b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))

		// Select packed components.
		comp0Packed := b.EmitSelect(typeUint, is2101010, comp0_2_10_10_10, b.EmitSelect(typeUint, is1010102, comp0_10_10_10_2, comp0_10_11_11))
		comp1Packed := b.EmitSelect(typeUint, is2101010, comp1_2_10_10_10, b.EmitSelect(typeUint, is1010102, comp1_10_10_10_2, comp1_10_11_11))
		comp2Packed := b.EmitSelect(typeUint, is2101010, comp2_2_10_10_10, b.EmitSelect(typeUint, is1010102, comp2_10_10_10_2, comp2_10_11_11))
		comp3Packed := b.EmitSelect(typeUint, is2101010, comp3_2_10_10_10, b.EmitSelect(typeUint, is1010102, comp3_10_10_10_2, ctx.GetConstId(ConstIdUint0)))

		raw0 := b.EmitSelect(typeUint, isPacked10Format, comp0Packed, b.EmitSelect(typeUint, is32BitSize, word0, b.EmitSelect(typeUint, is16BitSize, comp0Short, comp0Byte)))
		raw1 := b.EmitSelect(typeUint, isPacked10Format, comp1Packed, b.EmitSelect(typeUint, is32BitSize, word1, b.EmitSelect(typeUint, is16BitSize, comp1Short, comp1Byte)))
		raw2 := b.EmitSelect(typeUint, isPacked10Format, comp2Packed, b.EmitSelect(typeUint, is32BitSize, word2, b.EmitSelect(typeUint, is16BitSize, comp2Short, comp2Byte)))
		raw3 := b.EmitSelect(typeUint, isPacked10Format, comp3Packed, b.EmitSelect(typeUint, is32BitSize, word3, b.EmitSelect(typeUint, is16BitSize, comp3Short, comp3Byte)))

		// Divisor for UNORM.
		divisor := b.EmitSelect(typeFloat, is32BitSize, ctx.GetConstId(ConstIdFloat1), b.EmitSelect(typeFloat, is16BitSize, ctx.GetConstId(ConstIdFloat65535), ctx.GetConstId(ConstIdFloat255)))

		// Adjust divisor for packed formats (unorm max values).
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

		comp0Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw0), div0)
		comp1Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw1), div1)
		comp2Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw2), div2)
		comp3Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw3), div3)

		comp0UintF := b.EmitBitcast(typeFloat, raw0)
		comp1UintF := b.EmitBitcast(typeFloat, raw1)
		comp2UintF := b.EmitBitcast(typeFloat, raw2)
		comp3UintF := b.EmitBitcast(typeFloat, raw3)

		extendShift0 := b.EmitSelect(typeUint, isPacked10Format, b.EmitSelect(typeUint, is2101010, b.EmitConstantUint(typeUint, 22), b.EmitSelect(typeUint, is1010102, b.EmitConstantUint(typeUint, 30), b.EmitConstantUint(typeUint, 21))), b.EmitSelect(typeUint, is16BitSize, ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint24)))
		extendShift1 := b.EmitSelect(typeUint, isPacked10Format, b.EmitSelect(typeUint, is2101010, b.EmitConstantUint(typeUint, 22), b.EmitSelect(typeUint, is1010102, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 21))), b.EmitSelect(typeUint, is16BitSize, ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint24)))
		extendShift2 := b.EmitSelect(typeUint, isPacked10Format, b.EmitSelect(typeUint, is2101010, b.EmitConstantUint(typeUint, 22), b.EmitSelect(typeUint, is1010102, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 22))), b.EmitSelect(typeUint, is16BitSize, ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint24)))
		extendShift3 := b.EmitSelect(typeUint, isPacked10Format, b.EmitSelect(typeUint, is2101010, b.EmitConstantUint(typeUint, 30), b.EmitSelect(typeUint, is1010102, b.EmitConstantUint(typeUint, 22), ctx.GetConstId(ConstIdUint0))), b.EmitSelect(typeUint, is16BitSize, ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint24)))

		makeSigned := func(raw, shift SpirvId) SpirvId {
			shiftedLeft := b.EmitShiftLeftLogical(typeUint, raw, shift)
			intShifted := b.EmitBitcast(typeInt, shiftedLeft)
			return b.EmitShiftRightArithmetic(typeInt, intShifted, shift)
		}

		raw0Signed := makeSigned(raw0, extendShift0)
		raw1Signed := makeSigned(raw1, extendShift1)
		raw2Signed := makeSigned(raw2, extendShift2)
		raw3Signed := makeSigned(raw3, extendShift3)

		comp0SintF := b.EmitBitcast(typeFloat, raw0Signed)
		comp1SintF := b.EmitBitcast(typeFloat, raw1Signed)
		comp2SintF := b.EmitBitcast(typeFloat, raw2Signed)
		comp3SintF := b.EmitBitcast(typeFloat, raw3Signed)

		// Snorm divisors.
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

		comp0Snorm := b.EmitFDiv(typeFloat, b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw0Signed)), sdiv0)
		comp1Snorm := b.EmitFDiv(typeFloat, b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw1Signed)), sdiv1)
		comp2Snorm := b.EmitFDiv(typeFloat, b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw2Signed)), sdiv2)
		comp3Snorm := b.EmitFDiv(typeFloat, b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw3Signed)), sdiv3)

		// Snorm clamp; max(-1.0, val).
		negOne := b.EmitConstantFloat(typeFloat, -1.0)
		clampSnorm := func(val SpirvId) SpirvId {
			return b.EmitExtInst(typeFloat, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFMax, negOne, val)
		}
		comp0Snorm = clampSnorm(comp0Snorm)
		comp1Snorm = clampSnorm(comp1Snorm)
		comp2Snorm = clampSnorm(comp2Snorm)
		comp3Snorm = clampSnorm(comp3Snorm)

		// SnormOgl.
		makeSnormOgl := func(raw, div SpirvId) SpirvId {
			fRaw := b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw))
			num := b.EmitFAdd(typeFloat, b.EmitFMul(typeFloat, fRaw, ctx.GetConstId(ConstIdFloat2)), ctx.GetConstId(ConstIdFloat1))
			return b.EmitFDiv(typeFloat, num, div)
		}
		comp0SnormOgl := makeSnormOgl(raw0Signed, div0)
		comp1SnormOgl := makeSnormOgl(raw1Signed, div1)
		comp2SnormOgl := makeSnormOgl(raw2Signed, div2)
		comp3SnormOgl := makeSnormOgl(raw3Signed, div3)

		comp0Uscaled := b.EmitConvertUToF(typeFloat, raw0)
		comp1Uscaled := b.EmitConvertUToF(typeFloat, raw1)
		comp2Uscaled := b.EmitConvertUToF(typeFloat, raw2)
		comp3Uscaled := b.EmitConvertUToF(typeFloat, raw3)

		comp0Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw0Signed))
		comp1Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw1Signed))
		comp2Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw2Signed))
		comp3Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(typeInt, raw3Signed))

		unpackHalf := func(raw SpirvId) SpirvId {
			vec2 := b.EmitExtInst(typeVec2, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpUnpackHalf2x16, raw)
			return b.EmitCompositeExtract(typeFloat, vec2, 0)
		}

		ufloat11ToF32 := func(raw SpirvId) SpirvId {
			// uf11_to_f32 logic; (val << 17) float exponent adjustment.
			shifted := b.EmitShiftLeftLogical(typeUint, raw, b.EmitConstantUint(typeUint, 17))

			// If exp is 0, it's a denormal. For 10_11_11, we handle it.
			// Let's just bitcast to float and multiply by 2^-112 or similar.
			expExt := b.EmitBitFieldUExtract(typeUint, raw, b.EmitConstantUint(typeUint, 6), b.EmitConstantUint(typeUint, 5))
			isZeroExp := b.EmitIEqual(typeBool, expExt, ctx.GetConstId(ConstIdUint0))

			// Magic float multiplier approach.
			magic := b.EmitBitwiseOr(typeUint, shifted, b.EmitConstantUint(typeUint, 0x38000000))
			fMagic := b.EmitFSub(typeFloat, b.EmitBitcast(typeFloat, magic), b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0x38000000)))

			// Normal float approach.
			normalExp := b.EmitIAdd(typeUint, expExt, b.EmitConstantUint(typeUint, 112))
			normalInserted := b.EmitBitFieldInsert(typeUint, shifted, normalExp, b.EmitConstantUint(typeUint, 23), b.EmitConstantUint(typeUint, 8))
			fNormal := b.EmitBitcast(typeFloat, normalInserted)

			return b.EmitSelect(typeFloat, isZeroExp, fMagic, fNormal)
		}

		ufloat10ToF32 := func(raw SpirvId) SpirvId {
			shifted := b.EmitShiftLeftLogical(typeUint, raw, b.EmitConstantUint(typeUint, 18))
			expExt := b.EmitBitFieldUExtract(typeUint, raw, b.EmitConstantUint(typeUint, 5), b.EmitConstantUint(typeUint, 5))
			isZeroExp := b.EmitIEqual(typeBool, expExt, ctx.GetConstId(ConstIdUint0))

			magic := b.EmitBitwiseOr(typeUint, shifted, b.EmitConstantUint(typeUint, 0x38000000))
			fMagic := b.EmitFSub(typeFloat, b.EmitBitcast(typeFloat, magic), b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0x38000000)))

			normalExp := b.EmitIAdd(typeUint, expExt, b.EmitConstantUint(typeUint, 112))
			normalInserted := b.EmitBitFieldInsert(typeUint, shifted, normalExp, b.EmitConstantUint(typeUint, 23), b.EmitConstantUint(typeUint, 8))
			fNormal := b.EmitBitcast(typeFloat, normalInserted)

			return b.EmitSelect(typeFloat, isZeroExp, fMagic, fNormal)
		}

		comp0Float101111 := ufloat11ToF32(raw0)
		comp1Float101111 := ufloat11ToF32(raw1)
		comp2Float101111 := ufloat10ToF32(raw2)

		comp0Float16 := unpackHalf(raw0)
		comp1Float16 := unpackHalf(raw1)
		comp2Float16 := unpackHalf(raw2)
		comp3Float16 := unpackHalf(raw3)

		comp0Float := b.EmitSelect(typeFloat, is101111, comp0Float101111, b.EmitSelect(typeFloat, is16BitSize, comp0Float16, comp0UintF))
		comp1Float := b.EmitSelect(typeFloat, is101111, comp1Float101111, b.EmitSelect(typeFloat, is16BitSize, comp1Float16, comp1UintF))
		comp2Float := b.EmitSelect(typeFloat, is101111, comp2Float101111, b.EmitSelect(typeFloat, is16BitSize, comp2Float16, comp2UintF))
		comp3Float := b.EmitSelect(typeFloat, is16BitSize, comp3Float16, comp3UintF)

		// Component selection based on numeric type.
		selectComp := func(compUnorm, compSnorm, compSnormOgl, compSint, compFloat, compUscaled, compSscaled, compUint SpirvId) SpirvId {
			comp := b.EmitSelect(typeFloat, isUnorm, compUnorm, compUint)
			comp = b.EmitSelect(typeFloat, isSnorm, compSnorm, comp)
			comp = b.EmitSelect(typeFloat, isSnormOgl, compSnormOgl, comp)
			comp = b.EmitSelect(typeFloat, isSint, compSint, comp)
			comp = b.EmitSelect(typeFloat, isFloat, compFloat, comp)
			comp = b.EmitSelect(typeFloat, isUscaled, compUscaled, comp)
			comp = b.EmitSelect(typeFloat, isSscaled, compSscaled, comp)
			return comp
		}

		comp0 := selectComp(comp0Unorm, comp0Snorm, comp0SnormOgl, comp0SintF, comp0Float, comp0Uscaled, comp0Sscaled, comp0UintF)
		comp1 := selectComp(comp1Unorm, comp1Snorm, comp1SnormOgl, comp1SintF, comp1Float, comp1Uscaled, comp1Sscaled, comp1UintF)
		comp2 := selectComp(comp2Unorm, comp2Snorm, comp2SnormOgl, comp2SintF, comp2Float, comp2Uscaled, comp2Sscaled, comp2UintF)
		comp3 := selectComp(comp3Unorm, comp3Snorm, comp3SnormOgl, comp3SintF, comp3Float, comp3Uscaled, comp3Sscaled, comp3UintF)

		isAny1Comp := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is8Format, is16Format), is32Format)
		isAny2Comp := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is88Format, is1616Format), is3232Format)
		isAny3Comp := b.EmitLogicalOr(typeBool, is323232Format, is101111)
		isAny4Comp := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is8888Format, is16161616Format), b.EmitLogicalOr(typeBool, is32323232Format, b.EmitLogicalOr(typeBool, is1010102, is2101010)))
		isAny := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), isAny3Comp), isAny4Comp)

		f0 = b.EmitSelect(typeFloat, isAny, comp0, f0)
		f1 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAny2Comp, isAny3Comp), isAny4Comp), comp1, f1)
		f2 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, isAny3Comp, isAny4Comp), comp2, f2)
		f3 = b.EmitSelect(typeFloat, isAny4Comp, comp3, f3)

		f1 = b.EmitSelect(typeFloat, isAny1Comp, ctx.GetConstId(ConstIdFloat0), f1)
		f2 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), ctx.GetConstId(ConstIdFloat0), f2)
		f3 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), isAny3Comp), ctx.GetConstId(ConstIdFloat1), f3)
	}

	return b.EmitCompositeConstruct(typeVec4, f0, f1, f2, f3)
}
