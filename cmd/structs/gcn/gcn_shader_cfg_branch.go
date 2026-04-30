package gcn

import "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"

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
	spec.SoppOpCbranchScc0:   CondScc0,
	spec.SoppOpCbranchScc1:   CondScc1,
	spec.SoppOpCbranchVccz:   CondVccz,
	spec.SoppOpCbranchVccnz:  CondVccnz,
	spec.SoppOpCbranchExecz:  CondExecz,
	spec.SoppOpCbranchExecnz: CondExecnz,
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
