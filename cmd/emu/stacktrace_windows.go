//go:build windows

package emu

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

// PrintStackTrace prints the stack trace from given context.
func PrintStackTrace(ctx *sys_struct.CONTEXT) {
	logger.Println("Stack trace:")
	PrintAddress(uintptr(ctx.Rip))

	stackPtr := uintptr(ctx.Rsp)
	stackTop := GlobalModuleManager.Stack.Address + uintptr(structs.StackDefaultSize)
	for i := 0; i < 40; i++ {
		if stackPtr >= stackTop {
			break
		}
		address, ok := SafeReadUint64(stackPtr)
		if !ok {
			break
		}
		PrintAddress(uintptr(address))

		stackPtr += 8
	}
}
