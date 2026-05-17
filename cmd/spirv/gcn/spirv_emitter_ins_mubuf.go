package gcn

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func EmitMUBUF(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MubufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec4 := ctx.GetId(BlockContextIdTypeV4Float)
	typeImageBuffer := ctx.GetId(BlockContextIdTypeImageBuffer)

	// Load the texel buffer image from the descriptor set.
	texelBufferVar := ctx.GetTexelBufferVariable(details.Srsrc)
	image := b.EmitLoad(typeImageBuffer, texelBufferVar)

	// Load the format size and stride for this specific buffer from push constant.
	formatSize := ctx.LoadPushConstantValue(b, PushConstantTexelBuffer0FormatSize+details.Srsrc)

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
	memOffset := ctx.GetOperandUintValue(b, details.Soffset, 0)

	// Load buffer resource for stride and other params.
	res := structs.NewBufferResource(b, ctx, details.Srsrc)

	// Calculate final byte offset (buffer_offset in spec, DOES NOT include mem_offset).
	bufferOffset := structs.CalculateBufferOffset(b, ctx, res.Stride, res.SwizzleEn, res.ElementSize, res.IndexStride, res.AddTidEnable, index, offset)

	// Perform range check (requires bufferOffset and memOffset separately).
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idxenId := ctx.GetId(BlockContextIdFalse)
	if details.Idxen {
		idxenId = ctx.GetId(BlockContextIdTrue)
	}
	idxenOrAddTidEnable := b.EmitLogicalOr(typeBool, idxenId, res.AddTidEnable)
	outOfRange := structs.CalculateBufferRangeCheck(b, ctx, res, memOffset, index, offset, bufferOffset, idxenOrAddTidEnable)

	// Final byte address for coordinated access: buffer_offset + mem_offset.
	byteOffset := b.EmitIAdd(typeUint, bufferOffset, memOffset)

	// TexelCoord = byteOffset / formatSize.
	coord := b.EmitUDiv(typeUint, byteOffset, formatSize)

	// Fetch the formatted texel.
	fetchedVec4 := b.EmitImageFetch(typeVec4, image, coord, 0)

	// Determine how many components to store.
	var count uint32
	switch details.Op {
	case gcnSpec.MubufOpLoadFormatX, gcnSpec.MubufOpLoadDword:
		count = 1
	case gcnSpec.MubufOpLoadFormatXy, gcnSpec.MubufOpLoadDwordx2:
		count = 2
	case gcnSpec.MubufOpLoadFormatXyz, gcnSpec.MubufOpLoadDwordx3:
		count = 3
	case gcnSpec.MubufOpLoadFormatXyzw, gcnSpec.MubufOpLoadDwordx4:
		count = 4
	default:
		panic(fmt.Sprintf("unknown mubuf op %s", gcnSpec.Mnemotics[gcnSpec.EncMUBUF][details.Op]))
	}

	// If out of range, result is zero.
	for i := range count {
		// Extract X, Y, Z, or W.
		compFloat := b.EmitCompositeExtract(typeFloat, fetchedVec4, i)

		// Select between 0.0 and the fetched value.
		selectedFloat := b.EmitSelect(typeFloat, outOfRange, ctx.GetConstId(ConstIdFloat0), compFloat)

		// Store results back into VGPRs.
		ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+i, b.EmitBitcast(typeUint, selectedFloat))
	}

	/* formatId := b.EmitString(fmt.Sprintf("Mubuf 0x%X %%d: pos=(%%f, %%f, %%f, %%f)\n", instr.DwordOffset))
	vertexIndexId := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))

	px := b.EmitCompositeExtract(typeFloat, fetchedVec4, 0)
	py := b.EmitCompositeExtract(typeFloat, fetchedVec4, 1)
	pz := b.EmitCompositeExtract(typeFloat, fetchedVec4, 2)
	pw := b.EmitCompositeExtract(typeFloat, fetchedVec4, 3)

	b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1,
		formatId, vertexIndexId, px, py, pz, pw) */
}
