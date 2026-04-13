package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSMRD(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*SmrdDetails)
	switch details.Op {
	case SmrdOpLoadDword:
		emitSMRDLoad(b, instr, ctx, 1)
	case SmrdOpLoadDwordx2:
		emitSMRDLoad(b, instr, ctx, 2)
	case SmrdOpLoadDwordx4:
		emitSMRDLoad(b, instr, ctx, 4)
	case SmrdOpLoadDwordx8:
		emitSMRDLoad(b, instr, ctx, 8)
	case SmrdOpLoadDwordx16:
		emitSMRDLoad(b, instr, ctx, 16)
	case SmrdOpBufferLoadDword:
		emitSMRDLoad(b, instr, ctx, 1)
	case SmrdOpBufferLoadDwordx2:
		emitSMRDLoad(b, instr, ctx, 2)
	case SmrdOpBufferLoadDwordx4:
		emitSMRDLoad(b, instr, ctx, 4)
	case SmrdOpBufferLoadDwordx8:
		emitSMRDLoad(b, instr, ctx, 8)
	case SmrdOpBufferLoadDwordx16:
		emitSMRDLoad(b, instr, ctx, 16)
	default:
		panic(fmt.Sprintf("unknown smrd op %s", Mnemotics[EncSMRD][details.Op]))
	}
}

func emitSMRDLoad(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext, count uint32) {
	details := instr.Details.(*SmrdDetails)

	// Load constant RAM base address from push constant.
	idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	ptrBase := ctx.LoadPushConstantValue(b, PushConstantConstRamAddress)

	// Calculate offset in dwords.
	var offset uint32
	if details.ImmOff {
		if instr.HasLiteral {
			// 64-bit SMRD: offset is a 32-bit byte offset.
			offset = b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), instr.Literal/4)
		} else {
			// 32-bit SMRD: offset is an 8-bit dword offset.
			offset = b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), details.Offset)
		}
	} else {
		// Offset is an SGPR index containing a Dword offset.
		offset = ctx.LoadRegisterPointer(b, OpSgpr0+details.Offset)
	}

	for i := range count {
		var idx uint32
		if i == 0 {
			idx = offset
		} else {
			idx = b.EmitIAdd(ctx.GetId(BlockContextIdTypeUint), offset, ctx.GetGcnConstId(i))
		}
		ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, idx)
		val := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ptr, SpvMemoryAccessAligned, 4)
		ctx.StoreRegisterPointer(b, OpSgpr0+details.Dst+i, val)
	}
}
