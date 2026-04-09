package gcn

const (
	SoppOpNop                   = 0x0
	SoppOpEndpgm                = 0x1
	SoppOpBranch                = 0x2
	SoppOpCBranchScc0           = 0x4
	SoppOpCBranchScc1           = 0x5
	SoppOpCBranchVccZ           = 0x6
	SoppOpCBranchVccNz          = 0x7
	SoppOpCBranchExecZ          = 0x8
	SoppOpCBranchExecNz         = 0x9
	SoppOpBarrier               = 0xA
	SoppOpSetKill               = 0xB
	SoppOpWaitCnt               = 0xC
	SoppOpSetHalt               = 0xD
	SoppOpSleep                 = 0xE
	SoppOpSetPrio               = 0xF
	SoppOpSendMsg               = 0x10
	SoppOpSendMsgHalt           = 0x11
	SoppOpTrap                  = 0x12
	SoppOpIcacheInv             = 0x13
	SoppOpIncPerfLevel          = 0x14
	SoppOpDecPerfLevel          = 0x15
	SoppOpTTraceData            = 0x16
	SoppOpCbranchCdbgSys        = 0x17
	SoppOpCbranchCdbgUser       = 0x18
	SoppOpCbranchCdbgSysOrUser  = 0x19
	SoppOpCbranchCdbgSysAndUser = 0x1A
)

func (instr *Instruction) DecodeSOP2() {
	dw := instr.Dwords[0]
	instr.SSrc0 = dw & 0b1111_1111
	instr.SSrc1 = (dw >> 8) & 0b1111_1111
	instr.SDst = (dw >> 16) & 0b1111_111
	instr.SOp = (dw >> 23) & 0b1111_111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeSOP1() {
	dw := instr.Dwords[0]
	instr.SSrc0 = dw & 0b1111_1111
	instr.SOp = (dw >> 8) & 0b1111_1111
	instr.SDst = (dw >> 16) & 0b1111_111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeSOPC() {
	dw := instr.Dwords[0]
	instr.SSrc0 = dw & 0b1111_1111
	instr.SSrc1 = (dw >> 8) & 0b1111_1111
	instr.SOp = (dw >> 16) & 0b1111_111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeSOPP() {
	dw := instr.Dwords[0]
	instr.Simm16 = int16(dw & 0b1111_1111_1111_1111)
	instr.SOp = (dw >> 16) & 0b1111_111
}

// IsBranchTerminator returns true when a SOPP instruction terminates a block.
func (instr *Instruction) IsBranchTerminator() bool {
	if instr.Encoding != EncSOPP {
		return false
	}
	switch instr.SOp {
	case SoppOpEndpgm, SoppOpBranch,
		SoppOpCBranchScc0, SoppOpCBranchScc1,
		SoppOpCBranchVccZ, SoppOpCBranchVccNz,
		SoppOpCBranchExecZ, SoppOpCBranchExecNz:
		return true
	}

	return false
}

// IsConditionalBranch returns true for S_CBRANCH_* instructions.
func (instr *Instruction) IsConditionalBranch() bool {
	if instr.Encoding != EncSOPP {
		return false
	}
	switch instr.SOp {
	case SoppOpCBranchScc0, SoppOpCBranchScc1,
		SoppOpCBranchVccZ, SoppOpCBranchVccNz,
		SoppOpCBranchExecZ, SoppOpCBranchExecNz:
		return true
	}

	return false
}

// BranchTargetDwordOff returns dword offset of the branch target.
func (instr *Instruction) BranchTargetDwordOffset() uintptr {
	return uintptr(int(instr.DwordOffset+uintptr(instr.DwordLen)) + int(instr.Simm16))
}
