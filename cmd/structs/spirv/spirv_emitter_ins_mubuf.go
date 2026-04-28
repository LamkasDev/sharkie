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

	// We will accumulate the byte offset in this variable.
	byteOffset := ctx.GetConstId(ConstIdxUint0)

	// Load the format size and stride for this specific buffer from push constant.
	formatSize := ctx.LoadPushConstantValue(b, PushConstantTexelBuffer0FormatSize+details.Srsrc)
	formatStride := ctx.LoadPushConstantValue(b, PushConstantTexelBuffer0FormatStride+details.Srsrc)

	// stride * idx_vgpr (ignoring TIDinWave for now, only relevant if AddTidEnable=1).
	if details.Idxen {
		idxVgpr := ctx.LoadRegisterPointer(b, OpVgpr0+details.Vaddr)
		strideOffset := b.EmitIMul(typeUint, idxVgpr, formatStride)
		byteOffset = b.EmitIAdd(typeUint, byteOffset, strideOffset)
	}

	// off_vgpr (vector byte offset).
	if details.Offen {
		// As per spec located at VADDR (if Idxen=0) or VADDR+1 (if Idxen=1).
		voffsetReg := details.Vaddr
		if details.Idxen {
			voffsetReg += 1
		}
		offVgpr := ctx.LoadRegisterPointer(b, OpVgpr0+voffsetReg)
		byteOffset = b.EmitIAdd(typeUint, byteOffset, offVgpr)
	}

	// mem_offset (scalar byte offset).
	memOffset := ctx.GetOperandValue(b, details.Soffset, 0)
	byteOffset = b.EmitIAdd(typeUint, byteOffset, memOffset)

	// inst_offset (immediate offset).
	if details.Offset > 0 {
		instOffset := b.EmitConstantUint(typeUint, details.Offset)
		byteOffset = b.EmitIAdd(typeUint, byteOffset, instOffset)
	}

	// TexelCoord = byteOffset / formatSize.
	coord := b.EmitUDiv(typeUint, byteOffset, formatSize)

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
		ctx.StoreRegisterPointerMasked(b, OpVgpr0+details.Vdata+i, b.EmitBitcast(typeUint, compFloat))
	}
}
