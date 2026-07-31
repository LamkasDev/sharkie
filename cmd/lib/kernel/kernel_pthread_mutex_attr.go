package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

// 0x00000000000139E0
// __int64 __fastcall scePthreadMutexattrInit(__int64 *)
func libKernel_scePthreadMutexattrInit(attrHandlePtr *uintptr) uintptr {
	err := posix.Pthread_mutexattr_init(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013A60
// __int64 __fastcall scePthreadMutexattrSettype(_DWORD **, int)
func libKernel_scePthreadMutexattrSettype(attrHandlePtr *uintptr, attrType uintptr) uintptr {
	err := posix.Pthread_mutexattr_settype(attrHandlePtr, attrType)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013A00
// __int64 __fastcall scePthreadMutexattrDestroy(__int64 *)
func libKernel_scePthreadMutexattrDestroy(attrHandlePtr *uintptr) uintptr {
	err := posix.Pthread_attr_destroy(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}
