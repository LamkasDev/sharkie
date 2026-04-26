package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMUBUF(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*MubufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec4 := ctx.GetId(BlockContextIdTypeV4Float)
	typeImageBuffer := ctx.GetId(BlockContextIdTypeImageBuffer)

	// Load the texel buffer image from the descriptor set.
	texelBufferVar := ctx.GetTexelBufferVariable(details.Srsrc)
	image := b.EmitLoad(typeImageBuffer, texelBufferVar)

	// Calculate coordinate (Index).
	// Address = (Base + Soffset) + (InstOffset + Voffset + Vindex * Stride)
	// If both idxen and offen are 1, VADDR is Voffset, VADDR+1 is Vindex.
	var coord uint32
	if details.Idxen && details.Offen {
		coord = ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr+1)
	} else if details.Idxen {
		coord = ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
	} else {
		coord = ctx.GetConstId(ConstIdxUint0)
	}

	// Add Soffset if it is an index or if we want to support it as a base offset.
	// For now we treat it as an index addition, though it's technically a byte offset.
	soffset := ctx.GetOperandValue(b, details.Soffset, 0)
	coord = b.EmitIAdd(typeUint, coord, soffset)

	// Add InstOffset.
	if details.Offset > 0 {
		coord = b.EmitIAdd(typeUint, coord, b.EmitConstantUint(typeUint, details.Offset))
	}

	// Fetch the formatted texel (OpImageFetch handles the bounds check and converts the raw memory).
	fetchedVec4 := b.EmitImageFetch(typeVec4, image, coord)

	// Determine how many components to store.
	var count uint32
	switch details.Op {
	case MubufOpLoadFormatX, MubufOpLoadDword:
		count = 1
	case MubufOpLoadFormatXy, MubufOpLoadDwordx2:
		count = 2
	case MubufOpLoadFormatXyz, MubufOpLoadDwordx3:
		count = 3
	case MubufOpLoadFormatXyzw, MubufOpLoadDwordx4:
		count = 4
	default:
		panic(fmt.Sprintf("unknown mubuf op %s", Mnemotics[EncMUBUF][details.Op]))
	}

	for i := range count {
		// Extract X, Y, Z, or W.
		compFloat := b.EmitCompositeExtract(typeFloat, fetchedVec4, i)

		// Store results back into VGPRs.
		ctx.StoreRegisterPointer(b, OpVgpr0+details.Vdata+i, b.EmitBitcast(typeUint, compFloat))
	}
}
