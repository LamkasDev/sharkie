package lib

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
	"github.com/gookit/color"
)

// TODO: fix name creation

// 0x0000000000013AA0
// __int64 __fastcall scePthreadMutexInit(_QWORD *a1, __int64 a2, __int64 a3)
func libKernel_scePthreadMutexInit(mutexHandlePtr uintptr, attrPtr uintptr, namePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_init(mutexHandlePtr, attrPtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	// Retrieve structs back.
	mutexAddr := *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))

	// Set name.
	var name string
	if namePtr != 0 {
		name = ReadCString(namePtr)
	} else {
		name = fmt.Sprintf("Mutex_%x", mutexAddr)
	}
	mutex.Name = strings.Clone(name)

	// TODO: emulate __sys_namedobj_create?

	logger.Printf("%-132s %s named mutex %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadMutexInit"),
		GetMutexNameText(mutex, mutexAddr),
	)
	return 0
}

// 0x0000000000013C50
// __int64 scePthreadMutexDestroy()
func libKernel_scePthreadMutexDestroy(mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_destroy(mutexHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013C70
// __int64 __fastcall scePthreadMutexLock(__int64 *, __int64, int, int, int, int)
func libKernel_scePthreadMutexLock(mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013C90
// __int64 scePthreadMutexTrylock()
func libKernel_scePthreadMutexTrylock(mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_trylock(mutexHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013CD0
// __int64 __fastcall scePthreadMutexUnlock(__int64 *, __int64, __int64, __int64)
func libKernel_scePthreadMutexUnlock(mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013CB0
// __int64 scePthreadMutexTimedlock()
func libKernel_scePthreadMutexTimedlock(mutexHandlePtr uintptr, micros uintptr) uintptr {
	err := libKernel_pthread_mutex_reltimedlock_np(mutexHandlePtr, micros)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}
