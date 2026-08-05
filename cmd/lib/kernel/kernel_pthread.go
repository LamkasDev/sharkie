package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000146E0
// __int64 scePthreadGetthreadid()
func libKernel_scePthreadGetthreadid() uintptr {
	thread := emu.GetCurrentThread()

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePthreadGetthreadid"),
			color.Yellow.Sprintf("%d", thread.Id),
		)
	}
	return uintptr(thread.Id)
}

// 0x00000000000146E0
// __int64 scePthreadSelf()
func libKernel_scePthreadSelf() uintptr {
	return posix.Pthread_self()
}

// 0x0000000000013920
// __int64 scePthreadEqual()
func libKernel_scePthreadEqual(t1, t2 uintptr) uintptr {
	err := posix.Pthread_equal(t1, t2)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000138E0
// __int64 scePthreadCreate()
func libKernel_scePthreadCreate(threadPtr uintptr, attrHandlePtr *uintptr, entryPoint, arg uintptr, namePtr Cstring) uintptr {
	err := posix.Pthread_create_name_np(threadPtr, attrHandlePtr, entryPoint, arg, namePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013940
// void __fastcall __noreturn scePthreadExit(__int64)
func libKernel_scePthreadExit(retValue uintptr) uintptr {
	err := posix.Pthread_exit(retValue)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013DD0
// __int64 scePthreadRwlockRdlock()
func libKernel_scePthreadRwlockRdlock() uintptr {
	err := posix.Pthread_rwlock_rdlock()
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013E90
// __int64 scePthreadRwlockWrlock()
func libKernel_scePthreadRwlockWrlock() uintptr {
	err := posix.Pthread_rwlock_wrlock()
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013E70
// __int64 scePthreadRwlockUnlock()
func libKernel_scePthreadRwlockUnlock() uintptr {
	err := posix.Pthread_rwlock_unlock()
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013980
// __int64 __fastcall scePthreadJoin(__int64, __int64)
func libKernel_scePthreadJoin(threadPtr, retValPtr uintptr) uintptr {
	err := posix.Pthread_join(threadPtr, retValPtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013CF0
// __int64 __fastcall scePthreadOnce(volatile signed __int32 *, void (*)(void), double)
func libKernel_scePthreadOnce(onceControl *PthreadOnce, initRoutine uintptr) uintptr {
	err := posix.Pthread_once(onceControl, initRoutine)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014250
// __int64 scePthreadYield()
func libKernel_scePthreadYield() uintptr {
	return posix.Pthread_yield()
}
