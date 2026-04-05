package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
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
	return libKernel_pthread_self()
}

// 0x0000000000013920
// __int64 scePthreadEqual()
func libKernel_scePthreadEqual(t1, t2 uintptr) uintptr {
	err := libKernel_pthread_equal(t1, t2)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000138E0
// __int64 scePthreadCreate()
func libKernel_scePthreadCreate(threadPtr, attrHandlePtr, entryPoint, arg uintptr, namePtr Cstring) uintptr {
	err := libKernel_pthread_create_name_np(threadPtr, attrHandlePtr, entryPoint, arg, namePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013940
// void __fastcall __noreturn scePthreadExit(__int64)
func libKernel_scePthreadExit(retValue uintptr) uintptr {
	return libKernel_pthread_exit(retValue)
}
