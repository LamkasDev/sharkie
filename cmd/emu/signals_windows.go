//go:build windows

package emu

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// ExceptionHandlerGo is the actual exception handler written in Go.
// It's called directly by the assembly trampoline on the restored Go stack.
// We take a uintptr to avoid any potential pointer checks by the runtime during the call.
func ExceptionHandlerGo() uintptr {
	arg := asm.GlobalExceptionInfo
	exceptionInfo := (*sys_struct.EXCEPTION_POINTERS)(unsafe.Pointer(arg))
	code := exceptionInfo.ExceptionRecord.ExceptionCode
	ctx := exceptionInfo.ContextRecord

	switch code {
	case sys_struct.EXCEPTION_ACCESS_VIOLATION:
		if name, ok := elf.FakeAddressMap[ctx.Rip]; ok {
			fmt.Printf(
				"Called external symbol %s at %s...\n",
				color.Blue.Sprint(name),
				color.Yellow.Sprintf("0x%X", ctx.Rip),
			)
			PrintStackTrace(ctx)
			sys_struct.PrintContext(ctx)

			// The return address is on the stack. We need to pop it into RIP.
			// This simulates a RET instruction.
			ctx.Rip = *(*uint64)(unsafe.Pointer(uintptr(ctx.Rsp)))
			ctx.Rsp += 8

			return sys_struct.EXCEPTION_CONTINUE_EXECUTION
		}

		fmt.Printf(
			"Trapped %s at %s...\nAttempted to access address: %s\n",
			color.Red.Sprint("EXCEPTION_ACCESS_VIOLATION"),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
			color.Yellow.Sprintf("0x%X", exceptionInfo.ExceptionRecord.ExceptionInformation[1]),
		)
		sys_struct.PrintContext(ctx)
		PrintStackTrace(ctx)
		StopProfiling()
		os.Exit(1)
	default:
		fmt.Printf(
			"Trapped exception code %s at %s...\n",
			color.Red.Sprint("%d", code),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
		)
		sys_struct.PrintContext(ctx)
		PrintStackTrace(ctx)
		StopProfiling()
		os.Exit(1)
	}

	return sys_struct.EXCEPTION_CONTINUE_SEARCH
}

// SetupSignalHandler registers the assembly trampoline as the Vectored Exception Handler.
func SetupSignalHandler() {
	ret, _, err := sys_struct.AddVectoredExceptionHandler.Call(1, asm.ExceptionHandlerAddr)
	if ret == 0 {
		panic(fmt.Sprintf("Failed to add vectored exception handler: %v", err))
	}

	fmt.Printf(
		"Vectored Exception Handler registered at %s.\n",
		color.Yellow.Sprintf("0x%X", asm.ExceptionHandlerAddr),
	)
}
