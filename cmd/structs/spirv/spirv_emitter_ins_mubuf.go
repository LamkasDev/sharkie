package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMUBUF(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*MubufDetails)
	idUint := ctx.GetId(SpirvBlockContextIdTypeUint)
	idUint64 := ctx.GetId(SpirvBlockContextIdTypeUint64)

	// Resource descriptor (SRSRC) is 4 SGPRs.
	sgprBase := details.Srsrc * 4
	dw0 := ctx.LoadRegisterPointer(b, OpSgpr0+sgprBase)
	dw1 := ctx.LoadRegisterPointer(b, OpSgpr0+sgprBase+1)
	dw3 := ctx.LoadRegisterPointer(b, OpSgpr0+sgprBase+3)

	// Base address and stride from resource.
	base := ctx.GetResourceBaseAddress(b, dw0, dw1)
	stride := ctx.GetResourceStride(b, dw1)
	addTidEnableBool := ctx.TestMask(b, dw3, 1<<23)

	var addr uint32
	if details.Addr64 {
		// Address = base(T#) + vgprAddr[63:0] + instrOffset[11:0] + sOffset
		vgprAddrLo := ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
		vgprAddrHi := ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr+1)
		vgprAddr := ctx.Pack64(b, vgprAddrLo, vgprAddrHi)

		instrOffset := b.EmitConstantUint(idUint, details.Offset)
		sOffset := ctx.GetOperandUintValue(b, details.Soffset, 0)

		addr = b.EmitIAdd(idUint64, base, vgprAddr)
		addr = b.EmitIAdd(idUint64, addr, b.EmitUConvert(idUint64, b.EmitIAdd(idUint, instrOffset, sOffset)))
	} else {
		// Address = base(T#) + baseOffset + iOffset + vOffset + stride * (vIndex + threadId)
		baseOffset := ctx.GetOperandUintValue(b, details.Soffset, 0)
		iOffset := b.EmitConstantUint(idUint, details.Offset)

		vIndex := ctx.GetId(SpirvBlockContextIdConstUint0)
		if details.Idxen {
			vIndex = ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
		}

		threadId := b.EmitLoad(idUint, ctx.GetId(SpirvBlockContextIdSubgroupLocalInvocationId))
		vIndexWithThreadId := b.EmitSelect(idUint, addTidEnableBool, b.EmitIAdd(idUint, vIndex, threadId), vIndex)

		vOffset := ctx.GetId(SpirvBlockContextIdConstUint0)
		if details.Offen {
			vaddrOffset := details.Vaddr
			if details.Idxen {
				vaddrOffset++
			}
			vOffset = ctx.LoadRegisterPointer(b, OpVgpr0+vaddrOffset)
		}

		addr = b.EmitIAdd(idUint64, base, b.EmitIAdd(idUint64,
			b.EmitUConvert(idUint64, b.EmitIMul(idUint, vIndexWithThreadId, stride)),
			b.EmitUConvert(idUint64, b.EmitIAdd(idUint, vOffset, b.EmitIAdd(idUint, baseOffset, iOffset)))))
	}

	switch details.Op {
	case MubufOpLoadFormatXyzw:
		idPtrPsbUint := ctx.GetId(SpirvBlockContextIdPtrPsbUint)
		ptr := b.EmitConvertUToPtr(idPtrPsbUint, addr)
		for i := range uint32(4) {
			elementPtr := b.EmitPtrAccessChain(idPtrPsbUint, ptr, b.EmitConstantUint(idUint, i))
			val := b.EmitLoad(idUint, elementPtr, SpvMemoryAccessAligned, 4)
			ctx.StoreRegisterPointer(b, OpVgpr0+details.Vdata+i, val)
		}
	default:
		panic(fmt.Sprintf("unknown mubuf op %s", Mnemotics[EncMUBUF][details.Op]))
	}
}
