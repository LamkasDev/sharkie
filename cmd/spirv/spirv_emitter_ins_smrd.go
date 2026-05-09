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
		emitSMRDLoadScalar(b, instr, ctx, 1)
	case gcnSpec.SmrdOpLoadDwordx2:
		emitSMRDLoadScalar(b, instr, ctx, 2)
	case gcnSpec.SmrdOpLoadDwordx4:
		emitSMRDLoadScalar(b, instr, ctx, 4)
	case gcnSpec.SmrdOpLoadDwordx8:
		emitSMRDLoadScalar(b, instr, ctx, 8)
	case gcnSpec.SmrdOpLoadDwordx16:
		emitSMRDLoadScalar(b, instr, ctx, 16)
	case gcnSpec.SmrdOpBufferLoadDword:
		emitSMRDLoadBuffer(b, instr, ctx, 1)
	case gcnSpec.SmrdOpBufferLoadDwordx2:
		emitSMRDLoadBuffer(b, instr, ctx, 2)
	case gcnSpec.SmrdOpBufferLoadDwordx4:
		emitSMRDLoadBuffer(b, instr, ctx, 4)
	case gcnSpec.SmrdOpBufferLoadDwordx8:
		emitSMRDLoadBuffer(b, instr, ctx, 8)
	case gcnSpec.SmrdOpBufferLoadDwordx16:
		emitSMRDLoadBuffer(b, instr, ctx, 16)
	default:
		panic(fmt.Sprintf("unknown smrd op %s", gcnSpec.Mnemotics[gcnSpec.EncSMRD][details.Op]))
	}
}

func emitSMRDLoadScalar(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext, count uint32) {
	details := instr.Details.(*gcnSpec.SmrdDetails)

	// S_LOAD_* uses a 64-bit byte base address from SGPR[SBASE*2].
	base := details.Base * 2
	lo, hi := ctx.GetOperand64Value(b, gcnSpec.OpSgpr0+base, 0)
	base64 := ctx.Pack64(b, lo, hi)

	emitSMRDLoadFromBase(b, instr, ctx, count, base64)
}

func emitSMRDLoadBuffer(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext, count uint32) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	idFFFF := ctx.GetConstId(ConstIdUintFFFF)

	// S_BUFFER_LOAD_* uses a 4-SGPR buffer resource constant.
	// Base address is {SGPR[SBASE*2+1][15:0], SGPR[SBASE*2]}.
	lo := ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Base*2, 0)
	hi := ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Base*2+1, 0)
	hi = b.EmitBitwiseAnd(typeUint, hi, idFFFF)
	base64 := ctx.Pack64(b, lo, hi)

	emitSMRDLoadFromBase(b, instr, ctx, count, base64)
}

func emitSMRDLoadFromBase(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext, count uint32, base64 SpirvId) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	idNot3 := ctx.GetConstId(ConstId64UintNot3)

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
		// Offset is an SGPR index containing an unsigned byte offset.
		offsetVal := ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Offset, 0)
		byteOffset = offsetVal
	}

	// m_addr = (base + m_offset) & ~0x3
	byteOffset64 := b.EmitUConvert(typeUint64, byteOffset)
	addr64 := b.EmitIAdd(typeUint64, base64, byteOffset64)
	addr64Aligned := b.EmitBitwiseAnd(typeUint64, addr64, idNot3)

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
