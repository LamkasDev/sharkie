package spirv

import (
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

type SpirvShaderContext struct {
	NumThreads [3]uint32
}

type SpirvShader struct {
	Stage   GcnShaderStage
	Address uintptr
	Code    []uint32
}

func NewSpirvShader(shader *GcnShader, ctx SpirvShaderContext) (*SpirvShader, error) {
	b := NewSpvBuilder()

	// Capabilities.
	b.EmitCapability(SpvCapShader)
	b.EmitMemoryModel(SpvAddrModelLogical, SpvMemModelGLSL450)

	// Common types.
	idVoid := b.EmitTypeVoid()
	idBool := b.EmitTypeBool()
	idUint := b.EmitTypeInt(32, false)
	idFnType := b.EmitTypeFunction(idVoid)

	// Push constants.
	idConst16 := b.EmitConstantUint(idUint, 16)
	idArrUd := b.EmitTypeArray(idUint, idConst16)
	idUd := b.EmitTypeStruct(idArrUd)
	idPtrPc := b.EmitTypePointer(SpvStoragePushConstant, idUd)
	idPtrPcUint := b.EmitTypePointer(SpvStoragePushConstant, idUint)
	_ = idPtrPcUint

	// Annotations for the push-constant block.
	b.EmitDecorate(idArrUd, SpvDecorationArrayStride, 4)
	b.EmitDecorate(idUd, SpvDecorationBlock)
	b.EmitMemberDecorate(idUd, 0, SpvDecorationOffset, 0)

	// Global push-constant variable.
	idPCVar := b.EmitVariable(idPtrPc, SpvStoragePushConstant)
	_ = idPCVar

	// Stub boolean constant (condition in untranslated conditional branches).
	idFalse := b.EmitConstantFalse(idBool)

	// Entry point.
	idMain := b.AllocId()
	b.EmitEntryPoint(GncStageToSpvExecModel[shader.Stage], idMain, "main")

	// Execution modes.
	switch shader.Stage {
	case GcnShaderStageFragment:
		b.EmitExecutionMode(idMain, SpvExecModeOriginUpperLeft)
	case GcnShaderStageCompute:
		nt := ctx.NumThreads
		for i := range nt {
			if nt[i] == 0 {
				nt[i] = 1
			}
		}
		b.EmitExecutionMode(idMain, SpvExecModeLocalSize, nt[0], nt[1], nt[2])
	}

	// Pre-allocate SPIR-V labels ID for GCN CFG blocks.
	labelID := make([]uint32, len(shader.Cfg.Blocks))
	for i := range shader.Cfg.Blocks {
		labelID[i] = b.AllocId()
	}

	// Function body.
	b.EmitFunction(idVoid, SpvFunctionControlNone, idFnType, idMain)

	// Emit reachable blocks in reverse post-order (entry block first).
	rpoBlockIds := shader.Cfg.ReversePostOrder()
	emittedBlockIds := make([]bool, len(shader.Cfg.Blocks))
	for _, blockId := range rpoBlockIds {
		emitBlock(b, &shader.Cfg.Blocks[blockId], labelID, idFalse)
		emittedBlockIds[blockId] = true
	}

	// Emit any unreachable blocks.
	for i := range shader.Cfg.Blocks {
		if !emittedBlockIds[i] {
			b.EmitLabel(labelID[i])
			b.EmitUnreachable()
		}
	}

	// Andddd we're done :)
	b.EmitFunctionEnd()

	return &SpirvShader{
		Address: shader.Address,
		Stage:   shader.Stage,
		Code:    b.Assemble(),
	}, nil
}

// emitBlock emits the SPIR-V for a single block.
func emitBlock(b *SpvBuilder, block *GcnShaderCfgBlock, labelID []uint32, idFalse uint32) {
	b.EmitLabel(labelID[block.Id])

	// TODO: emit other instructions.

	switch block.Term {
	case TermCBranch:
		emitConditionalBranch(b, block, labelID, idFalse)
	case TermBranch, TermFallthrough:
		if len(block.Successors) > 0 {
			b.EmitBranch(labelID[block.Successors[0]])
		} else {
			b.EmitUnreachable()
		}
	case TermEndpgm, TermExpDone:
		b.EmitReturn()
	default:
		b.EmitReturn()
	}
}

// emitConditionalBranch handles TermCBranch.
// OpLoopMerge (loop headers) or OpSelectionMerge (selections) must appear immediately before the OpBranchConditional instruction.
func emitConditionalBranch(b *SpvBuilder, block *GcnShaderCfgBlock, labelId []uint32, idFalse uint32) {
	if block.IsLoopHeader {
		mergeLabelId := labelId[block.MergeBlockId]
		continueLabelId := labelId[block.ContinueBlockId]
		b.EmitLoopMerge(mergeLabelId, continueLabelId, SpvLoopControlNone)
	} else if block.MergeBlockId >= 0 {
		b.EmitSelectionMerge(labelId[block.MergeBlockId], SpvSelectionControlNone)
	}

	// TODO: we'll need to build the actual condition here.

	falseLabelId := labelId[block.Successors[0]] // fall-through.
	trueLabelId := labelId[block.Successors[1]]  // branch target.
	b.EmitBranchConditional(idFalse, trueLabelId, falseLabelId)
}
