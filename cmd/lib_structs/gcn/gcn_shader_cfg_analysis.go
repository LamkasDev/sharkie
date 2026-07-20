package gcn

// Analyze computes dominators, detects loops and annotates each block.
// Annotations are SPIR-V merge and continue targets.
func (cfg *GcnShaderCfg) Analyze() {
	if len(cfg.Blocks) == 0 {
		return
	}

	rpo := cfg.ReversePostOrder()
	idom := cfg.computeDominators(rpo)
	cfg.detectLoops(rpo, idom)
	cfg.computeMergeBlocks(idom)
}

// ReversePostOrder returns block IDs in reverse post-order starting from the entry.
func (cfg *GcnShaderCfg) ReversePostOrder() []int {
	n := len(cfg.Blocks)
	visitedBlockIds := make([]bool, n)
	poBlockIds := make([]int, 0, n)

	// Perform a depth first search to collect block IDs.
	var dfs func(blockId int)
	dfs = func(blockId int) {
		if blockId < 0 || blockId >= n || visitedBlockIds[blockId] {
			return
		}
		visitedBlockIds[blockId] = true
		for _, succ := range cfg.Blocks[blockId].Successors {
			dfs(succ)
		}
		poBlockIds = append(poBlockIds, blockId)
	}
	dfs(0)

	// Reverse to get reverse-post-order.
	rpoBlockIds := make([]int, len(poBlockIds))
	for i, id := range poBlockIds {
		rpoBlockIds[len(poBlockIds)-1-i] = id
	}

	return rpoBlockIds
}

// TODO: split into another function to reduce shared code.
// computeDominators returns immediate dominators for each block.
func (cfg *GcnShaderCfg) computeDominators(rpoBlockIds []int) []int {
	n := len(cfg.Blocks)

	// Create initial slice.
	immDom := make([]int, n)
	for i := range immDom {
		immDom[i] = -1
	}
	immDom[0] = 0

	// rpoIndex[id] = position of block id in the RPO slice.
	rpoIndex := make([]int, n)
	for pos, id := range rpoBlockIds {
		rpoIndex[id] = pos
	}

	// Walk in RPO order, skipping the entry.
	changed := true
	for changed {
		changed = false
		for _, blockId := range rpoBlockIds[1:] {
			// Find the first processed predecessor (any pred with immDom already set).
			newImmDom := -1
			for _, predBlockId := range cfg.Blocks[blockId].Predecessors {
				if immDom[predBlockId] != -1 {
					newImmDom = predBlockId
					break
				}
			}
			if newImmDom == -1 {
				continue // block not yet reachable.
			}

			// Intersect with all other processed predecessors.
			for _, predBlockId := range cfg.Blocks[blockId].Predecessors {
				if predBlockId == newImmDom || immDom[predBlockId] == -1 {
					continue
				}
				newImmDom = intersectDom(immDom, rpoIndex, predBlockId, newImmDom)
			}
			if immDom[blockId] != newImmDom {
				immDom[blockId] = newImmDom
				changed = true
			}
		}
	}

	return immDom
}

// intersectDom walks up the dominator tree from B1 and B2 until it finds their common ancestor.
func intersectDom(immDom, rpoIndex []int, b1, b2 int) int {
	for b1 != b2 {
		for rpoIndex[b1] > rpoIndex[b2] {
			b1 = immDom[b1]
		}
		for rpoIndex[b2] > rpoIndex[b1] {
			b2 = immDom[b2]
		}
	}

	return b1
}

// dominates returns true when block A dominates block B.
func (cfg *GcnShaderCfg) dominates(immDom []int, a, b int) bool {
	for b != -1 {
		if b == a {
			return true
		}
		if b == immDom[b] {
			return b == a // entry dominates itself; avoid infinite loop
		}
		b = immDom[b]
	}

	return false
}

