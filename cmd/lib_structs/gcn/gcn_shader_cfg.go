package gcn

import (
	"slices"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
)

// GcnShaderCfg is the Control Flow Graph of a shader.
type GcnShaderCfg struct {
	Blocks         []GcnShaderCfgBlock
	BlocksByOffset map[uintptr]int
}

func NewGcnShaderCfg(instructions []spec.Instruction) (GcnShaderCfg, error) {
	// Find leading block offsets.
	leaders := map[uintptr]bool{instructions[0].DwordOffset: true}
	for i := range instructions {
		instr := &instructions[i]
		if !instr.IsBranchTerminator() {
			continue
		}
		if instr.Details.(*spec.ScalarDetails).Op == spec.SoppOpEndpgm {
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
		if instr.Encoding == spec.EncEXP && instr.Details.(*spec.ExpDetails).Done {
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

	// Inject fake S_ENDPGM for empty blocks (usually targets outside the shader bounds).
	for i := range blocks {
		if len(blocks[i].Instructions) == 0 {
			blocks[i].Instructions = append(blocks[i].Instructions, spec.Instruction{
				Encoding: spec.EncSOPP,
				Details: &spec.ScalarDetails{
					Op: spec.SoppOpEndpgm,
				},
				DwordOffset: blocks[i].DwordOffset,
				DwordLen:    1,
			})
		}
	}

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

	// Analyze created graph for SPIR-V annotations.
	cfg.Analyze()

	return cfg, nil
}

// ClassifyTerminator returns Term, BranchCond and Successors for a block.
func (cfg *GcnShaderCfg) ClassifyTerminator(term *spec.Instruction) (TermKind, BranchCond, []int) {
	// S_ENDPGM. No successors.
	if term.Encoding == spec.EncSOPP && term.Details.(*spec.ScalarDetails).Op == spec.SoppOpEndpgm {
		return TermEndpgm, CondNone, nil
	}

	// EXP with done=true followed by S_ENDPGM. No successors.
	/* if term.Encoding == spec.EncEXP && term.Details.(*spec.ExpDetails).Done {
		return TermExpDone, CondNone, nil
	} */

	// S_BRANCH (unconditional). One successor.
	if term.Encoding == spec.EncSOPP && term.Details.(*spec.ScalarDetails).Op == spec.SoppOpBranch {
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

		return TermCBranch, NewBranchCond(term.Details.(*spec.ScalarDetails).Op), successors
	}

	// Block ends because the next block starts (no explicit branch). One successor.
	nextOffset := term.DwordOffset + uintptr(term.DwordLen)
	if nextId, ok := cfg.BlocksByOffset[nextOffset]; ok {
		return TermFallthrough, CondNone, []int{nextId}
	}

	return TermFallthrough, CondNone, nil
}
