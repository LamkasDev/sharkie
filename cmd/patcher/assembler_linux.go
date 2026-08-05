//go:build linux

package patcher

import (
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/bpfsnoop/gapstone"
)

// MapGapstoneRegister maps a gapstone register to the corresponding sys_struct.SIGNAL_CONTEXT register index.
func MapGapstoneRegister(reg uint) int {
	switch reg {
	case gapstone.X86_REG_RAX:
		return sys_struct.REG_RAX
	case gapstone.X86_REG_RCX:
		return sys_struct.REG_RCX
	case gapstone.X86_REG_RDX:
		return sys_struct.REG_RDX
	case gapstone.X86_REG_RBX:
		return sys_struct.REG_RBX
	case gapstone.X86_REG_RSP:
		return sys_struct.REG_RSP
	case gapstone.X86_REG_RBP:
		return sys_struct.REG_RBP
	case gapstone.X86_REG_RSI:
		return sys_struct.REG_RSI
	case gapstone.X86_REG_RDI:
		return sys_struct.REG_RDI
	case gapstone.X86_REG_R8:
		return sys_struct.REG_R8
	case gapstone.X86_REG_R9:
		return sys_struct.REG_R9
	case gapstone.X86_REG_R10:
		return sys_struct.REG_R10
	case gapstone.X86_REG_R11:
		return sys_struct.REG_R11
	case gapstone.X86_REG_R12:
		return sys_struct.REG_R12
	case gapstone.X86_REG_R13:
		return sys_struct.REG_R13
	case gapstone.X86_REG_R14:
		return sys_struct.REG_R14
	case gapstone.X86_REG_R15:
		return sys_struct.REG_R15
	default:
		return -1
	}
}
