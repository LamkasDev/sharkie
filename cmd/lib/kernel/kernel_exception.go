package kernel

import (
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/kernel"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000211D0
// __int64 __fastcall sceKernelInstallExceptionHandler(unsigned int a1, __int64 a2)
func libKernel_sceKernelInstallExceptionHandler(signum int, handler uintptr) uintptr {
	kernel.ExceptionHandlers[signum] = handler

	logger.Printf("%-132s %s added exception handler for signal %s (handler=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelInstallExceptionHandler"),
		color.Green.Sprint(signum),
		color.Yellow.Sprintf("0x%X", handler),
	)
	return 0
}

// 0x00000000000212F0
// __int64 __fastcall sceKernelRemoveExceptionHandler(unsigned int a1)
func libKernel_sceKernelRemoveExceptionHandler(signum int) uintptr {
	delete(kernel.ExceptionHandlers, signum)

	logger.Printf("%-132s %s removed exception handler for signal %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelRemoveExceptionHandler"),
		color.Green.Sprint(signum),
	)
	return 0
}

// 0x0000000000021430
// __int64 __fastcall sceKernelRaiseException(__int64, int)
func libKernel_sceKernelRaiseException(threadPtr uintptr, signum int) uintptr {
	thread := emu.GetThreadForPtr(threadPtr)
	if thread == nil {
		logger.Printf("%-132s %s failed due to invalid thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelRaiseException"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return 0x80020003
	}
	currentThread := emu.GetCurrentThread()
	if thread.Id == currentThread.Id {
		return 0x80020003
	}

	go func() {
		// Ensure thread is suspended in Go code.
		thread.Lock.Lock()
		thread.SuspendedByGC = true
		thread.Lock.Unlock()
		for thread.InGuest.Load() {
			runtime.Gosched()
		}

		// Populate context.
		uctx := &kernel.Ucontext{}
		threadContext := asm.ThreadContextRepo[uintptr(unsafe.Pointer(thread))]
		if threadContext.GlobalStubContext != 0 {
			stubCtx := (*asm.RegContext)(unsafe.Pointer(threadContext.GlobalStubContext))
			uctx.Mcontext.Rax = uint64(stubCtx.AX)
			uctx.Mcontext.Rcx = uint64(stubCtx.CX)
			uctx.Mcontext.Rdx = uint64(stubCtx.DX)
			uctx.Mcontext.Rbx = uint64(stubCtx.BX)
			uctx.Mcontext.Rbp = uint64(stubCtx.BP)
			uctx.Mcontext.Rsi = uint64(stubCtx.SI)
			uctx.Mcontext.Rdi = uint64(stubCtx.DI)
			uctx.Mcontext.R8 = uint64(stubCtx.R8)
			uctx.Mcontext.R9 = uint64(stubCtx.R9)
			uctx.Mcontext.R10 = uint64(stubCtx.R10)
			uctx.Mcontext.R11 = uint64(stubCtx.R11)
			uctx.Mcontext.R12 = uint64(stubCtx.R12)
			uctx.Mcontext.R13 = uint64(stubCtx.R13)
			uctx.Mcontext.R14 = uint64(stubCtx.R14)
			uctx.Mcontext.R15 = uint64(stubCtx.R15)
			uctx.Mcontext.Rip = *(*uint64)(unsafe.Pointer(threadContext.GlobalStubContext + 384))
			uctx.Mcontext.Rsp = uint64(threadContext.GlobalStubContext + 392)
		} else {
			// Thread might not have executed any stubs yet. Provide initial stack pointer.
			uctx.Mcontext.Rsp = uint64(thread.Stack.CurrentPointer)
			uctx.Mcontext.Rip = 0
		}

		// Dispatch exception handler.
		if handlerAddr, ok := kernel.ExceptionHandlers[signum]; ok {
			thread.CallException(handlerAddr, uintptr(signum), uintptr(unsafe.Pointer(uctx)))
		}

		// Unlock target thread.
		thread.Lock.Lock()
		thread.SuspendedByGC = false
		thread.SuspendCond.Broadcast()
		thread.Lock.Unlock()
	}()

	logger.Printf("%-132s %s raised signal %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelRaiseException"),
		color.Green.Sprint(signum),
	)
	return 0
}

// 0x0000000000022D40
// __int64 __fastcall sceKernelDebugRaiseException(__int64, __int64)
func libKernel_sceKernelDebugRaiseException(err, argsPtr uintptr) uintptr {
	logger.Printf("%-132s %s called with %s, exiting...\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelDebugRaiseException"),
		color.Red.Sprintf("0x%X", err),
	)
	logger.CleanupAndExit()

	return 0
}

// 0x0000000000022D60
// __int64 __fastcall sceKernelDebugRaiseExceptionOnReleaseMode(int, __int64)
func libKernel_sceKernelDebugRaiseExceptionOnReleaseMode(err, argsPtr uintptr) uintptr {
	logger.Printf("%-132s %s called with %s, exiting...\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelDebugRaiseExceptionOnReleaseMode"),
		color.Red.Sprintf("0x%X", err),
	)
	logger.CleanupAndExit()

	return 0
}
