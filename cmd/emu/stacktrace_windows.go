//go:build windows

package emu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

// SprintStackTrace prints the stack trace from given context.
func SprintStackTrace(ctx *sys_struct.CONTEXT) (result string) {
	thread := GetCurrentThread()
	result = "Stack trace:\n"
	result += SprintAddress(uintptr(ctx.Rip))

	stackPtr := uintptr(ctx.Rsp)
	if ctx.Rsp <= 0x1000 {
		return result
	}
	stackTop := thread.Stack.Address + structs.StackDefaultSize
	for i := 0; i < 40; i++ {
		if stackPtr >= stackTop {
			break
		}
		address := *(*uint64)(unsafe.Pointer(stackPtr))
		result += SprintAddress(uintptr(address))

		stackPtr += 8
	}

	return result
}
