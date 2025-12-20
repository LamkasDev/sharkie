//go:build windows

package emu

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

// PrintStackTrace prints the stack trace from given context.
func PrintStackTrace(ctx *sys_struct.CONTEXT) {
	fmt.Println("Stack trace:")
	PrintAddress(uintptr(ctx.Rip))

	stackPtr := uintptr(ctx.Rsp)
	stackTop := GlobalModuleManager.Stack.Address + uintptr(structs.StackDefaultSize)
	for i := 0; i < 20; i++ {
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
