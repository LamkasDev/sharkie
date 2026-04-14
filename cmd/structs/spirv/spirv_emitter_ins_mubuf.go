package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMUBUF(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*MubufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)

	// Resource descriptor (SRSRC) is 4 SGPRs.
	sgprBase := details.Srsrc * 4
	dw0 := ctx.LoadRegisterPointer(b, OpSgpr0+sgprBase)
	dw1 := ctx.LoadRegisterPointer(b, OpSgpr0+sgprBase+1)
	dw3 := ctx.LoadRegisterPointer(b, OpSgpr0+sgprBase+3)

	// Base address and stride from resource.
	base := ctx.GetResourceBaseAddress(b, dw0, dw1)
	stride := ctx.GetResourceStride(b, dw1)
	addTidEnableBool := ctx.TestMask(b, dw3, 1<<17)

	var addr uint32
	if details.Addr64 {
		// Address = base(T#) + vgprAddr[63:0] + instrOffset[11:0] + sOffset
		vgprAddrLo := ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
		vgprAddrHi := ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr+1)
		vgprAddr := ctx.Pack64(b, vgprAddrLo, vgprAddrHi)

		instrOffset := b.EmitConstantUint(typeUint, details.Offset)
		sOffset := ctx.GetOperandUintValue(b, details.Soffset, 0)

		addr = b.EmitIAdd(typeUint64, base, vgprAddr)
		addr = b.EmitIAdd(typeUint64, addr, b.EmitUConvert(typeUint64, b.EmitIAdd(typeUint, instrOffset, sOffset)))
	} else {
		// Address = base(T#) + baseOffset + iOffset + vOffset + stride * (vIndex + threadId)
		baseOffset := ctx.GetOperandUintValue(b, details.Soffset, 0)
		iOffset := b.EmitConstantUint(typeUint, details.Offset)

		vIndex := ctx.GetConstId(ConstIdxUint0)
		if details.Idxen {
			vIndex = ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
		}

		threadId := b.EmitLoad(typeUint, ctx.GetId(BlockContextIdSubgroupLocalInvocationId))
		vIndexWithThreadId := b.EmitSelect(typeUint, addTidEnableBool, b.EmitIAdd(typeUint, vIndex, threadId), vIndex)

		vOffset := ctx.GetConstId(ConstIdxUint0)
		if details.Offen {
			vaddrOffset := details.Vaddr
			if details.Idxen {
				vaddrOffset++
			}
			vOffset = ctx.LoadRegisterPointer(b, OpVgpr0+vaddrOffset)
		}

		addr = b.EmitIAdd(typeUint64, base, b.EmitIAdd(typeUint64,
			b.EmitUConvert(typeUint64, b.EmitIMul(typeUint, vIndexWithThreadId, stride)),
			b.EmitUConvert(typeUint64, b.EmitIAdd(typeUint, vOffset, b.EmitIAdd(typeUint, baseOffset, iOffset)))))
	}

	switch details.Op {
	case MubufOpLoadFormatX, MubufOpLoadDword:
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

func emitMUBUFLoad(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext, addr uint32, count uint32) {
	details := instr.Details.(*MubufDetails)
	idUint := ctx.GetId(BlockContextIdTypeUint)
	idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)

	ptr := b.EmitConvertUToPtr(idPtrPsbUint, addr)
	for i := range count {
		b.EmitLine(b.EmitString(fmt.Sprintf("load %d", i)), uint32(instr.DwordOffset), i)
		elementPtr := b.EmitPtrAccessChain(idPtrPsbUint, ptr, ctx.GetConstId(BlockContextId(i)))
		val := b.EmitLoad(idUint, elementPtr, SpvMemoryAccessAligned, 4)
		ctx.StoreRegisterPointer(b, OpVgpr0+details.Vdata+i, val)
	}
}
