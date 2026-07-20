//go:build linux

package emu

import (
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

// SprintExceptionStackTrace prints exception stack trace from given context.
func SprintExceptionStackTrace(ctx *sys_struct.SIGNAL_CONTEXT) (result string) {
	result = "Stack trace:\n"
	result += SprintAddress(ctx.GetRegister(sys_struct.REG_RIP))
	result += SprintStackTraceFromSP(ctx.GetRegister(sys_struct.REG_RSP))
	return result
}
