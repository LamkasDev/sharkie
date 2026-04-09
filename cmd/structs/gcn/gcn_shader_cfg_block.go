package gcn

// TermKind classifies how a block exits.
type TermKind uint8

const (
	// S_ENDPGM. No successors.
	TermEndpgm TermKind = iota

	// EXP with done=true followed by S_ENDPGM. No successors.
	TermExpDone

	// S_BRANCH (unconditional). One successor.
	TermBranch

	// S_CBRANCH_*. Two successors (fall-through & target).
	TermCBranch

	// Block ends because the next block starts (no explicit branch). One successor.
	TermFallthrough
)

var TermKindNames = map[TermKind]string{
	TermEndpgm:      "Engpgm",
	TermExpDone:     "ExpDone",
	TermBranch:      "Branch",
	TermCBranch:     "Cbranch",
	TermFallthrough: "Fallthrough",
}

func (t TermKind) String() string {
	return TermKindNames[t]
}

// GcnShaderCfgBlock is a sequence of instructions with one entry and exit point.
type GcnShaderCfgBlock struct {
	Id           int
	DwordOffset  uintptr
	Instructions []Instruction

	Term         TermKind
	BranchCond   BranchCond
	Successors   []int
	Predecessors []int

	IsLoopHeader    bool
	MergeBlockId    int
	ContinueBlockId int
}

func (b *GcnShaderCfgBlock) Terminator() *Instruction {
	return &b.Instructions[len(b.Instructions)-1]
}
