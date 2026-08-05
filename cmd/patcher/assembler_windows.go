//go:build windows

package patcher

import (
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/bpfsnoop/gapstone"
)

// SetContextRegister sets a specific register in the sys_struct.CONTEXT structure based on gapstone register.
func SetContextRegister(ctx *sys_struct.CONTEXT, reg uint, value uint64) {
	switch reg {
	case gapstone.X86_REG_RAX:
		ctx.Rax = value
	case gapstone.X86_REG_RCX:
		ctx.Rcx = value
	case gapstone.X86_REG_RDX:
		ctx.Rdx = value
	case gapstone.X86_REG_RBX:
		ctx.Rbx = value
	case gapstone.X86_REG_RSP:
		ctx.Rsp = value
	case gapstone.X86_REG_RBP:
		ctx.Rbp = value
	case gapstone.X86_REG_RSI:
		ctx.Rsi = value
	case gapstone.X86_REG_RDI:
		ctx.Rdi = value
	case gapstone.X86_REG_R8:
		ctx.R8 = value
	case gapstone.X86_REG_R9:
		ctx.R9 = value
	case gapstone.X86_REG_R10:
		ctx.R10 = value
	case gapstone.X86_REG_R11:
		ctx.R11 = value
	case gapstone.X86_REG_R12:
		ctx.R12 = value
	case gapstone.X86_REG_R13:
		ctx.R13 = value
	case gapstone.X86_REG_R14:
		ctx.R14 = value
	case gapstone.X86_REG_R15:
		ctx.R15 = value
	}
}
