package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSMRD(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*SmrdDetails)
	switch details.Op {
	case SmrdOpBufferLoadDwordx4:
		// Load constant RAM base address (pointer) from push constant.
		idPtrPsbUint := ctx.GetId(SpirvBlockContextIdPtrPsbUint)
		ptrPcPsbUint := b.EmitAccessChain(ctx.GetId(SpirvBlockContextIdPtrPcPsbUint), ctx.GetId(SpirvBlockContextIdPcVar), b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdTypeUint), 2))
		ptrBase := b.EmitLoad(idPtrPsbUint, ptrPcPsbUint)

		// Calculate offset in dwords.
		var offset uint32
		if !details.ImmOff {
			panic("s_buffer_load_dwordx4 with non-immediate offset not implemented")
		}
		if instr.HasLiteral {
			// 64-bit SMRD: offset is a 32-bit byte offset.
			offset = b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdTypeUint), instr.Literal/4)
		} else {
			// 32-bit SMRD: offset is an 8-bit dword offset.
			offset = b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdTypeUint), details.Offset)
		}

		for i := range uint32(4) {
			var idx uint32
			if i == 0 {
				idx = offset
			} else {
				idx = b.EmitIAdd(ctx.GetId(SpirvBlockContextIdTypeUint), offset, ctx.GetConstId(i))
			}
			ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, idx)
			val := b.EmitLoad(ctx.GetId(SpirvBlockContextIdTypeUint), ptr, SpvMemoryAccessAligned, 4)
			ctx.StoreRegisterPointer(b, details.Dst+i, val)
		}
	default:
		panic(fmt.Sprintf("unknown smrd op %s", Mnemotics[EncSMRD][details.Op]))
	}
}
