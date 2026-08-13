package structs

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitFormatUnpackHelper(b *SpvBuilder, ctx *SpirvBlockContext, bindingIndex uint32, offset SpirvId, dataFormat, numFormat uint32, outOfRange SpirvId, op uint32) SpirvId {
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

	translatedAddress := ctx.TranslateAddress(b, bindingIndex, offset)
	inRange := b.EmitLogicalNot(typeBool, outOfRange)

	isTypedLoad := false
	switch op {
	case gcnSpec.MubufOpLoadFormatX, gcnSpec.MubufOpLoadFormatXy, gcnSpec.MubufOpLoadFormatXyz, gcnSpec.MubufOpLoadFormatXyzw:
		isTypedLoad = true
	}

	needsWord1 := false
	needsWord2 := false
	needsWord3 := false

	is8BitMem := false
	is16BitMem := false
	is32BitMem := true

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

	if isTypedLoad {
		is8 = dataFormat == 1
		is16 = dataFormat == 2
		is88 = dataFormat == 3
		is32 = dataFormat == 4
		is1616 = dataFormat == 5
		is101111 = dataFormat == 6
		is1010102 = dataFormat == 8
		is2101010 = dataFormat == 9
		is8888 = dataFormat == 10
		is3232 = dataFormat == 11
		is16161616 = dataFormat == 12
		is323232 = dataFormat == 13
		is32323232 = dataFormat == 14

		is8BitMem = is8
		is16BitMem = is16 || is88
		is32BitMem = !(is8BitMem || is16BitMem)

		needsWord1 = is3232 || is16161616 || is323232 || is32323232
		needsWord2 = is323232 || is32323232
		needsWord3 = is32323232
	} else {
		switch op {
		case gcnSpec.MubufOpLoadDwordx2:
			needsWord1 = true
		case gcnSpec.MubufOpLoadDwordx3:
			needsWord1 = true
			needsWord2 = true
		case gcnSpec.MubufOpLoadDwordx4:
			needsWord1 = true
			needsWord2 = true
			needsWord3 = true
		}
	}

	// Load memory values.
	const0_8 := b.EmitConstantUint(typeUint8, 0)
	const0_16 := b.EmitConstantUint(typeUint16, 0)

	load8 := func(byteOffset SpirvId, cond SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(byteOffset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint8, addr)
		val8 := b.EmitLoadConditional(typeUint8, ptr, cond, const0_8, spec.SpvMemoryAccessAligned, 1)
		return b.EmitUConvert(typeUint, val8)
	}
	load16 := func(byteOffset SpirvId, cond SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(byteOffset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint16, addr)
		val16 := b.EmitLoadConditional(typeUint16, ptr, cond, const0_16, spec.SpvMemoryAccessAligned, 2)
		return b.EmitUConvert(typeUint, val16)
	}
	load32 := func(byteOffset SpirvId, cond SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(byteOffset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		return b.EmitLoadConditional(typeUint, ptr, cond, ctx.GetConstId(ConstIdUint0), spec.SpvMemoryAccessAligned, 4)
	}

	var word0, word1, word2, word3 SpirvId

	if is32BitMem {
		word0 = load32(ConstId64Uint0, inRange)
	} else if is16BitMem {
		word0 = load16(ConstId64Uint0, inRange)
	} else {
		word0 = load8(ConstId64Uint0, inRange)
	}

	if needsWord1 {
		word1 = load32(ConstId64Uint4, inRange)
	} else {
		word1 = ctx.GetConstId(ConstIdUint0)
	}
	if needsWord2 {
		word2 = load32(ConstId64Uint8, inRange)
	} else {
		word2 = ctx.GetConstId(ConstIdUint0)
	}
	if needsWord3 {
		word3 = load32(ConstId64Uint12, inRange)
	} else {
		word3 = ctx.GetConstId(ConstIdUint0)
	}

	f0 := b.EmitBitcast(typeFloat, word0)
	f1 := b.EmitBitcast(typeFloat, word1)
	f2 := b.EmitBitcast(typeFloat, word2)
	f3 := b.EmitBitcast(typeFloat, word3)

	if isTypedLoad {
		// Format conversions & unpacking.
		isUnorm := numFormat == 0
		isSnorm := numFormat == 1
		isUscaled := numFormat == 2
		isSscaled := numFormat == 3
		isUintF := numFormat == 4
		isSintF := numFormat == 5
		isSnormOgl := numFormat == 6
		isFloatF := numFormat == 7
		isAnyNorm := isUnorm || isSnorm || isSnormOgl
		isAnyIntOrNorm := isAnyNorm || isUintF || isUscaled || isSscaled || isSintF || isFloatF

		is8Format := is8 && isAnyIntOrNorm
		is16Format := is16 && isAnyIntOrNorm
		is88Format := is88 && isAnyIntOrNorm
		is32Format := is32 && isAnyIntOrNorm
		is1616Format := is1616 && isAnyIntOrNorm
		is8888Format := is8888 && isAnyIntOrNorm
		is3232Format := is3232 && isAnyIntOrNorm
		is16161616Format := is16161616 && isAnyIntOrNorm
		is323232Format := is323232 && isAnyIntOrNorm
		is32323232Format := is32323232 && isAnyIntOrNorm
		isPacked10Format := (is101111 || is1010102 || is2101010) && isAnyIntOrNorm

		is16BitSize := is16Format || is1616Format || is16161616Format
		is32BitSize := is32Format || is3232Format || is323232Format || is32323232Format

		// Standard byte and short extraction.
		var raw0, raw1, raw2, raw3 SpirvId

		if isPacked10Format {
			if is2101010 {
				raw0 = b.EmitBitFieldUExtract(typeUint, word0, ctx.GetConstId(ConstIdUint0), b.EmitConstantUint(typeUint, 10))
				raw1 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 10), b.EmitConstantUint(typeUint, 10))
				raw2 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 20), b.EmitConstantUint(typeUint, 10))
				raw3 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 30), b.EmitConstantUint(typeUint, 2))
			} else if is1010102 {
				raw0 = b.EmitBitFieldUExtract(typeUint, word0, ctx.GetConstId(ConstIdUint0), b.EmitConstantUint(typeUint, 2))
				raw1 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 2), b.EmitConstantUint(typeUint, 10))
				raw2 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 12), b.EmitConstantUint(typeUint, 10))
				raw3 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))
			} else if is101111 {
				raw0 = b.EmitBitFieldUExtract(typeUint, word0, ctx.GetConstId(ConstIdUint0), b.EmitConstantUint(typeUint, 11))
				raw1 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 11), b.EmitConstantUint(typeUint, 11))
				raw2 = b.EmitBitFieldUExtract(typeUint, word0, b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 10))
				raw3 = ctx.GetConstId(ConstIdUint0)
			}
		} else if is32BitSize {
			raw0 = word0
			raw1 = word1
			raw2 = word2
			raw3 = word3
		} else if is16BitSize {
			raw0 = b.EmitBitwiseAnd(typeUint, word0, cFFFF)
			raw1 = b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint16))
			raw2 = b.EmitBitwiseAnd(typeUint, word1, cFFFF)
			raw3 = b.EmitShiftRightLogical(typeUint, word1, ctx.GetConstId(ConstIdUint16))
		} else { // 8 bit size
			raw0 = b.EmitBitwiseAnd(typeUint, word0, cFF)
			raw1 = b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint8)), cFF)
			raw2 = b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint16)), cFF)
			raw3 = b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint24)), cFF)
		}

		// Divisor for UNORM.
		var div0, div1, div2, div3 SpirvId
		if isPacked10Format {
			div_2_10_10_10_x := b.EmitConstantFloat(typeFloat, 1023.0)
			div_2_10_10_10_w := b.EmitConstantFloat(typeFloat, 3.0)
			if is2101010 {
				div0 = div_2_10_10_10_x
				div1 = div_2_10_10_10_x
				div2 = div_2_10_10_10_x
				div3 = div_2_10_10_10_w
			} else if is1010102 {
				div0 = div_2_10_10_10_w
				div1 = div_2_10_10_10_x
				div2 = div_2_10_10_10_x
				div3 = div_2_10_10_10_x
			} else {
				div0 = div_2_10_10_10_x
				div1 = div_2_10_10_10_x
				div2 = div_2_10_10_10_x
				div3 = ctx.GetConstId(ConstIdFloat1)
			}
		} else {
			var divisor SpirvId
			if is32BitSize {
				divisor = ctx.GetConstId(ConstIdFloat1)
			} else if is16BitSize {
				divisor = ctx.GetConstId(ConstIdFloat65535)
			} else {
				divisor = ctx.GetConstId(ConstIdFloat255)
			}
			div0, div1, div2, div3 = divisor, divisor, divisor, divisor
		}

		comp0Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw0), div0)
		comp1Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw1), div1)
		comp2Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw2), div2)
		comp3Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw3), div3)

		comp0UintF := b.EmitBitcast(typeFloat, raw0)
		comp1UintF := b.EmitBitcast(typeFloat, raw1)
		comp2UintF := b.EmitBitcast(typeFloat, raw2)
		comp3UintF := b.EmitBitcast(typeFloat, raw3)

		var extendShift0, extendShift1, extendShift2, extendShift3 SpirvId
		if isPacked10Format {
			if is2101010 {
				extendShift0, extendShift1, extendShift2, extendShift3 = b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 30)
			} else if is1010102 {
				extendShift0, extendShift1, extendShift2, extendShift3 = b.EmitConstantUint(typeUint, 30), b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 22), b.EmitConstantUint(typeUint, 22)
			} else {
				extendShift0, extendShift1, extendShift2, extendShift3 = b.EmitConstantUint(typeUint, 21), b.EmitConstantUint(typeUint, 21), b.EmitConstantUint(typeUint, 22), ctx.GetConstId(ConstIdUint0)
			}
		} else if is16BitSize {
			extendShift0, extendShift1, extendShift2, extendShift3 = ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint16)
		} else {
			extendShift0, extendShift1, extendShift2, extendShift3 = ctx.GetConstId(ConstIdUint24), ctx.GetConstId(ConstIdUint24), ctx.GetConstId(ConstIdUint24), ctx.GetConstId(ConstIdUint24)
		}

		makeSigned := func(raw, shift SpirvId) SpirvId {
			if shift == ctx.GetConstId(ConstIdUint0) {
				return raw
			}
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
			shifted := b.EmitShiftLeftLogical(typeUint, raw, b.EmitConstantUint(typeUint, 17))
			expExt := b.EmitBitFieldUExtract(typeUint, raw, b.EmitConstantUint(typeUint, 6), b.EmitConstantUint(typeUint, 5))
			isZeroExp := b.EmitIEqual(typeBool, expExt, ctx.GetConstId(ConstIdUint0))
			magic := b.EmitBitwiseOr(typeUint, shifted, b.EmitConstantUint(typeUint, 0x38000000))
			fMagic := b.EmitFSub(typeFloat, b.EmitBitcast(typeFloat, magic), b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0x38000000)))
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

		var comp0Float101111, comp1Float101111, comp2Float101111 SpirvId
		var comp0Float16, comp1Float16, comp2Float16, comp3Float16 SpirvId
		if is101111 {
			comp0Float101111 = ufloat11ToF32(raw0)
			comp1Float101111 = ufloat11ToF32(raw1)
			comp2Float101111 = ufloat10ToF32(raw2)
		}
		if is16BitSize {
			comp0Float16 = unpackHalf(raw0)
			comp1Float16 = unpackHalf(raw1)
			comp2Float16 = unpackHalf(raw2)
			comp3Float16 = unpackHalf(raw3)
		}

		var comp0Float, comp1Float, comp2Float, comp3Float SpirvId
		if is101111 {
			comp0Float = comp0Float101111
			comp1Float = comp1Float101111
			comp2Float = comp2Float101111
			comp3Float = comp3UintF
		} else if is16BitSize {
			comp0Float = comp0Float16
			comp1Float = comp1Float16
			comp2Float = comp2Float16
			comp3Float = comp3Float16
		} else {
			comp0Float = comp0UintF
			comp1Float = comp1UintF
			comp2Float = comp2UintF
			comp3Float = comp3UintF
		}

		var comp0, comp1, comp2, comp3 SpirvId
		if isUnorm {
			comp0, comp1, comp2, comp3 = comp0Unorm, comp1Unorm, comp2Unorm, comp3Unorm
		} else if isSnorm {
			comp0, comp1, comp2, comp3 = comp0Snorm, comp1Snorm, comp2Snorm, comp3Snorm
		} else if isSnormOgl {
			comp0, comp1, comp2, comp3 = comp0SnormOgl, comp1SnormOgl, comp2SnormOgl, comp3SnormOgl
		} else if isSintF {
			comp0, comp1, comp2, comp3 = comp0SintF, comp1SintF, comp2SintF, comp3SintF
		} else if isFloatF {
			comp0, comp1, comp2, comp3 = comp0Float, comp1Float, comp2Float, comp3Float
		} else if isUscaled {
			comp0, comp1, comp2, comp3 = comp0Uscaled, comp1Uscaled, comp2Uscaled, comp3Uscaled
		} else if isSscaled {
			comp0, comp1, comp2, comp3 = comp0Sscaled, comp1Sscaled, comp2Sscaled, comp3Sscaled
		} else { // UintF
			comp0, comp1, comp2, comp3 = comp0UintF, comp1UintF, comp2UintF, comp3UintF
		}

		isAny1Comp := is8Format || is16Format || is32Format
		isAny2Comp := is88Format || is1616Format || is3232Format
		isAny3Comp := is323232Format || is101111
		isAny4Comp := is8888Format || is16161616Format || is32323232Format || is1010102 || is2101010
		isAny := isAny1Comp || isAny2Comp || isAny3Comp || isAny4Comp

		if isAny {
			f0 = comp0
		}
		if isAny2Comp || isAny3Comp || isAny4Comp {
			f1 = comp1
		} else if isAny1Comp {
			f1 = ctx.GetConstId(ConstIdFloat0)
		}
		if isAny3Comp || isAny4Comp {
			f2 = comp2
		} else if isAny1Comp || isAny2Comp {
			f2 = ctx.GetConstId(ConstIdFloat0)
		}
		if isAny4Comp {
			f3 = comp3
		} else if isAny1Comp || isAny2Comp || isAny3Comp {
			f3 = ctx.GetConstId(ConstIdFloat1)
		}
	}

	return b.EmitCompositeConstruct(typeVec4, f0, f1, f2, f3)
}
