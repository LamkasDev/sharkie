//go:build windows

package emu

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// ExceptionHandlerGo is the actual exception handler written in Go.
// It's called directly by the assembly trampoline on the restored Go stack.
// We take a uintptr to avoid any potential pointer checks by the runtime during the call.
func ExceptionHandlerGo() uintptr {
	thread := GetCurrentThread()
	threadContext := asm.GetCurrentThreadContext()

	exceptionInfo := (*sys_struct.EXCEPTION_POINTERS)(unsafe.Pointer(threadContext.GlobalExceptionInfo))
	code := exceptionInfo.ExceptionRecord.ExceptionCode
	ctx := exceptionInfo.ContextRecord

	switch code {
	case sys_struct.EXCEPTION_ACCESS_VIOLATION:
		if name, ok := elf.FakeAddressMap[ctx.Rip]; ok {
			result := fmt.Sprintf(
				"[%s] Called external symbol %s at %s...\n",
				color.Green.Sprint(thread.Name),
				color.Blue.Sprint(name),
				color.Yellow.Sprintf("0x%X", ctx.Rip),
			)
			result += SprintException(ctx)
			logger.Print(result)

			// The return address is on the stack. We need to pop it into RIP.
			ctx.Rip = *(*uint64)(unsafe.Pointer(uintptr(ctx.Rsp)))
			ctx.Rsp += 8

			return sys_struct.EXCEPTION_CONTINUE_EXECUTION
		}

		result := fmt.Sprintf(
			"[%s] Trapped %s at %s...\nAttempted to access address: %s\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprint("EXCEPTION_ACCESS_VIOLATION"),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
			color.Yellow.Sprintf("0x%X", exceptionInfo.ExceptionRecord.ExceptionInformation[1]),
		)
		result += SprintException(ctx)
		logger.Print(result)
		logger.CleanupAndExit()
		break
	case sys_struct.EXCEPTION_SINGLE_STEP:
		result := fmt.Sprintf(
			"[%s] Trapped %s at %s...\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprint("EXCEPTION_SINGLE_STEP"),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
		)
		result += SprintException(ctx)
		logger.Print(result)
		ctx.Dr6 = 0

		return sys_struct.EXCEPTION_CONTINUE_EXECUTION
	default:
		result := fmt.Sprintf(
			"[%s] Trapped exception code %s at %s...\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprintf("0x%X", code),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
		)
		result += SprintException(ctx)
		logger.Print(result)
		logger.CleanupAndExit()
		break
	}

	return sys_struct.EXCEPTION_CONTINUE_SEARCH
}

func SprintException(ctx *sys_struct.CONTEXT) (result string) {
	result += sys_struct.SprintContext(ctx)
	result += sys_struct.SprintRegister("TCB", uint64(uintptr(unsafe.Pointer(asm.GetCurrentThreadContext()))))
	result += SprintStackTrace(ctx)

	return result
}

// SetupSignalHandler registers the assembly trampoline as the Vectored Exception Handler.
func SetupSignalHandler() {
	ret, _, err := sys_struct.AddVectoredExceptionHandler.Call(1, asm.ExceptionHandlerAddr)
	if ret == 0 {
		panic(fmt.Sprintf("Failed to add vectored exception handler: %v", err))
	}

	logger.Printf(
		"Vectored Exception Handler registered at %s.\n",
		color.Yellow.Sprintf("0x%X", asm.ExceptionHandlerAddr),
	)
}
