package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

// 0x0000000000013D10
// __int64 __fastcall scePthreadRwlockInit(_QWORD *, __int64, __int64)
func libKernel_scePthreadRwlockInit(rwlockHandlePtr, attrHandlePtr *uintptr, namePtr Cstring) uintptr {
	err := posix.Pthread_rwlock_init(rwlockHandlePtr, attrHandlePtr)
	if err != 0 {
		return err
	}

	return 0
}

// 0x0000000000013DD0
// __int64 scePthreadRwlockRdlock()
func libKernel_scePthreadRwlockRdlock(rwlockHandlePtr *uintptr) uintptr {
	err := posix.Pthread_rwlock_rdlock(rwlockHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013E90
// __int64 scePthreadRwlockWrlock()
func libKernel_scePthreadRwlockWrlock(rwlockHandlePtr *uintptr) uintptr {
	err := posix.Pthread_rwlock_wrlock(rwlockHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013E70
// __int64 scePthreadRwlockUnlock()
func libKernel_scePthreadRwlockUnlock(rwlockHandlePtr *uintptr) uintptr {
	err := posix.Pthread_rwlock_unlock(rwlockHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}
