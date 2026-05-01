package spirv

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func emitMUBUF(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.MubufDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeVec4 := ctx.GetId(BlockContextIdTypeV4Float)
	typeImageBuffer := ctx.GetId(BlockContextIdTypeImageBuffer)

	// Load the texel buffer image from the descriptor set.
	texelBufferVar := ctx.GetTexelBufferVariable(details.Srsrc)
	image := b.EmitLoad(typeImageBuffer, texelBufferVar)

	// We will accumulate the byte offset in this variable.
	byteOffset := ctx.GetConstId(ConstIdUint0)

	// Load the format size and stride for this specific buffer from push constant.
	formatSize := ctx.LoadPushConstantValue(b, PushConstantTexelBuffer0FormatSize+details.Srsrc)
	formatStride := ctx.LoadPushConstantValue(b, PushConstantTexelBuffer0FormatStride+details.Srsrc)

	// stride * idx_vgpr (ignoring TIDinWave for now, only relevant if AddTidEnable=1).
	if details.Idxen {
		idxVgpr := ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+details.Vaddr)
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
		offVgpr := ctx.LoadRegisterPointer(b, gcnSpec.OpVgpr0+voffsetReg)
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

	for i := range count {
		// Extract X, Y, Z, or W.
		compFloat := b.EmitCompositeExtract(typeFloat, fetchedVec4, i)

		// Store results back into VGPRs.
		ctx.StoreRegisterPointerMasked(b, gcnSpec.OpVgpr0+details.Vdata+i, b.EmitBitcast(typeUint, compFloat))
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
