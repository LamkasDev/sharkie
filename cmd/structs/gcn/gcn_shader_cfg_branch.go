package gcn

// BranchCond names the GCN condition register that controls a conditional branch.
type BranchCond uint8

const (
	CondNone   BranchCond = iota // unconditional
	CondScc0                     // branch if SCC == 0
	CondScc1                     // branch if SCC == 1
	CondVccZ                     // branch if VCC == 0
	CondVccNz                    // branch if VCC != 0
	CondExecZ                    // branch if EXEC == 0
	CondExecNz                   // branch if EXEC != 0
)

var BranchCondNames = map[BranchCond]string{
	CondNone:   "None",
	CondScc0:   "Scc0",
	CondScc1:   "Scc1",
	CondVccZ:   "VccZ",
	CondVccNz:  "VccNz",
	CondExecZ:  "ExecZ",
	CondExecNz: "ExecNz",
}

var BranchCondMap = map[uint32]BranchCond{
	SoppOpCBranchScc0:   CondScc0,
	SoppOpCBranchScc1:   CondScc1,
	SoppOpCBranchVccZ:   CondVccZ,
	SoppOpCBranchVccNz:  CondVccNz,
	SoppOpCBranchExecZ:  CondExecZ,
	SoppOpCBranchExecNz: CondExecNz,
}

func (c BranchCond) String() string {
	return BranchCondNames[c]
}

func NewBranchCond(op uint32) BranchCond {
	cond, ok := BranchCondMap[op]
	if ok {
		return cond
	}

	return CondNone
}
