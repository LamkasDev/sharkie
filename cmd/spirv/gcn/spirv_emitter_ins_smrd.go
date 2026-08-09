package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitSMRD(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	count := SmrdLoadDwordCount(details.Op)
	switch {
	case details.Op <= gcnSpec.SmrdOpLoadDwordx16:
		emitSMRDLoadScalar(b, instr, ctx, count)
	case details.Op <= gcnSpec.SmrdOpBufferLoadDwordx16:
		emitSMRDLoadBuffer(b, instr, ctx, count)
	default:
		panic(fmt.Sprintf("unknown smrd op %s", gcnSpec.Mnemotics[gcnSpec.EncSMRD][details.Op]))
	}
}

func SmrdLoadDwordCount(op uint32) uint32 {
	switch op {
	case gcnSpec.SmrdOpLoadDword, gcnSpec.SmrdOpBufferLoadDword:
		return 1
	case gcnSpec.SmrdOpLoadDwordx2, gcnSpec.SmrdOpBufferLoadDwordx2:
		return 2
	case gcnSpec.SmrdOpLoadDwordx4, gcnSpec.SmrdOpBufferLoadDwordx4:
		return 4
	case gcnSpec.SmrdOpLoadDwordx8, gcnSpec.SmrdOpBufferLoadDwordx8:
		return 8
	case gcnSpec.SmrdOpLoadDwordx16, gcnSpec.SmrdOpBufferLoadDwordx16:
		return 16
	default:
		return 0
	}
}

func emitSMRDLoadScalar(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext, count uint32) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	// S_LOAD_* uses a 64-bit byte base address from SGPR[SBASE*2].
	base := details.Base * 2
	lo, hi := ctx.GetOperand64Value(b, gcnSpec.OpSgpr0+base, 0)
	base64 := ctx.Pack64(b, lo, hi)

	// Calculate offset in bytes.
	var byteOffset SpirvId
	if details.ImmOff {
		if instr.HasLiteral {
			// 64-bit SMRD: offset is a 32-bit byte offset.
			byteOffset = b.EmitConstantUint(typeUint, instr.Literal)
		} else {
			// 32-bit SMRD: offset is an 8-bit unsigned dword offset.
			byteOffset = b.EmitConstantUint(typeUint, details.Offset*4)
		}
	} else {
		// Offset is an SGPR index containing an unsigned byte offset.
		byteOffset = ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Offset, 0)
	}

	// S_LOAD_DWORD does not have out-of-bounds checks, so always false.
	outOfRange := ctx.GetId(BlockContextIdFalse)

	emitSMRDLoadFromBase(b, instr, ctx, count, base64, byteOffset, outOfRange)
}

func emitSMRDLoadBuffer(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext, count uint32) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idFFFF := ctx.GetConstId(ConstIdUintFFFF)

	// S_BUFFER_LOAD_* uses a 4-SGPR buffer resource constant.
	// Base address is {SGPR[SBASE*2+1][15:0], SGPR[SBASE*2]}.
	lo := ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Base*2, 0)
	hi := ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Base*2+1, 0)
	hiPacked := b.EmitBitwiseAnd(typeUint, hi, idFFFF)
	base64 := ctx.Pack64(b, lo, hiPacked)

	// Stride is dw1[29:16].
	dw2 := ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Base*2+2, 0)
	stride := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, hi, ctx.GetConstId(ConstIdUint16)), ctx.GetConstId(ConstIdUint3FFF))
	numRecords := dw2

	// Calculate offset in bytes.
	var byteOffset SpirvId
	if details.ImmOff {
		if instr.HasLiteral {
			// 64-bit SMRD: offset is a 32-bit byte offset.
			byteOffset = b.EmitConstantUint(typeUint, instr.Literal)
		} else {
			// 32-bit SMRD: offset is an 8-bit unsigned dword offset.
			byteOffset = b.EmitConstantUint(typeUint, details.Offset*4)
		}
	} else {
		byteOffset = ctx.GetOperandUintValue(b, gcnSpec.OpSgpr0+details.Offset, 0)
	}

	// Buffer resource bounds check.
	// Stride is zero: clamp if (offset >= num_records).
	// Stride is non-zero: clamp if (offset >= (stride * num_records)).
	strideIsZero := b.EmitIEqual(typeBool, stride, ctx.GetConstId(ConstIdUint0))
	outOfRangeZero := b.EmitUGreaterThanEqual(typeBool, byteOffset, numRecords)
	outOfRangeNonZero := b.EmitUGreaterThanEqual(typeBool, byteOffset, b.EmitIMul(typeUint, stride, numRecords))
	outOfRange := b.EmitSelect(typeBool, strideIsZero, outOfRangeZero, outOfRangeNonZero)

	emitSMRDLoadFromBase(b, instr, ctx, count, base64, byteOffset, outOfRange)
}

func emitSMRDLoadFromBase(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext, count uint32, base64, byteOffset, outOfRange SpirvId) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	idNot3 := ctx.GetConstId(ConstId64UintNot3)

	// m_addr = (base + m_offset) & ~0x3
	byteOffset64 := b.EmitUConvert(typeUint64, byteOffset)
	addr64 := b.EmitIAdd(typeUint64, base64, byteOffset64)
	addr64Aligned := b.EmitBitwiseAnd(typeUint64, addr64, idNot3)

	// Translate and cast to pointer.
	translatedAddr64 := ctx.TranslateAddress(b, addr64Aligned)
	ptrBase := b.EmitBitcast(idPtrPsbUint, translatedAddr64)

	// Perform bounds checking.
	isValid := b.EmitLogicalNot(typeBool, outOfRange)
	isValidAddress := b.EmitINotEqual(typeBool, translatedAddr64, ctx.GetConstId(ConstId64Uint0))
	isValid = b.EmitLogicalAnd(typeBool, isValid, isValidAddress)
	validLabel := b.AllocId()
	invalidLabel := b.AllocId()
	mergeLabel := b.AllocId()

	b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(isValid, validLabel, invalidLabel)

	// Load dwords.
	b.EmitLabel(validLabel)
	for i := range count {
		ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, ctx.GetConstId(ConstIdUint0+SpirvId(i)))
		val := b.EmitLoad(typeUint, ptr, spec.SpvMemoryAccessAligned, 4)
		ctx.StoreRegisterPointer(b, gcnSpec.OpSgpr0+details.Dst+i, val)
	}
	b.EmitBranch(mergeLabel)

	// If the memory address is out-of-range (clamped), the operation is not performed.
	b.EmitLabel(invalidLabel)
	ctx.EmitDebugPrintf(b, "TranslateAddress failed (0x%lx): unmapped address 0x%lx + base 0x%lx + offset 0x%x\n",
		b.EmitConstantUint64(typeUint64, uint64(ctx.Address)), addr64Aligned, base64, byteOffset)
	b.EmitBranch(mergeLabel)

	b.EmitLabel(mergeLabel)
}
