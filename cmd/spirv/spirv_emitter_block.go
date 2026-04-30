package spirv

import (
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
)

const (
	PushConstantUserDataAddress            = 0
	PushConstantOnionMemoryBaseAddress     = 1
	PushConstantGarlicMemoryBaseAddress    = 2
	PushConstantTexelBuffer0FormatSize     = 3
	PushConstantTexelBuffer1FormatSize     = 4
	PushConstantTexelBuffer2FormatSize     = 5
	PushConstantTexelBuffer3FormatSize     = 6
	PushConstantTexelBuffer0FormatStride   = 7
	PushConstantTexelBuffer1FormatStride   = 8
	PushConstantTexelBuffer2FormatStride   = 9
	PushConstantTexelBuffer3FormatStride   = 10
	PushConstantTexelBuffer0FormatElements = 11
	PushConstantTexelBuffer1FormatElements = 12
	PushConstantTexelBuffer2FormatElements = 13
	PushConstantTexelBuffer3FormatElements = 14
)

// emitBlock emits the SPIR-V for a single block.
func emitBlock(b *SpvBuilder, block *GcnShaderCfgBlock, ctx *SpirvBlockContext) {
	// Start current block.
	b.EmitLabel(ctx.GetLabelId(block.Id))

	// Declare variables in entry block.
	if block.DwordOffset == 0 {
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		idC0 := ctx.GetConstId(ConstIdUint0)

		// Load user data buffer address from the push constant.
		b.EmitString("load user data buffer address")
		idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
		ptrBase := ctx.LoadPushConstantValue(b, PushConstantUserDataAddress)

		// Load 16 user data registers into s0-s15.
		b.EmitString("load user data registers")
		stageOffset := gpu.GcnStageToUserDataOffset[ctx.Stage]
		for i := range uint32(16) {
			ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, ctx.GetConstId(SpirvId(stageOffset+i)))
			value := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ptr, spec.SpvMemoryAccessAligned, 4)
			ctx.SetGcnSgprId(b, i, value)
		}

		// Load vertex index and instance index into v0 and v1.
		if ctx.Stage == GcnShaderStageVertex {
			b.EmitString("load vertex and instance index")
			v0 := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))
			ctx.SetGcnVgprId(b, 0, v0)
			v1 := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdInstanceIndex))
			ctx.SetGcnVgprId(b, 1, v1)
		}

		// Initialize EXEC and VCC.
		// EXEC is initialized to the subgroup's active mask.
		b.EmitString("initialize exec and vcc")
		typeV4Uint := ctx.GetId(BlockContextIdTypeV4Uint)
		idC3 := ctx.GetConstId(ConstIdUint3) // Subgroup
		ballot := b.EmitGroupNonUniformBallot(typeV4Uint, idC3, ctx.GetId(BlockContextIdTrue))
		execLo := b.EmitCompositeExtract(typeUint, ballot, 0)
		execHi := b.EmitCompositeExtract(typeUint, ballot, 1)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, execLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, execHi)

		// VCC is initialized to 0.
		b.EmitString("set vcc to 0")
		ctx.StoreRegisterPointer(b, gcnSpec.OpVccLo, idC0)
		ctx.StoreRegisterPointer(b, gcnSpec.OpVccHi, idC0)
	}

	// Reset condition ID.
	ctx.GcnConditionId = ctx.GetId(BlockContextIdFalse)

	// Emit instructions for current block.
	for i := range block.Instructions {
		emitInstruction(b, &block.Instructions[i], ctx)
	}

	// Terminate current block.
	switch block.Term {
	case TermCBranch:
		EmitConditionalBranch(b, block, ctx)
	case TermBranch, TermFallthrough:
		if len(block.Successors) > 0 {
			b.EmitBranch(ctx.GetLabelId(block.Successors[0]))
		} else {
			b.EmitUnreachable()
		}
	case TermEndpgm, TermExpDone:
		if ctx.Stage == GcnShaderStageVertex {
			formatId := b.EmitString("Vertex %d: pos=(%f, %f, %f, %f)\n")
			typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
			typeFloat := ctx.GetId(BlockContextIdTypeFloat)
			posId := b.EmitLoad(typeV4Float, ctx.GetId(BlockContextIdPosOut))
			vertexIndexId := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))

			px := b.EmitCompositeExtract(typeFloat, posId, 0)
			py := b.EmitCompositeExtract(typeFloat, posId, 1)
			pz := b.EmitCompositeExtract(typeFloat, posId, 2)
			pw := b.EmitCompositeExtract(typeFloat, posId, 3)

			b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1,
				formatId, vertexIndexId, px, py, pz, pw)
		} else if ctx.Stage == GcnShaderStageFragment {
			formatId := b.EmitString("Vertex %d: color=(%f, %f, %f, %f)\n")
			typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
			typeFloat := ctx.GetId(BlockContextIdTypeFloat)
			colorId := b.EmitLoad(typeV4Float, ctx.GetId(BlockContextIdColorOut0))
			vertexIndexId := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))

			cx := b.EmitCompositeExtract(typeFloat, colorId, 0)
			cy := b.EmitCompositeExtract(typeFloat, colorId, 1)
			cz := b.EmitCompositeExtract(typeFloat, colorId, 2)
			cw := b.EmitCompositeExtract(typeFloat, colorId, 3)

			b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1,
				formatId, vertexIndexId, cx, cy, cz, cw)
		}
		// ctx.EmitDebugPrintRegisters(b)
		b.EmitReturn()
	default:
		b.EmitReturn()
	}
}
