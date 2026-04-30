package spirv

import (
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

// EmitBranch emits OpBranch.
func (b *SpvBuilder) EmitBranch(targetLabel SpirvId) {
	b.instr(&b.code, spec.SpvOpBranch, uint32(targetLabel))
}

// EmitBranchConditional emits OpBranchConditional.
func (b *SpvBuilder) EmitBranchConditional(condID, trueLabel, falseLabel SpirvId) {
	b.instr(&b.code, spec.SpvOpBranchConditional, uint32(condID), uint32(trueLabel), uint32(falseLabel))
}

// EmitSelectionMerge emits OpSelectionMerge (must appear immediately before the OpBranchConditional or OpSwitch it governs).
func (b *SpvBuilder) EmitSelectionMerge(mergeBlock SpirvId, selectionControl uint32) {
	b.instr(&b.code, spec.SpvOpSelectionMerge, uint32(mergeBlock), selectionControl)
}

// EmitLoopMerge emits OpLoopMerge (must appear immediately before the branch instruction that closes the loop header).
func (b *SpvBuilder) EmitLoopMerge(mergeBlock, continueBlock SpirvId, loopControl uint32) {
	b.instr(&b.code, spec.SpvOpLoopMerge, uint32(mergeBlock), uint32(continueBlock), loopControl)
}

// EmitConditionalBranch handles TermCBranch.
// OpLoopMerge (loop headers) or OpSelectionMerge (selections) must appear immediately before the OpBranchConditional instruction.
func EmitConditionalBranch(b *SpvBuilder, block *GcnShaderCfgBlock, ctx *SpirvBlockContext) {
	if block.IsLoopHeader {
		mergeLabelId := ctx.GetLabelId(block.MergeBlockId)
		continueLabelId := ctx.GetLabelId(block.ContinueBlockId)
		b.EmitLoopMerge(mergeLabelId, continueLabelId, spec.SpvLoopControlNone)
	} else if block.MergeBlockId >= 0 {
		b.EmitSelectionMerge(ctx.GetLabelId(block.MergeBlockId), spec.SpvSelectionControlNone)
	}

	falseLabelId := ctx.GetLabelId(block.Successors[0]) // fall-through.
	trueLabelId := ctx.GetLabelId(block.Successors[1])  // branch target.
	b.EmitBranchConditional(ctx.GcnConditionId, trueLabelId, falseLabelId)
}
