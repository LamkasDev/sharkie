//go:build linux

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

/*
	#include <signal.h>
	#include <ucontext.h>
	#include <string.h>
	#include <stdlib.h>

	static int setup_signal_stack() {
		stack_t ss;
		ss.ss_sp = malloc(SIGSTKSZ);
		if (ss.ss_sp == NULL) {
			return -1;
		}
		ss.ss_size = SIGSTKSZ;
		ss.ss_flags = 0;
		if (sigaltstack(&ss, NULL) == -1) {
			free(ss.ss_sp);
			return -1;
		}

		return 0;
	}

	static int add_signal_handlers(void* handler_address) {
		struct sigaction sa;
		memset(&sa, 0, sizeof(sa));

		sa.sa_sigaction = (void (*)(int sig, siginfo_t* info, void* ctx))handler_address;
		sa.sa_flags = SA_SIGINFO | SA_ONSTACK | SA_NODEFER;
		sigemptyset(&sa.sa_mask);

		if (sigaction(SIGSEGV, &sa, NULL) != 0) return -1;
		if (sigaction(SIGBUS, &sa, NULL) != 0) return -2;
		if (sigaction(SIGILL, &sa, NULL) != 0) return -3;
		if (sigaction(SIGTRAP, &sa, NULL) != 0) return -4;
		if (sigaction(SIGFPE, &sa, NULL) != 0) return -5;
		if (sigaction(SIGABRT, &sa, NULL) != 0) return -6;
		if (sigaction(SIGSYS, &sa, NULL) != 0) return -7;

		return 0;
	}
*/
import "C"

// ExceptionHandlerGo is the actual exception handler written in Go.
// It's called directly by the assembly trampoline on the restored Go stack.
// We take a uintptr to avoid any potential pointer checks by the runtime during the call.
func ExceptionHandlerGo() uintptr {
	thread := GetCurrentThread()
	threadContext := asm.GetCurrentThreadContext()

	signalContext := (*sys_struct.SIGNAL_CONTEXT)(unsafe.Pointer(threadContext.GlobalExceptionInfo))
	code := signalContext.GetCode()
	rip := signalContext.GetRegister(sys_struct.REG_RIP)
	rsp := signalContext.GetRegister(sys_struct.REG_RSP)

	switch code {
	case sys_struct.SIGNAL_SIGSEGV, sys_struct.SIGNAL_SIGBUS:
		if name, ok := elf.FakeAddressMap[rip]; ok {
			result := fmt.Sprintf(
				"[%s] Called external symbol %s at %s...\n",
				color.Green.Sprint(thread.Name),
				color.Blue.Sprint(name),
				color.Yellow.Sprintf("0x%X", rip),
			)
			result += SprintException(signalContext)
			logger.Print(result)

			// The return address is on the stack. We need to pop it into RIP.
			signalContext.SetRegister(sys_struct.REG_RIP, *(*uintptr)(unsafe.Pointer(rsp)))
			signalContext.SetRegister(sys_struct.REG_RSP, rsp+8)

			return sys_struct.EXCEPTION_CONTINUE_EXECUTION
		}

		result := fmt.Sprintf(
			"[%s] Trapped %s at %s (%s)...\nAttempted to access address: %s\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprint("EXCEPTION_ACCESS_VIOLATION"),
			color.Yellow.Sprintf("0x%X", rip),
			GlobalModuleManager.GetCallSiteTextShort(),
			color.Yellow.Sprintf("0x%X", signalContext.GetFaultAddress()),
		)
		result += SprintException(signalContext)
		logger.Print(result)
		logger.CleanupAndExit()
		break
	case sys_struct.SIGNAL_SIGTRAP:
		result := fmt.Sprintf(
			"[%s] Trapped %s at %s (%s)...\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprint("EXCEPTION_SINGLE_STEP"),
			color.Yellow.Sprintf("0x%X", rip),
			GlobalModuleManager.GetCallSiteTextShort(),
		)
		result += SprintException(signalContext)
		logger.Print(result)
		// TODO: clear debug register?

		return sys_struct.EXCEPTION_CONTINUE_EXECUTION
	default:
		result := fmt.Sprintf(
			"[%s] Trapped exception %s at %s (%s)...\n",
			color.Green.Sprint(thread.Name),
			color.Red.Sprintf("%s (0x%X)", signalContext.GetName(), signalContext.GetCode()),
			color.Yellow.Sprintf("0x%X", rip),
			GlobalModuleManager.GetCallSiteTextShort(),
		)
		result += SprintException(signalContext)
		logger.Print(result)
		logger.CleanupAndExit()
		break
	}

	return sys_struct.EXCEPTION_CONTINUE_SEARCH
}

func SprintException(ctx *sys_struct.SIGNAL_CONTEXT) (result string) {
	result += sys_struct.SprintContext(ctx)
	threadContext := asm.GetCurrentThreadContext()
	result += sys_struct.SprintRegister("TCB", uint64(uintptr(unsafe.Pointer(threadContext))))
	result += SprintStackTrace(ctx)

	return result
}

// SetupSignalHandler registers the assembly trampoline for specified signals.
func SetupSignalHandler() {
	if C.setup_signal_stack() != 0 {
		panic("failed to setup signal stack")
	}
	ret := C.add_signal_handlers(unsafe.Pointer(asm.ExceptionHandlerAddr))
	if ret != 0 {
		panic(fmt.Sprintf("failed to add signal handlers: %v", ret))
	}

	logger.Printf("Signal Handlers registered.\n")
}
