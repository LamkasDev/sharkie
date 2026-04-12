package gcn

// BranchCond names the GCN condition register that controls a conditional branch.
type BranchCond uint8

const (
	CondNone   BranchCond = iota // unconditional
	CondScc0                     // branch if SCC == 0
	CondScc1                     // branch if SCC == 1
	CondVccz                     // branch if VCC == 0
	CondVccnz                    // branch if VCC != 0
	CondExecz                    // branch if EXEC == 0
	CondExecnz                   // branch if EXEC != 0
)

var BranchCondNames = map[BranchCond]string{
	CondNone:   "None",
	CondScc0:   "Scc0",
	CondScc1:   "Scc1",
	CondVccz:   "Vccz",
	CondVccnz:  "Vccnz",
	CondExecz:  "Execz",
	CondExecnz: "Execnz",
}

var BranchCondMap = map[uint32]BranchCond{
	SoppOpCbranchScc0:   CondScc0,
	SoppOpCbranchScc1:   CondScc1,
	SoppOpCbranchVccz:   CondVccz,
	SoppOpCbranchVccnz:  CondVccnz,
	SoppOpCbranchExecz:  CondExecz,
	SoppOpCbranchExecnz: CondExecnz,
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