// detectLoops identifies loop headers by finding back-edges (pred => succ) where succ dominates pred.
func (cfg *GcnShaderCfg) detectLoops(rpoBlockIds []int, immDom []int) {
	n := len(cfg.Blocks)
	onStack := make([]bool, n)
	visited := make([]bool, n)

	// Perform a depth first search to identify loop headers.
	var dfs func(blockId int)
	dfs = func(blockId int) {
		if visited[blockId] {
			return
		}
		visited[blockId] = true
		onStack[blockId] = true
		for _, succBlockId := range cfg.Blocks[blockId].Successors {
			if onStack[succBlockId] {
				// Back-edge => succ (succ is a loop header).
				cfg.Blocks[succBlockId].IsLoopHeader = true

				// The continue block (loop latch) is the block that owns the back-edge.
				if cfg.Blocks[succBlockId].ContinueBlockId == -1 {
					cfg.Blocks[succBlockId].ContinueBlockId = blockId
				}
			} else {
				dfs(succBlockId)
			}
		}
		onStack[blockId] = false
	}
	dfs(0)
}

// isReachable returns true if there is a path from start to end in the CFG.
func (cfg *GcnShaderCfg) isReachable(start, end int) bool {
	if start == end {
		return true
	}
	visited := make([]bool, len(cfg.Blocks))
	var dfs func(int) bool
	dfs = func(node int) bool {
		if node == end {
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
		for _, succ := range cfg.Blocks[node].Successors {
			if dfs(succ) {
				return true
			}
		}
		return false
	}
	return dfs(start)
}

// computeMergeBlocks sets MergeBlockId on every block that requires one for SPIR-V structured control flow.
func (cfg *GcnShaderCfg) computeMergeBlocks(immDom []int) {
	// Compute post-dominators by building the reverse GcnShaderCfg and running the same dominator algorithm.
	postImmDom := cfg.computePostDominators()
	rpo := cfg.ReversePostOrder()
	rpoIndex := make([]int, len(cfg.Blocks))
	for pos, id := range rpo {
		rpoIndex[id] = pos
	}

	for blockId := range cfg.Blocks {
		block := &cfg.Blocks[blockId]
		if block.IsLoopHeader {
			// Loop merge block (the successor of the header that is NOT dominated by the header => the loop exit).
			for _, succBlockId := range block.Successors {
				if !cfg.dominates(immDom, block.Id, succBlockId) {
					block.MergeBlockId = succBlockId
					break
				}
			}

			// Fallback (use post-dominator if no clear exit found).
			if block.MergeBlockId == -1 && postImmDom[block.Id] != -1 && postImmDom[block.Id] != block.Id {
				if cfg.dominates(immDom, block.Id, postImmDom[block.Id]) {
					block.MergeBlockId = postImmDom[block.Id]
				}
			}
			continue
		}

		if block.Term == TermCBranch {
			// A valid merge block must be strictly dominated by the header (immediate child in dom tree).
			// And it must be reachable from all divergent paths (both successors) that do not exit early.
			var children []int
			for i, dom := range immDom {
				if dom == block.Id && i != block.Id {
					children = append(children, i)
				}
			}

			bestM := -1
			bestRpo := len(cfg.Blocks)
			for _, child := range children {
				reachableFromAll := true
				for _, succ := range block.Successors {
					if !cfg.isReachable(succ, child) {
						reachableFromAll = false
						break
					}
				}
				if reachableFromAll {
					if rpoIndex[child] < bestRpo {
						bestM = child
						bestRpo = rpoIndex[child]
					}
				}
			}

			// If no structural merge block exists (e.g. all paths return or break to an outer scope),
			// we will later create a dummy unreachable block for it.
			block.MergeBlockId = bestM
		}
	}

	// SPIR-V requires that a merge block can only be a merge block for one header.
	// We iterate through headers and inject dummy merge blocks for any shared merge blocks,
	// or for headers that didn't find any structural merge block (m == -1).
	mergeToHeader := make(map[int]int)
	for _, blockId := range cfg.ReversePostOrder() {
		block := &cfg.Blocks[blockId]
		m := block.MergeBlockId

		if m == -1 {
			if block.Term != TermCBranch && !block.IsLoopHeader {
				continue
			}
		} else {
			if _, exists := mergeToHeader[m]; !exists {
				mergeToHeader[m] = blockId
				continue
			}
		}

		// Find all escaping edges from this construct.
		// An escaping edge is an edge from a block in the construct to a block NOT in the construct.
		// The construct is the set of all blocks strictly dominated by blockId, plus blockId itself.
		var escapingTargets []int
		for i := range cfg.Blocks {
			if i != blockId && !cfg.dominates(immDom, blockId, i) {
				continue // i is not in the construct
			}
			for _, succ := range cfg.Blocks[i].Successors {
				if succ != blockId && !cfg.dominates(immDom, blockId, succ) {
					// succ is outside the construct, add to escaping targets if not present.
					found := false
					for _, t := range escapingTargets {
						if t == succ {
							found = true
							break
						}
					}
					if !found {
						escapingTargets = append(escapingTargets, succ)
					}
				}
			}
		}

		// This merge block is either shared or non-existent, we must create a unique dummy merge block.
		newBlockId := len(cfg.Blocks)
		dummyBlock := GcnShaderCfgBlock{
			Id:              newBlockId,
			Term:            TermBranch,
			Successors:      nil,
			MergeBlockId:    -1,
			ContinueBlockId: -1,
		}

		if len(escapingTargets) > 0 {
			dummyBlock.Successors = []int{escapingTargets[0]}
		} else if m != -1 {
			dummyBlock.Successors = []int{m}
		}
		cfg.Blocks = append(cfg.Blocks, dummyBlock)

		// Redirect any branches from within this construct that targeted `escapingTargets[0]` or `m` to target `newBlockId` instead.
		for i := range cfg.Blocks {
			// Skip newly added dummy blocks.
			if i >= newBlockId {
				continue
			}
			if i != blockId && !cfg.dominates(immDom, blockId, i) {
				continue
			}

			b := &cfg.Blocks[i]
			for j, succ := range b.Successors {
				if (len(escapingTargets) > 0 && succ == escapingTargets[0]) || (m != -1 && succ == m) {
					b.Successors[j] = newBlockId
				}
			}
		}

		// Update the header's merge block.
		cfg.Blocks[blockId].MergeBlockId = newBlockId
	}
}

type PostGcnShaderCfg struct {
	NumNodes     int
	Entry        int
	Predecessors func(id int) []int
	Successors   func(id int) []int
}

// TODO: split into another function to reduce shared code.
// computePostDominators builds the post-dominator tree.
// Post-dominators are dominators in the reversed GcnShaderCfg with all exit nodes
// (TermEndpgm, TermExpDone, blocks with no successors) treated as a single virtual entry.
func (cfg *GcnShaderCfg) computePostDominators() []int {
	n := len(cfg.Blocks)
	postImmDom := make([]int, n)
	for i := range postImmDom {
		postImmDom[i] = -1
	}

	// Find all exit blocks (no successors or explicit ENDPGM/EXPDONE).
	var exits []int
	for blockId := range cfg.Blocks {
		block := &cfg.Blocks[blockId]
		if len(block.Successors) == 0 || block.Term == TermEndpgm || block.Term == TermExpDone {
			exits = append(exits, block.Id)
		}
	}
	if len(exits) == 0 {
		return postImmDom
	}

	// Build reverse-GcnShaderCfg adjacency (predecessors in original = successors in reverse).
	revSuccessors := make([][]int, n)
	for i := range cfg.Blocks {
		for _, s := range cfg.Blocks[i].Successors {
			revSuccessors[s] = append(revSuccessors[s], i)
		}
	}

	// When multiple exits exist, introduce a virtual exit node (ID = n).
	// When single exit, use it directly.
	var postGcnShaderCfg PostGcnShaderCfg
	if len(exits) == 1 {
		postGcnShaderCfg = PostGcnShaderCfg{
			Successors:   func(id int) []int { return revSuccessors[id] },
			Predecessors: func(id int) []int { return cfg.Blocks[id].Successors },
			NumNodes:     n,
			Entry:        exits[0],
		}
	} else {
		// Virtual exit n (edges from each exit to virtual and reverse).
		revSuccsWithVirtual := make([][]int, n+1)
		for i := range n {
			revSuccsWithVirtual[i] = revSuccessors[i]
		}
		predsWithVirtual := make([][]int, n+1)
		for i := range n {
			predsWithVirtual[i] = cfg.Blocks[i].Successors
		}
		for _, exitBlockId := range exits {
			revSuccsWithVirtual[n] = append(revSuccsWithVirtual[n], exitBlockId)
			predsWithVirtual[exitBlockId] = append(predsWithVirtual[exitBlockId], n)
		}
		predsWithVirtual[n] = []int{} // virtual exit has no predecessors in reverse.

		postGcnShaderCfg = PostGcnShaderCfg{
			Successors:   func(id int) []int { return revSuccsWithVirtual[id] },
			Predecessors: func(id int) []int { return predsWithVirtual[id] },
			NumNodes:     n + 1,
			Entry:        n,
		}
		postImmDom = append(postImmDom, -1) // slot for virtual node.
	}

	// Perform a depth first search to collect block IDs.
	visitedBlockIds := make([]bool, postGcnShaderCfg.NumNodes)
	rpoBlockIds := make([]int, 0, postGcnShaderCfg.NumNodes)
	var dfs2 func(blockId int)
	dfs2 = func(blockId int) {
		if blockId < 0 || blockId >= postGcnShaderCfg.NumNodes || visitedBlockIds[blockId] {
			return
		}
		visitedBlockIds[blockId] = true
		for _, s := range postGcnShaderCfg.Successors(blockId) {
			dfs2(s)
		}
		rpoBlockIds = append(rpoBlockIds, blockId)
	}
	dfs2(postGcnShaderCfg.Entry)

	// Reverse to get reverse-post-order.
	for l, r := 0, len(rpoBlockIds)-1; l < r; l, r = l+1, r-1 {
		rpoBlockIds[l], rpoBlockIds[r] = rpoBlockIds[r], rpoBlockIds[l]
	}

	// rpoIndex[id] = position of block id in the RPO slice.
	rpoIndex := make([]int, postGcnShaderCfg.NumNodes)
	for pos, id := range rpoBlockIds {
		rpoIndex[id] = pos
	}

	// Walk in RPO order, skipping the entry.
	postImmDom[postGcnShaderCfg.Entry] = postGcnShaderCfg.Entry
	changed := true
	for changed {
		changed = false
		for _, blockId := range rpoBlockIds[1:] {
			if blockId >= len(postImmDom) {
				postImmDom = append(postImmDom, -1)
			}

			// Find the first processed predecessor (any pred with immDom already set).
			newImmDom := -1
			for _, predBlockId := range postGcnShaderCfg.Predecessors(blockId) {
				if predBlockId < len(postImmDom) && postImmDom[predBlockId] != -1 {
					newImmDom = predBlockId
					break
				}
			}
			if newImmDom == -1 {
				continue // block not yet reachable.
			}

			// Intersect with all other processed predecessors.
			for _, predBlockId := range postGcnShaderCfg.Predecessors(blockId) {
				if predBlockId >= len(postImmDom) || predBlockId == newImmDom || postImmDom[predBlockId] == -1 {
					continue
				}
				newImmDom = intersectDom(postImmDom, rpoIndex, predBlockId, newImmDom)
			}
			if blockId < len(postImmDom) && postImmDom[blockId] != newImmDom {
				postImmDom[blockId] = newImmDom
				changed = true
			}
		}
	}

	// Strip the virtual-exit entry if we added one.
	// Map virtual post-dominator back to -1 for real nodes that post-dom = virtual.
	if len(exits) > 1 {
		virtual := n
		for i := range n {
			if i < len(postImmDom) && postImmDom[i] == virtual {
				postImmDom[i] = -1 // post-dominated only by virtual exit
			}
		}
		postImmDom = postImmDom[:n]
	}

	return postImmDom
}
