package kernel

import (
	"os"

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
	nativeSignum := emu.OrbisToPlatformSignal(signum)
	if err := emu.Tgkill(os.Getpid(), int(thread.OsThreadId), nativeSignum); err != nil {
		panic(err)
	}

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
