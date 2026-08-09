package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitDS(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.DsDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)

	// Read base address from VGPR and M0 for bounds checking.
	addr := ctx.GetOperandUintValue(b, gcnSpec.OpVgpr0+details.Addr, 0)
	m0 := b.EmitLoad(typeUint, ctx.GetGcnSpecialId(GcnSpecIdM0))

	// Offsets are in bytes, scaled by 8 for B64 instructions.
	offset0 := b.EmitConstantUint(typeUint, details.Offset0*8)
	offset1 := b.EmitConstantUint(typeUint, details.Offset1*8)
	addr0 := b.EmitIAdd(typeUint, addr, offset0)
	addr1 := b.EmitIAdd(typeUint, addr, offset1)

	// Determine base pointer based on memory space.
	var ptrBase SpirvId
	if details.Gds {
		gdsBase := ctx.LoadPushConstantValue(b, PushConstantGdsMemoryBaseAddress)
		ptrBase = b.EmitConvertUToPtr(ctx.GetId(BlockContextIdPtrPsbUint), gdsBase)
	} else {
		ptrBase = ctx.GetId(BlockContextIdLdsArray)
	}

	// Helper function to emit bounded memory access for an address.
	emitAccess := func(targetAddr SpirvId, isAddr1 bool) {
		addrPlus8 := b.EmitIAdd(typeUint, targetAddr, b.EmitConstantUint(typeUint, 8))

		// Perform bounds checking.
		isValid := b.EmitULessThanEqual(typeBool, addrPlus8, m0)
		validLabel := b.AllocId()
		mergeLabel := b.AllocId()

		b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(isValid, validLabel, mergeLabel)

		b.EmitLabel(validLabel)

		// Convert byte address to uint32 index (targetAddr >> 2).
		idxLo := b.EmitShiftRightLogical(typeUint, targetAddr, b.EmitConstantUint(typeUint, 2))
		idxHi := b.EmitIAdd(typeUint, idxLo, ctx.GetConstId(ConstIdUint1))

		// Fetch the pointers for the low and high 32-bit halves.
		var ptrType, ptrLo, ptrHi SpirvId
		if details.Gds {
			ptrType = ctx.GetId(BlockContextIdPtrPsbUint)
			ptrLo = b.EmitPtrAccessChain(ptrType, ptrBase, idxLo)
			ptrHi = b.EmitPtrAccessChain(ptrType, ptrBase, idxHi)
		} else {
			// The first index directly accesses the array element in SPIR-V.
			ptrType = b.EmitTypePointer(spec.SpvStorageWorkgroup, typeUint)
			ptrLo = b.EmitAccessChain(ptrType, ptrBase, idxLo)
			ptrHi = b.EmitAccessChain(ptrType, ptrBase, idxHi)
		}

		switch details.Op {
		case gcnSpec.DsOpRead2B64:
			valLo := b.EmitLoad(typeUint, ptrLo, spec.SpvMemoryAccessAligned, 4)
			valHi := b.EmitLoad(typeUint, ptrHi, spec.SpvMemoryAccessAligned, 4)

			destBase := details.Vdst
			if isAddr1 {
				destBase += 2
			}
			ctx.StoreRegisterPointer(b, gcnSpec.OpVgpr0+destBase, valLo)
			ctx.StoreRegisterPointer(b, gcnSpec.OpVgpr0+destBase+1, valHi)
		case gcnSpec.DsOpWrite2B64:
			srcBase := details.Data0
			if isAddr1 {
				srcBase = details.Data1
			}
			valLo := ctx.GetOperandUintValue(b, gcnSpec.OpVgpr0+srcBase, 0)
			valHi := ctx.GetOperandUintValue(b, gcnSpec.OpVgpr0+srcBase+1, 0)

			b.EmitStore(ptrLo, valLo, spec.SpvMemoryAccessAligned, 4)
			b.EmitStore(ptrHi, valHi, spec.SpvMemoryAccessAligned, 4)
		}

		b.EmitBranch(mergeLabel)
		b.EmitLabel(mergeLabel)
	}

	switch details.Op {
	case gcnSpec.DsOpRead2B64, gcnSpec.DsOpWrite2B64:
		emitAccess(addr0, false)
		emitAccess(addr1, true)
	default:
		panic(fmt.Sprintf("unknown ds op %d", details.Op))
	}
}
