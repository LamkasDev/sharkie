package structs

import (
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func EmitFormatUnpackHelper(b *SpvBuilder, ctx *SpirvBlockContext, baseAddress, byteOffset, dw3 SpirvId, op uint32) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec4 := ctx.GetId(BlockContextIdTypeV4Float)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	typePtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	cFF := ctx.GetConstId(ConstIdUint255)
	cFFFF := ctx.GetConstId(ConstIdUintFFFF)

	// Calculate final address.
	byteOffsetAligned := b.EmitBitwiseAnd(typeUint, byteOffset, ctx.GetConstId(ConstIdUintFFFFFFFFC))
	byteOffset64 := b.EmitUConvert(typeUint64, byteOffsetAligned)
	totalAddress := b.EmitIAdd(typeUint64, baseAddress, byteOffset64)
	translatedAddress := ctx.TranslateAddress(b, totalAddress)

	// Load number of components based on instruction.
	loadWord := func(offset SpirvId) SpirvId {
		addr := b.EmitIAdd(typeUint64, translatedAddress, ctx.GetConstId(offset))
		ptr := b.EmitConvertUToPtr(typePtrPsbUint, addr)
		return b.EmitLoad(typeUint, ptr, spec.SpvMemoryAccessAligned, 4)
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
		dataFormat := b.EmitBitFieldUExtract(typeUint, dw3, ctx.GetConstId(ConstIdUint15), ctx.GetConstId(ConstIdUint4))
		numFormat := b.EmitBitFieldUExtract(typeUint, dw3, ctx.GetConstId(ConstIdUint12), ctx.GetConstId(ConstIdUint3))

		isUnorm := b.EmitLogicalOr(typeBool,
			b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint0)),
			b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint1))) // Treat snorm as unorm
		isUint := b.EmitLogicalOr(typeBool,
			b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint4)),
			b.EmitIEqual(typeBool, numFormat, ctx.GetConstId(ConstIdUint5))) // Treat sint as uint
		isAnyIntOrNorm := b.EmitLogicalOr(typeBool, isUnorm, isUint)

		// 8 (DataFormat 1)
		is8 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint1))
		is8Format := b.EmitLogicalAnd(typeBool, is8, isAnyIntOrNorm)

		// 16 (DataFormat 2)
		is16 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint2))
		is16Format := b.EmitLogicalAnd(typeBool, is16, isAnyIntOrNorm)

		// 8_8 (DataFormat 3)
		is88 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint3))
		is88Format := b.EmitLogicalAnd(typeBool, is88, isAnyIntOrNorm)

		// 16_16 (DataFormat 5)
		is1616 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint5))
		is1616Format := b.EmitLogicalAnd(typeBool, is1616, isAnyIntOrNorm)

		// 8_8_8_8 (DataFormat 10)
		is8888 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint10))
		is8888Format := b.EmitLogicalAnd(typeBool, is8888, isAnyIntOrNorm)

		// 16_16_16_16 (DataFormat 12)
		is16161616 := b.EmitIEqual(typeBool, dataFormat, ctx.GetConstId(ConstIdUint12))
		is16161616Format := b.EmitLogicalAnd(typeBool, is16161616, isAnyIntOrNorm)

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

		// Select raw values based on format.
		is16BitSize := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, is16Format, is1616Format), is16161616Format)
		raw0 := b.EmitSelect(typeUint, is16BitSize, comp0Short, comp0Byte)
		raw1 := b.EmitSelect(typeUint, is16BitSize, comp1Short, comp1Byte)
		raw2 := b.EmitSelect(typeUint, is16BitSize, comp2Short, comp2Byte)
		raw3 := b.EmitSelect(typeUint, is16BitSize, comp3Short, comp3Byte)

		// Divisor for UNORM.
		divisor := b.EmitSelect(typeFloat, is16BitSize, ctx.GetConstId(ConstIdFloat65535), ctx.GetConstId(ConstIdFloat255))

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

		// Final component values.
		comp0 := b.EmitSelect(typeFloat, isUnorm, comp0Unorm, comp0UintF)
		comp1 := b.EmitSelect(typeFloat, isUnorm, comp1Unorm, comp1UintF)
		comp2 := b.EmitSelect(typeFloat, isUnorm, comp2Unorm, comp2UintF)
		comp3 := b.EmitSelect(typeFloat, isUnorm, comp3Unorm, comp3UintF)

		// Apply based on format match.
		isAny1Comp := b.EmitLogicalOr(typeBool, is8Format, is16Format)
		isAny2Comp := b.EmitLogicalOr(typeBool, is88Format, is1616Format)
		isAny4Comp := b.EmitLogicalOr(typeBool, is8888Format, is16161616Format)
		isAny := b.EmitLogicalOr(typeBool, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), isAny4Comp)

		f0 = b.EmitSelect(typeFloat, isAny, comp0, f0)
		f1 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, isAny2Comp, isAny4Comp), comp1, f1)
		f2 = b.EmitSelect(typeFloat, isAny4Comp, comp2, f2)
		f3 = b.EmitSelect(typeFloat, isAny4Comp, comp3, f3)

		f1 = b.EmitSelect(typeFloat, isAny1Comp, ctx.GetConstId(ConstIdFloat0), f1)
		f2 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), ctx.GetConstId(ConstIdFloat0), f2)
		f3 = b.EmitSelect(typeFloat, b.EmitLogicalOr(typeBool, isAny1Comp, isAny2Comp), ctx.GetConstId(ConstIdFloat1), f3)
	}

	return b.EmitCompositeConstruct(typeVec4, f0, f1, f2, f3)
}
