package spirv

import (
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

// EmitBranch emits OpBranch.
func (b *SpvBuilder) EmitBranch(targetLabel uint32) {
	b.instr(&b.code, SpvOpBranch, targetLabel)
}

// EmitBranchConditional emits OpBranchConditional.
func (b *SpvBuilder) EmitBranchConditional(condID, trueLabel, falseLabel uint32) {
	b.instr(&b.code, SpvOpBranchConditional, condID, trueLabel, falseLabel)
}

// EmitSelectionMerge emits OpSelectionMerge (must appear immediately before the OpBranchConditional or OpSwitch it governs).
func (b *SpvBuilder) EmitSelectionMerge(mergeBlock, selectionControl uint32) {
	b.instr(&b.code, SpvOpSelectionMerge, mergeBlock, selectionControl)
}

// EmitLoopMerge emits OpLoopMerge (must appear immediately before the branch instruction that closes the loop header).
func (b *SpvBuilder) EmitLoopMerge(mergeBlock, continueBlock, loopControl uint32) {
	b.instr(&b.code, SpvOpLoopMerge, mergeBlock, continueBlock, loopControl)
}

// emitConditionalBranch handles TermCBranch.
// OpLoopMerge (loop headers) or OpSelectionMerge (selections) must appear immediately before the OpBranchConditional instruction.
func emitConditionalBranch(b *SpvBuilder, block *GcnShaderCfgBlock, ctx SpirvBlockContext) {
	if block.IsLoopHeader {
		mergeLabelId := ctx.GetLabelId(block.MergeBlockId)
		continueLabelId := ctx.GetLabelId(block.ContinueBlockId)
		b.EmitLoopMerge(mergeLabelId, continueLabelId, SpvLoopControlNone)
	} else if block.MergeBlockId >= 0 {
		b.EmitSelectionMerge(ctx.GetLabelId(block.MergeBlockId), SpvSelectionControlNone)
	}

	falseLabelId := ctx.GetLabelId(block.Successors[0]) // fall-through.
	trueLabelId := ctx.GetLabelId(block.Successors[1])  // branch target.
	b.EmitBranchConditional(ctx.GcnConditionId, trueLabelId, falseLabelId)
}
