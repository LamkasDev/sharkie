//go:build windows

package emu

import (
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

// SprintExceptionStackTrace prints exception stack trace from given context.
func SprintExceptionStackTrace(ctx *sys_struct.CONTEXT) (result string) {
	result = "Stack trace:\n"
	result += SprintAddress(uintptr(ctx.Rip))
	result += SprintStackTraceFromSP(uintptr(ctx.Rsp))
	return result
}
