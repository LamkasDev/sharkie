package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

func EmitMUBUF(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MubufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec4 := ctx.GetId(BlockContextIdTypeV4Float)

	// index = (inst_idxen ? vgpr_index : 0).
	index := ctx.GetConstId(ConstIdUint0)
	if details.Idxen {
		index = ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr)
	}

	// offset = (inst_offen ? vgpr_offset : 0) + inst_offset.
	offset := ctx.GetConstId(ConstIdUint0)
	if details.Offen {
		voffsetReg := details.Vaddr
		if details.Idxen {
			voffsetReg += 1
		}
		offset = ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+voffsetReg)
	}

	// inst_offset (immediate offset).
	if details.Offset > 0 {
		instOffset := b.EmitConstantUint(typeUint, details.Offset)
		offset = b.EmitIAdd(typeUint, offset, instOffset)
	}

	// mem_offset (scalar byte offset, aka sgpr_offset).
	sgprOffset := ctx.GetOperandUintValue(b, details.Soffset, 0)

	// Load buffer resource for stride and other params.
	res := structs.NewBufferResource(b, ctx, details.Srsrc)

	// Calculate buffer offset.
	bufferOffset := structs.CalculateBufferOffset(b, ctx, res, index, offset)

	// Perform range check.
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idxenId := ctx.GetId(BlockContextIdFalse)
	if details.Idxen {
		idxenId = ctx.GetId(BlockContextIdTrue)
	}
	idxenOrAddTidEnable := b.EmitLogicalOr(typeBool, idxenId, res.AddTidEnable)
	outOfRange := structs.CalculateBufferRangeCheck(b, ctx, res, index, offset, bufferOffset, sgprOffset, idxenOrAddTidEnable)

	// Byte address for coordinated access (buffer_offset + mem_offset).
	byteOffset := b.EmitIAdd(typeUint, bufferOffset, sgprOffset)

	// Calculate 64-bit byte offset for address translation.
	typeUint64 := ctx.GetId(BlockContextIdTypeUint64)
	byteOffset64 := b.EmitUConvert(typeUint64, byteOffset)

	bindingIndex := ctx.StaticLayout[instr].BindingIndex
	dataFormat, numFormat := ctx.GetMubufFormat(uint32(instr.DwordOffset))

	// Determine operation type and how many components to process.
	var count uint32
	isStore := false
	switch details.Op {
	case gcnSpec.MubufOpLoadFormatX, gcnSpec.MubufOpLoadDword:
		count = 1
	case gcnSpec.MubufOpLoadFormatXy, gcnSpec.MubufOpLoadDwordx2:
		count = 2
	case gcnSpec.MubufOpLoadFormatXyz, gcnSpec.MubufOpLoadDwordx3:
		count = 3
	case gcnSpec.MubufOpLoadFormatXyzw, gcnSpec.MubufOpLoadDwordx4:
		count = 4
	case gcnSpec.MubufOpStoreFormatX, gcnSpec.MubufOpStoreDword:
		count = 1
		isStore = true
	case gcnSpec.MubufOpStoreFormatXy, gcnSpec.MubufOpStoreDwordx2:
		count = 2
		isStore = true
	case gcnSpec.MubufOpStoreFormatXyz, gcnSpec.MubufOpStoreDwordx3:
		count = 3
		isStore = true
	case gcnSpec.MubufOpStoreFormatXyzw, gcnSpec.MubufOpStoreDwordx4:
		count = 4
		isStore = true
	default:
		panic(fmt.Sprintf("unknown mubuf op %s", gcnSpec.Mnemotics[gcnSpec.EncMUBUF][details.Op]))
	}

	if isStore {
		// Gather registers into Vec4.
		var comps [4]SpirvId
		for i := 0; i < 4; i++ {
			if i < int(count) {
				rawVgpr := ctx.GetOperandUintValue(b, gcnSpec.OpVgpr0+details.Vdata+uint32(i), 0)
				comps[i] = b.EmitBitcast(typeFloat, rawVgpr)
			} else {
				if i == 3 {
					comps[i] = ctx.GetConstId(ConstIdFloat1)
				} else {
					comps[i] = ctx.GetConstId(ConstIdFloat0)
				}
			}
		}
		storeVec4 := b.EmitCompositeConstruct(typeVec4, comps[0], comps[1], comps[2], comps[3])

		// Branch around the store if we are out of bounds.
		storeLabel := b.AllocId()
		mergeLabel := b.AllocId()
		b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
		b.EmitBranchConditional(outOfRange, mergeLabel, storeLabel)

		// Store values.
		b.EmitLabel(storeLabel)
		structs.EmitFormatPackHelper(b, ctx, bindingIndex, byteOffset64, dataFormat, numFormat, uint32(details.Op), storeVec4)
		b.EmitBranch(mergeLabel)

		// Merge.
		b.EmitLabel(mergeLabel)
	} else {
		// Fetch the components and unpack.
		fetchedVec4 := structs.EmitFormatUnpackHelper(b, ctx, bindingIndex, byteOffset64, dataFormat, numFormat, outOfRange, uint32(details.Op))

		// Pre-extract all 4 components.
		compR := b.EmitCompositeExtract(typeFloat, fetchedVec4, 0)
		compG := b.EmitCompositeExtract(typeFloat, fetchedVec4, 1)
		compB := b.EmitCompositeExtract(typeFloat, fetchedVec4, 2)
		compA := b.EmitCompositeExtract(typeFloat, fetchedVec4, 3)

		// Put components in correct slots.
		for i := uint32(0); i < count; i++ {
			// Extract destination selector for current channel.
			shiftAmount := ctx.GetConstId(ConstIdUint0 + SpirvId(i*3))
			shiftedDword := b.EmitShiftRightLogical(typeUint, res.Dw3, shiftAmount)
			dstSel := b.EmitBitwiseAnd(typeUint, shiftedDword, b.EmitConstantUint(typeUint, 7))

			// Generate conditions for selectors.
			is0 := b.EmitIEqual(typeBool, dstSel, ctx.GetConstId(ConstIdUint0))
			is1 := b.EmitIEqual(typeBool, dstSel, ctx.GetConstId(ConstIdUint1))
			isR := b.EmitIEqual(typeBool, dstSel, ctx.GetConstId(ConstIdUint4))
			isG := b.EmitIEqual(typeBool, dstSel, ctx.GetConstId(ConstIdUint5))
			isB := b.EmitIEqual(typeBool, dstSel, ctx.GetConstId(ConstIdUint6))

			// Build selection chain (default to A).
			compFloat := compA
			compFloat = b.EmitSelect(typeFloat, isB, compB, compFloat)
			compFloat = b.EmitSelect(typeFloat, isG, compG, compFloat)
			compFloat = b.EmitSelect(typeFloat, isR, compR, compFloat)
			compFloat = b.EmitSelect(typeFloat, is1, ctx.GetConstId(ConstIdFloat1), compFloat)
			compFloat = b.EmitSelect(typeFloat, is0, ctx.GetConstId(ConstIdFloat0), compFloat)

			// Store results back into VGPRs.
			ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+i, b.EmitBitcast(typeUint, compFloat))
		}
	}
}
