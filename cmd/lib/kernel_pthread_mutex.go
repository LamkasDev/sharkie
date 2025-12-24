package lib

import (
	"fmt"
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/structs"
)

// TODO: fix name creation

// 0x0000000000013AA0
// __int64 __fastcall scePthreadMutexInit(_QWORD *a1, __int64 a2, __int64 a3)
func libKernel_scePthreadMutexInit(mutexHandlePtr uintptr, attrPtr uintptr, namePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_init(mutexHandlePtr, attrPtr)
	if err != 0 {
		return err - 0x7FFE0000
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
	realNamePtr := GlobalGoAllocator.Malloc(uintptr(len(name) + 1))
	WriteCString(realNamePtr, name)
	mutex.NamePtr = realNamePtr

	// TODO: emulate __sys_namedobj_create?

	return 0
}

// 0x0000000000013C50
// __int64 scePthreadMutexDestroy()
func libKernel_scePthreadMutexDestroy(mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_destroy(mutexHandlePtr)
	if err != 0 {
		return err - 0x7FFE0000
	}

	return 0
}

// 0x0000000000013C70
// __int64 __fastcall scePthreadMutexLock(__int64 *, __int64, int, int, int, int)
func libKernel_scePthreadMutexLock(mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err - 0x7FFE0000
	}

	return 0
}

// 0x0000000000013CD0
// __int64 __fastcall scePthreadMutexUnlock(__int64 *, __int64, __int64, __int64)
func libKernel_scePthreadMutexUnlock(mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err - 0x7FFE0000
	}

	return 0
}
