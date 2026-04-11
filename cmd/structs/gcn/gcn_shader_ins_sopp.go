package gcn

const (
	SoppOpNop                   = 0x0
	SoppOpEndpgm                = 0x1
	SoppOpBranch                = 0x2
	SoppOpCBranchScc0           = 0x4
	SoppOpCBranchScc1           = 0x5
	SoppOpCBranchVccz           = 0x6
	SoppOpCBranchVccnz          = 0x7
	SoppOpCBranchExecz          = 0x8
	SoppOpCBranchExecnz         = 0x9
	SoppOpBarrier               = 0xA
	SoppOpSetkill               = 0xB
	SoppOpWaitcnt               = 0xC
	SoppOpSethalt               = 0xD
	SoppOpSleep                 = 0xE
	SoppOpSetprio               = 0xF
	SoppOpSendmsg               = 0x10
	SoppOpSendmsghalt           = 0x11
	SoppOpTrap                  = 0x12
	SoppOpIcacheInv             = 0x13
	SoppOpIncperflevel          = 0x14
	SoppOpDecperflevel          = 0x15
	SoppOpTtraceData            = 0x16
	SoppOpCbranchCdbgsys        = 0x17
	SoppOpCbranchCdbguser       = 0x18
	SoppOpCbranchCdbgsysOrUser  = 0x19
	SoppOpCbranchCdbgsysAndUser = 0x1A
)

func (instr *Instruction) DecodeSOPP() {
	dw := instr.Dwords[0]
	instr.Details = &ScalarDetails{
		Imm16: int16(dw & 0b1111_1111_1111_1111),
		Op:    (dw >> 16) & 0b1111_111,
	}
}

// IsBranchTerminator returns true when a SOPP instruction terminates a block.
func (instr *Instruction) IsBranchTerminator() bool {
	if instr.Encoding != EncSOPP {
		return false
	}
	switch instr.Details.(*ScalarDetails).Op {
	case SoppOpEndpgm, SoppOpBranch,
		SoppOpCBranchScc0, SoppOpCBranchScc1,
		SoppOpCBranchVccz, SoppOpCBranchVccnz,
		SoppOpCBranchExecz, SoppOpCBranchExecnz:
		return true
	}

	return false
}

// IsConditionalBranch returns true for S_CBRANCH_* instructions.
func (instr *Instruction) IsConditionalBranch() bool {
	if instr.Encoding != EncSOPP {
		return false
	}
	switch instr.Details.(*ScalarDetails).Op {
	case SoppOpCBranchScc0, SoppOpCBranchScc1,
		SoppOpCBranchVccz, SoppOpCBranchVccnz,
		SoppOpCBranchExecz, SoppOpCBranchExecnz:
		return true
	}

	return false
}

// BranchTargetDwordOff returns dword offset of the branch target.
func (instr *Instruction) BranchTargetDwordOffset() uintptr {
	return uintptr(int(instr.DwordOffset+uintptr(instr.DwordLen)) + int(instr.Details.(*ScalarDetails).Imm16))
}
