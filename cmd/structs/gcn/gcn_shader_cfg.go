package gcn

import (
	"slices"
)

// GcnShaderCfg is the Control Flow Graph of a shader.
type GcnShaderCfg struct {
	Blocks         []GcnShaderCfgBlock
	BlocksByOffset map[uintptr]int
}

func NewGcnShaderCfg(instructions []Instruction) (GcnShaderCfg, error) {
	// Find leading block offsets.
	leaders := map[uintptr]bool{instructions[0].DwordOffset: true}
	for i := range instructions {
		instr := &instructions[i]
		if !instr.IsBranchTerminator() {
			continue
		}
		if instr.SOp == SoppOpEndpgm {
			continue
		}

		// The instruction immediately after a branch starts a new block.
		nextOffset := instr.DwordOffset + uintptr(instr.DwordLen)
		leaders[nextOffset] = true
		leaders[instr.BranchTargetDwordOffset()] = true
	}

	// Treat EXP(done) as a block boundary (terminates PS before ENDPGM).
	for i := range instructions {
		instr := &instructions[i]
		if instr.Encoding == EncEXP && instr.ExpDone {
			nextOffset := instr.DwordOffset + uintptr(instr.DwordLen)
			leaders[nextOffset] = true
		}
	}

	// Sort leaders so we can assign block IDs in order.
	sortedLeaders := make([]uintptr, 0, len(leaders))
	for offset := range leaders {
		sortedLeaders = append(sortedLeaders, offset)
	}
	slices.Sort(sortedLeaders)

	// Map dword offsets to block IDs.
	leadersToIds := make(map[uintptr]int, len(sortedLeaders))
	for id, offset := range sortedLeaders {
		leadersToIds[offset] = id
	}

	// Split leaders into blocks.
	blocks := make([]GcnShaderCfgBlock, len(sortedLeaders))
	for id, offset := range sortedLeaders {
		blocks[id] = GcnShaderCfgBlock{
			Id:              id,
			DwordOffset:     offset,
			MergeBlockId:    -1,
			ContinueBlockId: -1,
		}
	}

	// Construct blocks by walking through.
	currentBlockId := 0
	for i := range instructions {
		instr := &instructions[i]
		offset := instr.DwordOffset

		// Switch to a new block when we reach a leader.
		if id, isLeader := leadersToIds[offset]; isLeader && id != currentBlockId {
			currentBlockId = id
		}

		blocks[currentBlockId].Instructions = append(blocks[currentBlockId].Instructions, *instr)
	}

	// Remove blocks with no instructions.
	nonEmptyBlocks := blocks[:0]
	for i := range blocks {
		if len(blocks[i].Instructions) > 0 {
			nonEmptyBlocks = append(nonEmptyBlocks, blocks[i])
		}
	}
	blocks = nonEmptyBlocks

	// Re-assign block IDs after filtering.
	blocksByOffset := make(map[uintptr]int, len(blocks))
	for i := range blocks {
		blocks[i].Id = i
		blocksByOffset[blocks[i].DwordOffset] = i
	}

	// Link edges.
	cfg := GcnShaderCfg{Blocks: blocks, BlocksByOffset: blocksByOffset}
	for i := range cfg.Blocks {
		block := &cfg.Blocks[i]
		block.Term, block.BranchCond, block.Successors = cfg.ClassifyTerminator(block.Terminator())
	}

	// Backfill predecessor lists.
	for i := range cfg.Blocks {
		for _, succID := range cfg.Blocks[i].Successors {
			cfg.Blocks[succID].Predecessors = append(cfg.Blocks[succID].Predecessors, i)
		}
	}

	return cfg, nil
}

// ClassifyTerminator returns Term, BranchCond and Successors for a block.
func (cfg *GcnShaderCfg) ClassifyTerminator(term *Instruction) (TermKind, BranchCond, []int) {
	// S_ENDPGM. No successors.
	if term.Encoding == EncSOPP && term.SOp == SoppOpEndpgm {
		return TermEndpgm, CondNone, nil
	}

	// EXP with done=true followed by S_ENDPGM. No successors.
	if term.Encoding == EncEXP && term.ExpDone {
		return TermExpDone, CondNone, nil
	}

	// S_BRANCH (unconditional). One successor.
	if term.Encoding == EncSOPP && term.SOp == SoppOpBranch {
		targetId, ok := cfg.BlocksByOffset[term.BranchTargetDwordOffset()]
		if !ok {
			return TermBranch, CondNone, nil
		}
		return TermBranch, CondNone, []int{targetId}
	}

	// S_CBRANCH_*. Two successors (fall-through & target).
	if term.IsConditionalBranch() {
		var successors []int
		fallthroughOffset := term.DwordOffset + uintptr(term.DwordLen)
		fallthroughId, fallthroughOk := cfg.BlocksByOffset[fallthroughOffset]
		if fallthroughOk {
			successors = append(successors, fallthroughId)
		}
		targetOffset := term.BranchTargetDwordOffset()
		targetId, targetOk := cfg.BlocksByOffset[targetOffset]
		if targetOk {
			successors = append(successors, targetId)
		}

		return TermCBranch, NewBranchCond(term.SOp), successors
	}

	// Block ends because the next block starts (no explicit branch). One successor.
	nextOffset := term.DwordOffset + uintptr(term.DwordLen)
	if nextId, ok := cfg.BlocksByOffset[nextOffset]; ok {
		return TermFallthrough, CondNone, []int{nextId}
	}

	return TermFallthrough, CondNone, nil
}
