package lib

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

const (
	LibSceLibcInternalCxaGuardMutexOffset = uintptr(0x154CF0)
	LibSceLibcInternalCxaGuardCondOffset  = uintptr(0x154CF8)
)

// 0x00000000000D3720
// void __fastcall __noreturn sub_D3720(__int64, __int64, __int64, __int64, __int64, __int64, __m128 _XMM0, __m128 _XMM1, __m128 _XMM2, __m128 _XMM3, __m128 _XMM4, __m128 _XMM5, __m128 _XMM6, __m128 _XMM7, char)
func libSceLibcInternal_printErrAbort(message string) {
	logger.Printf(message)
	logger.CleanupAndExit()
}

// 0x00000000000CD8E0
// void _cxa_guard_release(__guard *)
func libSceLibcInternal___cxa_guard_release(guardPtr uintptr) uintptr {
	module := emu.GlobalModuleManager.ModulesMap["libSceLibcInternal.sprx"]
	return cxaGuardRelease(
		module.BaseAddress+LibSceLibcInternalCxaGuardMutexOffset,
		module.BaseAddress+LibSceLibcInternalCxaGuardCondOffset,
		guardPtr,
	)
}

func cxaGuardRelease(mutexAddr, condAddr, guardPtr uintptr) uintptr {
	if libKernel_pthread_mutex_lock(mutexAddr) != 0 {
		libSceLibcInternal_printErrAbort(
			fmt.Sprintf("%-120s %s failed to acquire mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("__cxa_guard_release"),
			),
		)
		return 0
	}
	*(*byte)(unsafe.Pointer(guardPtr)) = 1
	logger.Printf("%-120s %s marked guard %s as initialized.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__cxa_guard_release"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	if libKernel_pthread_mutex_unlock(mutexAddr) != 0 {
		libSceLibcInternal_printErrAbort(
			fmt.Sprintf("%-120s %s failed to release mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("__cxa_guard_release"),
			),
		)
	}
	if libKernel_pthread_cond_broadcast(condAddr) != 0 {
		libSceLibcInternal_printErrAbort(
			fmt.Sprintf("%-120s %s failed to broadcast condition variable.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("__cxa_guard_release"),
			),
		)
	}

	return 0
}
