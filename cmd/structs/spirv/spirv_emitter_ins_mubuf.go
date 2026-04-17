package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMUBUF(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*MubufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)

	// Resource descriptor (SRSRC).
	res := ctx.LoadBufferResource(b, details.Srsrc)

	// sgpr_offset (from instruction SOFFSET).
	sgprOffset := ctx.GetOperandUintValue(b, details.Soffset, 0)

	// base = const_base + sgpr_offset
	base := b.EmitIAdd(typeUint64, res.BaseAddress, b.EmitUConvert(typeUint64, sgprOffset))

	var addr uint32
	if details.Addr64 {
		// ADDR64 mode: address = base(T#) + vgprAddr[63:0] + instrOffset[11:0] + sOffset
		// Base and size in resource is ignored, but base is still usually added?
		// "If set, buffer address is 64-bits (base and size in resource is ignored)."
		vgprAddrLo := ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
		vgprAddrHi := ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr+1)
		vgprAddr := ctx.Pack64(b, vgprAddrLo, vgprAddrHi)

		instrOffset := b.EmitConstantUint(typeUint, details.Offset)
		totalOffset := b.EmitIAdd(typeUint, instrOffset, sgprOffset)

		// Documentation says base and size in resource is ignored if ADDR64 is set.
		// However, many implementations still add base if it's non-zero.
		// For now let's follow the simple addition.
		addr = b.EmitIAdd(typeUint64, vgprAddr, b.EmitUConvert(typeUint64, totalOffset))
	} else {
		// Standard mode (linear or swizzled).
		vIndex := ctx.GetConstId(ConstIdxUint0)
		if details.Idxen {
			vIndex = ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
		}

		vOffset := ctx.GetConstId(ConstIdxUint0)
		if details.Offen {
			vaddrOffset := details.Vaddr
			if details.Idxen {
				vaddrOffset++
			}
			vOffset = ctx.LoadRegisterPointer(b, OpVgpr0+vaddrOffset)
		}

		// inst_offset
		instOffset := b.EmitConstantUint(typeUint, details.Offset)

		// offset = (inst_offen ? vgpr_offset : 0) + inst_offset
		offset := b.EmitIAdd(typeUint, vOffset, instOffset)

		// Calculate buffer_offset (handles linear/swizzled and TID).
		bufferOffset := ctx.CalculateBufferOffset(b,
			res.Stride, res.SwizzleEn, res.ElementSize, res.IndexStride, res.AddTidEnable,
			vIndex, offset)

		// Final address = base + bufferOffset
		addr = b.EmitIAdd(typeUint64, base, b.EmitUConvert(typeUint64, bufferOffset))
	}

	switch details.Op {
	case MubufOpLoadFormatX, MubufOpLoadDword, MubufOpLoadUbyte, MubufOpLoadSbyte, MubufOpLoadUshort, MubufOpLoadSshort:
		emitMUBUFLoad(b, instr, ctx, addr, 1)
	case MubufOpLoadFormatXy, MubufOpLoadDwordx2:
		emitMUBUFLoad(b, instr, ctx, addr, 2)
	case MubufOpLoadFormatXyz, MubufOpLoadDwordx3:
		emitMUBUFLoad(b, instr, ctx, addr, 3)
	case MubufOpLoadFormatXyzw, MubufOpLoadDwordx4:
		emitMUBUFLoad(b, instr, ctx, addr, 4)
	default:
		panic(fmt.Sprintf("unknown mubuf op %s", Mnemotics[EncMUBUF][details.Op]))
	}
}

func emitMUBUFLoad(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext, addr, count uint32) {
	details := instr.Details.(*MubufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)

	inRange := b.EmitConstantTrue(typeBool)
	ptr := b.EmitConvertUToPtr(idPtrPsbUint, addr)
	for i := range count {
		b.EmitLine(b.EmitString(fmt.Sprintf("load %d", i)), uint32(instr.DwordOffset), i)
		elementPtr := b.EmitPtrAccessChain(idPtrPsbUint, ptr, ctx.GetConstId(BlockContextId(i)))

		// Load and handle out-of-range (return 0).
		val := b.EmitLoadConditional(typeUint, elementPtr, inRange, ctx.GetConstId(ConstIdxUint0), SpvMemoryAccessAligned, 4)
		ctx.StoreRegisterPointer(b, OpVgpr0+details.Vdata+i, val)
	}
}
