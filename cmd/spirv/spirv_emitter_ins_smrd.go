package spirv

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func emitSMRD(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	switch details.Op {
	case gcnSpec.SmrdOpLoadDword:
		emitSMRDLoad(b, instr, ctx, 1)
	case gcnSpec.SmrdOpLoadDwordx2:
		emitSMRDLoad(b, instr, ctx, 2)
	case gcnSpec.SmrdOpLoadDwordx4:
		emitSMRDLoad(b, instr, ctx, 4)
	case gcnSpec.SmrdOpLoadDwordx8:
		emitSMRDLoad(b, instr, ctx, 8)
	case gcnSpec.SmrdOpLoadDwordx16:
		emitSMRDLoad(b, instr, ctx, 16)
	case gcnSpec.SmrdOpBufferLoadDword:
		emitSMRDLoad(b, instr, ctx, 1)
	case gcnSpec.SmrdOpBufferLoadDwordx2:
		emitSMRDLoad(b, instr, ctx, 2)
	case gcnSpec.SmrdOpBufferLoadDwordx4:
		emitSMRDLoad(b, instr, ctx, 4)
	case gcnSpec.SmrdOpBufferLoadDwordx8:
		emitSMRDLoad(b, instr, ctx, 8)
	case gcnSpec.SmrdOpBufferLoadDwordx16:
		emitSMRDLoad(b, instr, ctx, 16)
	default:
		panic(fmt.Sprintf("unknown smrd op %s", gcnSpec.Mnemotics[gcnSpec.EncSMRD][details.Op]))
	}
}

func emitSMRDLoad(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext, count uint32) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)

	// Load 64-bit base address from SGPRs.
	// Base index is dword-based, so SBASE * 2.
	lo, hi := ctx.GetOperand64Value(b, gcnSpec.OpSgpr0+details.Base*2, 0)

	// Base address is always 48 bits in GCN3 for SMRD.
	hi = b.EmitBitwiseAnd(typeUint, hi, b.EmitConstantUint(typeUint, 0xFFFF))
	base64 := ctx.Pack64(b, lo, hi)

	// Calculate offset in bytes.
	var byteOffset SpirvId
	if details.ImmOff {
		if instr.HasLiteral {
			// TODO: use built-ins.
			// 64-bit SMRD: offset is a 32-bit byte offset.
			byteOffset = b.EmitConstantUint(typeUint, instr.Literal)
		} else {
			// TODO: use built-ins.
			// 32-bit SMRD: offset is an 8-bit unsigned dword offset.
			byteOffset = b.EmitConstantUint(typeUint, details.Offset*4)
		}
	} else {
		// Offset is an SGPR index containing a dword offset.
		offsetVal := ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Offset, 0)
		byteOffset = b.EmitIMul(typeUint, offsetVal, ctx.GetConstId(ConstIdUint4))
	}

	// m_addr = (base + m_offset) & ~0x3
	byteOffset64 := b.EmitUConvert(typeUint64, byteOffset)
	addr64 := b.EmitIAdd(typeUint64, base64, byteOffset64)
	mask64 := b.EmitConstantUint64(typeUint64, ^uint64(0x3))
	addr64Aligned := b.EmitBitwiseAnd(typeUint64, addr64, mask64)

	// Translate and cast to pointer.
	translatedAddr64 := ctx.TranslateAddress(b, addr64Aligned)
	ptrBase := b.EmitBitcast(idPtrPsbUint, translatedAddr64)

	for i := range count {
		// Load each dword.
		ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, ctx.GetConstId(ConstIdUint0+SpirvId(i)))
		val := b.EmitLoad(typeUint, ptr, spec.SpvMemoryAccessAligned, 4)
		ctx.StoreRegisterPointer(b, gcnSpec.OpSgpr0+details.Dst+i, val)
	}
}
