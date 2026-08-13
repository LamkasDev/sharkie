package structs

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitFormatPackHelper(b *SpvBuilder, ctx *SpirvBlockContext, bindingIndex uint32, offset SpirvId, dataFormat, numFormat uint32, op uint32, storeVec4 SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint8 := ctx.GetId(BlockContextIdTypeUint8)
	typeUint16 := ctx.GetId(BlockContextIdTypeUint16)
	typeInt := ctx.GetId(BlockContextIdTypeInt)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec2 := ctx.GetId(BlockContextIdTypeV2Float)
	typePtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	typePtrPsbUint8 := ctx.GetId(BlockContextIdPtrPsbUint8)
	typePtrPsbUint16 := ctx.GetId(BlockContextIdPtrPsbUint16)
	cFF := ctx.GetConstId(ConstIdUint255)
	cFFFF := ctx.GetConstId(ConstIdUintFFFF)

	translatedAddress := ctx.TranslateAddress(b, bindingIndex, offset)

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

	is8 := false
	is16 := false
	is88 := false
	is32 := false
	is1616 := false
	is101111 := false
	is1010102 := false
	is2101010 := false
	is8888 := false
	is3232 := false
	is16161616 := false
	is323232 := false
	is32323232 := false

	is16BitSize := false
	is32BitSize := true
	isPacked10Format := false

	isUnorm := false
	isSnorm := false
	isUscaled := false
	isSscaled := false
	isUintF := false
	isSintF := false
	isSnormOgl := false
	isFloatF := false
	isAnySnorm := false

	if isTypedStore {
		isUnorm = numFormat == 0
		isSnorm = numFormat == 1
		isUscaled = numFormat == 2
		isSscaled = numFormat == 3
		isUintF = numFormat == 4
		isSintF = numFormat == 5
		isSnormOgl = numFormat == 6
		isFloatF = numFormat == 7
		isAnySnorm = isSnorm || isSnormOgl

		isAnyNorm := isUnorm || isAnySnorm
		isAnyIntOrNorm := isAnyNorm || isUintF || isUscaled || isSscaled || isSintF || isFloatF

		is8 = dataFormat == 1 && isAnyIntOrNorm
		is16 = dataFormat == 2 && isAnyIntOrNorm
		is88 = dataFormat == 3 && isAnyIntOrNorm
		is32 = dataFormat == 4 && isAnyIntOrNorm
		is1616 = dataFormat == 5 && isAnyIntOrNorm
		is101111 = dataFormat == 6 && isAnyIntOrNorm
		is1010102 = dataFormat == 8 && isAnyIntOrNorm
		is2101010 = dataFormat == 9 && isAnyIntOrNorm
		is8888 = dataFormat == 10 && isAnyIntOrNorm
		is3232 = dataFormat == 11 && isAnyIntOrNorm
		is16161616 = dataFormat == 12 && isAnyIntOrNorm
		is323232 = dataFormat == 13 && isAnyIntOrNorm
		is32323232 = dataFormat == 14 && isAnyIntOrNorm

		isPacked10Format = is101111 || is1010102 || is2101010
		is16BitSize = is16 || is1616 || is16161616
		is32BitSize = is32 || is3232 || is323232 || is32323232
	}

	var divisor SpirvId
	if is32BitSize {
		divisor = ctx.GetConstId(ConstIdFloat1)
	} else if is16BitSize {
		divisor = ctx.GetConstId(ConstIdFloat65535)
	} else {
		divisor = ctx.GetConstId(ConstIdFloat255)
	}

	var div0, div1, div2, div3 SpirvId
	if isPacked10Format {
		div_2_10_10_10_x := b.EmitConstantFloat(typeFloat, 1023.0)
		div_2_10_10_10_w := b.EmitConstantFloat(typeFloat, 3.0)
		if is2101010 {
			div0, div1, div2, div3 = div_2_10_10_10_x, div_2_10_10_10_x, div_2_10_10_10_x, div_2_10_10_10_w
		} else if is1010102 {
			div0, div1, div2, div3 = div_2_10_10_10_w, div_2_10_10_10_x, div_2_10_10_10_x, div_2_10_10_10_x
		} else {
			div0, div1, div2, div3 = div_2_10_10_10_x, div_2_10_10_10_x, div_2_10_10_10_x, ctx.GetConstId(ConstIdFloat1)
		}
	} else {
		div0, div1, div2, div3 = divisor, divisor, divisor, divisor
	}

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

	var sdiv0, sdiv1, sdiv2, sdiv3 SpirvId
	if isPacked10Format {
		sdiv_2_10_10_10_x := b.EmitConstantFloat(typeFloat, 511.0)
		sdiv_2_10_10_10_w := b.EmitConstantFloat(typeFloat, 1.0)
		if is2101010 {
			sdiv0, sdiv1, sdiv2, sdiv3 = sdiv_2_10_10_10_x, sdiv_2_10_10_10_x, sdiv_2_10_10_10_x, sdiv_2_10_10_10_w
		} else if is1010102 {
			sdiv0, sdiv1, sdiv2, sdiv3 = sdiv_2_10_10_10_w, sdiv_2_10_10_10_x, sdiv_2_10_10_10_x, sdiv_2_10_10_10_x
		} else {
			sdiv0, sdiv1, sdiv2, sdiv3 = sdiv_2_10_10_10_x, sdiv_2_10_10_10_x, sdiv_2_10_10_10_x, ctx.GetConstId(ConstIdFloat1)
		}
	} else {
		var sdivStandard SpirvId
		if is32BitSize {
			sdivStandard = b.EmitConstantFloat(typeFloat, 2147483647.0)
		} else if is16BitSize {
			sdivStandard = b.EmitConstantFloat(typeFloat, 32767.0)
		} else {
			sdivStandard = b.EmitConstantFloat(typeFloat, 127.0)
		}
		sdiv0, sdiv1, sdiv2, sdiv3 = sdivStandard, sdivStandard, sdivStandard, sdivStandard
	}

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

	var comp0Float101111, comp1Float101111, comp2Float101111 SpirvId
	if is101111 {
		comp0Float101111 = f32ToUfloat11(f0)
		comp1Float101111 = f32ToUfloat11(f1)
		comp2Float101111 = f32ToUfloat10(f2)
	}

	var comp0Float16, comp1Float16, comp2Float16, comp3Float16 SpirvId
	if is16BitSize {
		comp0Float16 = packHalf(f0, ctx.GetConstId(ConstIdFloat0))
		comp1Float16 = packHalf(f1, ctx.GetConstId(ConstIdFloat0))
		comp2Float16 = packHalf(f2, ctx.GetConstId(ConstIdFloat0))
		comp3Float16 = packHalf(f3, ctx.GetConstId(ConstIdFloat0))
	}

	var comp0Float, comp1Float, comp2Float, comp3Float SpirvId
	if is101111 {
		comp0Float = comp0Float101111
		comp1Float = comp1Float101111
		comp2Float = comp2Float101111
		comp3Float = comp3Uint
	} else if is16BitSize {
		comp0Float = comp0Float16
		comp1Float = comp1Float16
		comp2Float = comp2Float16
		comp3Float = comp3Float16
	} else {
		comp0Float = comp0Uint
		comp1Float = comp1Uint
		comp2Float = comp2Uint
		comp3Float = comp3Uint
	}

	var raw0, raw1, raw2, raw3 SpirvId
	if isUnorm {
		raw0, raw1, raw2, raw3 = comp0Unorm, comp1Unorm, comp2Unorm, comp3Unorm
	} else if isAnySnorm {
		raw0, raw1, raw2, raw3 = comp0Snorm, comp1Snorm, comp2Snorm, comp3Snorm
	} else if isSintF {
		raw0, raw1, raw2, raw3 = comp0Sint, comp1Sint, comp2Sint, comp3Sint
	} else if isFloatF {
		raw0, raw1, raw2, raw3 = comp0Float, comp1Float, comp2Float, comp3Float
	} else if isUscaled {
		raw0, raw1, raw2, raw3 = comp0Uscaled, comp1Uscaled, comp2Uscaled, comp3Uscaled
	} else if isSscaled {
		raw0, raw1, raw2, raw3 = comp0Sscaled, comp1Sscaled, comp2Sscaled, comp3Sscaled
	} else {
		raw0, raw1, raw2, raw3 = comp0Uint, comp1Uint, comp2Uint, comp3Uint
	}

	raw0_8 := b.EmitBitwiseAnd(typeUint, raw0, cFF)
	raw1_8 := b.EmitBitwiseAnd(typeUint, raw1, cFF)
	raw2_8 := b.EmitBitwiseAnd(typeUint, raw2, cFF)
	raw3_8 := b.EmitBitwiseAnd(typeUint, raw3, cFF)

	raw0_16 := b.EmitBitwiseAnd(typeUint, raw0, cFFFF)
	raw1_16 := b.EmitBitwiseAnd(typeUint, raw1, cFFFF)
	raw2_16 := b.EmitBitwiseAnd(typeUint, raw2, cFFFF)
	raw3_16 := b.EmitBitwiseAnd(typeUint, raw3, cFFFF)

	var word0Packed SpirvId
	if isPacked10Format {
		if is2101010 {
			word0_2_10_10_10 := b.EmitBitFieldInsert(typeUint, raw0, raw1, b.EmitConstantUint(typeUint, 10), b.EmitConstantUint(typeUint, 10))
			word0_2_10_10_10 = b.EmitBitFieldInsert(typeUint, word0_2_10_10_10, raw2, b.EmitConstantUint(typeUint, 20), b.EmitConstantUint(typeUint, 10))
			word0Packed = b.EmitBitFieldInsert(typeUint, word0_2_10_10_10, raw3, b.EmitConstantUint(typeUint, 30), b.EmitConstantUint(typeUint, 2))
		} else if is1010102 {
			word0_10_10_10_2 := b.EmitBitFieldInsert(typeUint, raw0, raw1, b.EmitConstantUint(typeUint, 2), b.EmitConstantUint(typeUint, 10))
			word0_10_10_10_2 = b.EmitBitFieldInsert(typeUint, word0_10_10_10_2, raw2, b.EmitConstantUint(typeUint, 12), b.EmitConstantUint(typeUint, 10))
			word0Packed = b.EmitBitFieldInsert(typeUint, word0_10_10_10_2, raw3, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))
		} else if is101111 {
			word0_10_11_11 := b.EmitBitFieldInsert(typeUint, raw0, raw1, b.EmitConstantUint(typeUint, 11), b.EmitConstantUint(typeUint, 11))
			word0Packed = b.EmitBitFieldInsert(typeUint, word0_10_11_11, raw2, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))
		}
	}

	word0_88 := b.EmitBitwiseOr(typeUint, raw0_8, b.EmitShiftLeftLogical(typeUint, raw1_8, ctx.GetConstId(ConstIdUint8)))
	word0_1616 := b.EmitBitwiseOr(typeUint, raw0_16, b.EmitShiftLeftLogical(typeUint, raw1_16, ctx.GetConstId(ConstIdUint16)))
	word0_8888_lo := b.EmitBitwiseOr(typeUint, raw0_8, b.EmitShiftLeftLogical(typeUint, raw1_8, ctx.GetConstId(ConstIdUint8)))
	word0_8888_hi := b.EmitBitwiseOr(typeUint, b.EmitShiftLeftLogical(typeUint, raw2_8, ctx.GetConstId(ConstIdUint16)), b.EmitShiftLeftLogical(typeUint, raw3_8, ctx.GetConstId(ConstIdUint24)))
	word0_8888 := b.EmitBitwiseOr(typeUint, word0_8888_lo, word0_8888_hi)

	var word0 SpirvId
	if isPacked10Format {
		word0 = word0Packed
	} else if is32BitSize {
		word0 = raw0
	} else if is16BitSize {
		if is1616 || is16161616 {
			word0 = word0_1616
		} else {
			word0 = raw0_16
		}
	} else {
		if is88 {
			word0 = word0_88
		} else if is8888 {
			word0 = word0_8888
		} else {
			word0 = raw0_8
		}
	}

	var word1 SpirvId
	if is16161616 {
		word1 = b.EmitBitwiseOr(typeUint, raw2_16, b.EmitShiftLeftLogical(typeUint, raw3_16, ctx.GetConstId(ConstIdUint16)))
	} else {
		word1 = raw1
	}
	word2 := raw2
	word3 := raw3

	// Store memory values.
	storeWord0 := func(val SpirvId) {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(ConstId64Uint0))
		if is8 {
			ptr8 := b.EmitConvertUToPtr(typePtrPsbUint8, addr)
			val8 := b.EmitUConvert(typeUint8, val)
			b.EmitStore(ptr8, val8, spec.SpvMemoryAccessAligned, 1)
		} else if is16 || is88 {
			ptr16 := b.EmitConvertUToPtr(typePtrPsbUint16, addr)
			val16 := b.EmitUConvert(typeUint16, val)
			b.EmitStore(ptr16, val16, spec.SpvMemoryAccessAligned, 2)
		} else {
			ptr32 := b.EmitConvertUToPtr(typePtrPsbUint, addr)
			b.EmitStore(ptr32, val, spec.SpvMemoryAccessAligned, 4)
		}
	}

	storeWord := func(byteOffset SpirvId, val SpirvId) {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(byteOffset))
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
