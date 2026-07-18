package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

func EmitMTBUF(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	panic("format conversion not supported")

	details := instr.Details.(*gcnSpec.MtbufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)

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

	// Fetch the components and unpack.
	fetchedVec4 := structs.EmitFormatUnpackHelper(b, ctx, res.BaseAddress, byteOffset, res.Dw3, uint32(details.Op))

	// Determine how many components to store.
	var count uint32
	switch details.Op {
	case gcnSpec.MtbufOpTbufferLoadFormatX:
		count = 1
	case gcnSpec.MtbufOpTbufferLoadFormatXy:
		count = 2
	case gcnSpec.MtbufOpTbufferLoadFormatXyz:
		count = 3
	case gcnSpec.MtbufOpTbufferLoadFormatXyzw:
		count = 4
	default:
		panic(fmt.Sprintf("unknown mubuf op %s", gcnSpec.Mnemotics[gcnSpec.EncMUBUF][details.Op]))
	}

	// Pre-extract all 4 components.
	compR := b.EmitCompositeExtract(typeFloat, fetchedVec4, 0)
	compG := b.EmitCompositeExtract(typeFloat, fetchedVec4, 1)
	compB := b.EmitCompositeExtract(typeFloat, fetchedVec4, 2)
	compA := b.EmitCompositeExtract(typeFloat, fetchedVec4, 3)

	// Put components in correct slots.
	for i := range count {
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

		// Select between 0.0 and the fetched value.
		selectedFloat := b.EmitSelect(typeFloat, outOfRange, ctx.GetConstId(ConstIdFloat0), compFloat)

		// Store results back into VGPRs.
		ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+i, b.EmitBitcast(typeUint, selectedFloat))
	}
}
