//go:build windows

package emu

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/kernel"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/patcher"
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

	// TODO: finish this.
	if len(kernel.ExceptionHandlers) > 0 {
		panic("not implemented")
	}

	switch code {
	case sys_struct.EXCEPTION_ACCESS_VIOLATION:
		if patcher.GlobalPatcherRuntime.IsTcbAccess(uint64(ctx.Rip)) {
			if patcher.EmulateTcbAccess(ctx, uint64(ctx.Rip)) {
				return sys_struct.EXCEPTION_CONTINUE_EXECUTION
			}
		}

		result := fmt.Sprintf(
			"[%s] Trapped %s at %s (%s)...\nAttempted to access address: %s\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprint("EXCEPTION_ACCESS_VIOLATION"),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
			GlobalModuleManager.GetCallSiteTextShort(),
			color.Yellow.Sprintf("0x%X", exceptionInfo.ExceptionRecord.ExceptionInformation[1]),
		)
		result += SprintException(ctx)
		logger.Print(result)
		logger.CleanupAndExit()
		break
	case sys_struct.EXCEPTION_SINGLE_STEP:
		result := fmt.Sprintf(
			"[%s] Trapped %s at %s (%s)...\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprint("EXCEPTION_SINGLE_STEP"),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
			GlobalModuleManager.GetCallSiteTextShort(),
		)
		result += SprintException(ctx)
		logger.Print(result)
		ctx.Dr6 = 0

		return sys_struct.EXCEPTION_CONTINUE_EXECUTION
	default:
		result := fmt.Sprintf(
			"[%s] Trapped exception code %s at %s (%s)...\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprintf("0x%X", code),
			color.Yellow.Sprintf("0x%X", ctx.Rip),
			GlobalModuleManager.GetCallSiteTextShort(),
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
	result += SprintExceptionStackTrace(ctx)

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
