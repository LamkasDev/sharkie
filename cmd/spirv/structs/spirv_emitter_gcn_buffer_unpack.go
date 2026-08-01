package structs

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitFormatUnpackHelper(b *SpvBuilder, ctx *SpirvBlockContext, absoluteAddress, dataFormat, numFormat, outOfRange SpirvId, op uint32) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeInt := ctx.GetId(BlockContextIdTypeInt)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec4 := ctx.GetId(BlockContextIdTypeV4Float)
	typeVec2 := ctx.GetId(BlockContextIdTypeV2Float)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typePtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
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

	// Load number of components based on instruction.
	inRange := b.EmitLogicalNot(typeBool, outOfRange)
	loadWord := func(offset SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(offset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		rawVal := b.EmitLoadConditional(typeUint, ptr, inRange, ctx.GetConstId(ConstIdUint0), spec.SpvMemoryAccessAligned, 4)
		return b.EmitShiftRightLogical(typeUint, rawVal, bitShift)
	}
	word0 := loadWord(ConstId64Uint0)
	f0 := b.EmitBitcast(typeFloat, word0)
	f1 := ctx.GetConstId(ConstIdFloat0)
	f2 := ctx.GetConstId(ConstIdFloat0)
	f3 := ctx.GetConstId(ConstIdFloat1)
	switch op {
	case gcnSpec.MubufOpLoadDwordx2, gcnSpec.MubufOpLoadFormatXy:
		f1 = b.EmitBitcast(typeFloat, loadWord(ConstId64Uint4))
	case gcnSpec.MubufOpLoadDwordx3, gcnSpec.MubufOpLoadFormatXyz:
		f1 = b.EmitBitcast(typeFloat, loadWord(ConstId64Uint4))
		f2 = b.EmitBitcast(typeFloat, loadWord(ConstId64Uint8))
	case gcnSpec.MubufOpLoadDwordx4, gcnSpec.MubufOpLoadFormatXyzw:
		f1 = b.EmitBitcast(typeFloat, loadWord(ConstId64Uint4))
		f2 = b.EmitBitcast(typeFloat, loadWord(ConstId64Uint8))
		f3 = b.EmitBitcast(typeFloat, loadWord(ConstId64Uint12))
	}

	// Format unpacking logic.
	switch op {
	case gcnSpec.MubufOpLoadFormatX, gcnSpec.MubufOpLoadFormatXy, gcnSpec.MubufOpLoadFormatXyz, gcnSpec.MubufOpLoadFormatXyzw:
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
		is8 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint1))
		is8Format := b.EmitLogicalAnd(typeBool, is8, isAnyIntOrNorm)
		is16 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint2))
		is16Format := b.EmitLogicalAnd(typeBool, is16, isAnyIntOrNorm)
		is88 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint3))
		is88Format := b.EmitLogicalAnd(typeBool, is88, isAnyIntOrNorm)
		is32 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint4))
		is32Format := b.EmitLogicalAnd(typeBool, is32, isAnyIntOrNorm)
		is1616 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint5))
		is1616Format := b.EmitLogicalAnd(typeBool, is1616, isAnyIntOrNorm)
		is8888 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint10))
		is8888Format := b.EmitLogicalAnd(typeBool, is8888, isAnyIntOrNorm)
		is3232 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint11))
		is3232Format := b.EmitLogicalAnd(typeBool, is3232, isAnyIntOrNorm)
		is16161616 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint12))
		is16161616Format := b.EmitLogicalAnd(typeBool, is16161616, isAnyIntOrNorm)
		is323232 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint13))
		is323232Format := b.EmitLogicalAnd(typeBool, is323232, isAnyIntOrNorm)
		is32323232 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint14))
		is32323232Format := b.EmitLogicalAnd(typeBool, is32323232, isAnyIntOrNorm)

		// Select raw values based on format.
		is16BitSize := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is16Format, is1616Format), is16161616Format)
		is32BitSize := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is32Format, is3232Format), b.EmitLogicalOr(typeBool, is323232Format, is32323232Format))

		// Extract bytes for 8-bit formats.
		comp0Byte := b.EmitBitwiseAnd(typeUint, word0, cFF)
		comp1Byte := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint8)), cFF)
		comp2Byte := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint16)), cFF)
		comp3Byte := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint24)), cFF)

		// Extract shorts for 16-bit formats.
		comp0Short := b.EmitBitwiseAnd(typeUint, word0, cFFFF)
		comp1Short := b.EmitShiftRightLogical(typeUint, word0, ctx.GetConstId(ConstIdUint16))

		word1 := b.EmitBitcast(typeUint, f1)
		comp2Short := b.EmitBitwiseAnd(typeUint, word1, cFFFF)
		comp3Short := b.EmitShiftRightLogical(typeUint, word1, ctx.GetConstId(ConstIdUint16))

		word2 := b.EmitBitcast(typeUint, f2)
		word3 := b.EmitBitcast(typeUint, f3)
		raw0 := b.EmitSelect(typeUint, is32BitSize, word0, b.EmitSelect(typeUint, is16BitSize, comp0Short, comp0Byte))
		raw1 := b.EmitSelect(typeUint, is32BitSize, word1, b.EmitSelect(typeUint, is16BitSize, comp1Short, comp1Byte))
		raw2 := b.EmitSelect(typeUint, is32BitSize, word2, b.EmitSelect(typeUint, is16BitSize, comp2Short, comp2Byte))
		raw3 := b.EmitSelect(typeUint, is32BitSize, word3, b.EmitSelect(typeUint, is16BitSize, comp3Short, comp3Byte))

		// Divisor for UNORM.
		divisor := b.EmitSelect(typeFloat, is32BitSize, ctx.GetConstId(ConstIdFloat1), b.EmitSelect(typeFloat, is16BitSize, ctx.GetConstId(ConstIdFloat65535), ctx.GetConstId(ConstIdFloat255)))

		// Unorm values.
		comp0Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw0), divisor)
		comp1Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw1), divisor)
		comp2Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw2), divisor)
		comp3Unorm := b.EmitFDiv(typeFloat, b.EmitConvertUToF(typeFloat, raw3), divisor)

		// Uint values.
		comp0UintF := b.EmitBitcast(typeFloat, raw0)
		comp1UintF := b.EmitBitcast(typeFloat, raw1)
		comp2UintF := b.EmitBitcast(typeFloat, raw2)
		comp3UintF := b.EmitBitcast(typeFloat, raw3)

		// For 16-bit, shift left 16 then arithmetic shift right 16. For 8-bit, use 24.
		extendShift := b.EmitSelect(typeUint, is16BitSize, ctx.GetConstId(ConstIdUint16), ctx.GetConstId(ConstIdUint24))
		makeSigned := func(raw SpirvId) SpirvId {
			shiftedLeft := b.EmitShiftLeftLogical(typeUint, raw, extendShift)
			intShifted := b.EmitBitcast(typeInt, shiftedLeft)
			return b.EmitShiftRightArithmetic(typeInt, intShifted, extendShift)
		}

		raw0Signed := makeSigned(raw0)
		raw1Signed := makeSigned(raw1)
		raw2Signed := makeSigned(raw2)
		raw3Signed := makeSigned(raw3)

		// Sint values.
		comp0SintF := b.EmitBitcast(typeFloat, raw0Signed)
		comp1SintF := b.EmitBitcast(typeFloat, raw1Signed)
		comp2SintF := b.EmitBitcast(typeFloat, raw2Signed)
		comp3SintF := b.EmitBitcast(typeFloat, raw3Signed)

		// Uscaled values.
		comp0Uscaled := b.EmitConvertUToF(typeFloat, raw0)
		comp1Uscaled := b.EmitConvertUToF(typeFloat, raw1)
		comp2Uscaled := b.EmitConvertUToF(typeFloat, raw2)
		comp3Uscaled := b.EmitConvertUToF(typeFloat, raw3)

		// Sscaled values.
		comp0Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(ctx.GetId(BlockContextIdTypeInt), raw0Signed))
		comp1Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(ctx.GetId(BlockContextIdTypeInt), raw1Signed))
		comp2Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(ctx.GetId(BlockContextIdTypeInt), raw2Signed))
		comp3Sscaled := b.EmitConvertSToF(typeFloat, b.EmitBitcast(ctx.GetId(BlockContextIdTypeInt), raw3Signed))

		// If it's a 32-bit float, raw0 is already a 32-bit float. We only unpack 16-bit.
		unpackHalf := func(raw SpirvId) SpirvId {
			vec2 := b.EmitExtInst(typeVec2, ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpUnpackHalf2x16, raw)
			return b.EmitCompositeExtract(typeFloat, vec2, 0)
		}
		comp0Float16 := unpackHalf(raw0)
		comp1Float16 := unpackHalf(raw1)
		comp2Float16 := unpackHalf(raw2)
		comp3Float16 := unpackHalf(raw3)

		// Float values.
		comp0Float := b.EmitSelect(typeFloat, is16BitSize, comp0Float16, comp0UintF)
		comp1Float := b.EmitSelect(typeFloat, is16BitSize, comp1Float16, comp1UintF)
		comp2Float := b.EmitSelect(typeFloat, is16BitSize, comp2Float16, comp2UintF)
		comp3Float := b.EmitSelect(typeFloat, is16BitSize, comp3Float16, comp3UintF)

		// Selection chain.
		selectComp := func(compUnorm, compSint, compFloat, compUscaled, compSscaled, compUint SpirvId) SpirvId {
			comp := b.EmitSelect(typeFloat, isUnorm, compUnorm, compUint)
			comp = b.EmitSelect(typeFloat, isSint, compSint, comp)
			comp = b.EmitSelect(typeFloat, isFloat, compFloat, comp)
			comp = b.EmitSelect(typeFloat, isUscaled, compUscaled, comp)
			comp = b.EmitSelect(typeFloat, isSscaled, compSscaled, comp)
			return comp
		}
		comp0 := selectComp(comp0Unorm, comp0SintF, comp0Float, comp0Uscaled, comp0Sscaled, comp0UintF)
		comp1 := selectComp(comp1Unorm, comp1SintF, comp1Float, comp1Uscaled, comp1Sscaled, comp1UintF)
		comp2 := selectComp(comp2Unorm, comp2SintF, comp2Float, comp2Uscaled, comp2Sscaled, comp2UintF)
		comp3 := selectComp(comp3Unorm, comp3SintF, comp3Float, comp3Uscaled, comp3Sscaled, comp3UintF)

		// Apply based on format match.
		isAny1Comp := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is8Format, is16Format), is32Format)
		isAny2Comp := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is88Format, is1616Format), is3232Format)
		isAny3Comp := is323232Format
		isAny4Comp := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is8888Format, is16161616Format), is32323232Format)
		isAny := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), isAny3Comp), isAny4Comp)

		f0 = b.EmitSelect(typeFloat, isAny, comp0, f0)
		f1 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAny2Comp, isAny3Comp), isAny4Comp), comp1, f1)
		f2 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, isAny3Comp, isAny4Comp), comp2, f2)
		f3 = b.EmitSelect(typeFloat, isAny4Comp, comp3, f3)

		f1 = b.EmitSelect(typeFloat, isAny1Comp, ctx.GetConstId(ConstIdFloat0), f1)
		f2 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), ctx.GetConstId(ConstIdFloat0), f2)
		f3 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), isAny3Comp), ctx.GetConstId(ConstIdFloat1), f3)

		// Debug unpack behavior for lane 0.
		/* ctx.EmitDebugPrintfLane(b, 0,
			"mubuf_unpack op=%d dfmt=%d nfmt=%d raw=%08x,%08x out=%f,%f,%f,%f\n",
			b.EmitConstantUint(typeUint, op), dataFormat, numFormat, word0, word1, f0, f1, f2, f3,
		) */
	}

	return b.EmitCompositeConstruct(typeVec4, f0, f1, f2, f3)
}
