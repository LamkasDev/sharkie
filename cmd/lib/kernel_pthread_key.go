package lib

import . "github.com/LamkasDev/sharkie/cmd/structs"

// 0x00000000000139A0
// __int64 __fastcall scePthreadKeyCreate(_DWORD *, __int64)
func libKernel_scePthreadKeyCreate(keyPtr, destructor uintptr) uintptr {
	err := libKernel_pthread_key_create(keyPtr, destructor)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013950
// __int64 scePthreadGetspecific()
func libKernel_scePthreadGetspecific(key uint32) uintptr {
	err := libKernel_pthread_getspecific(key)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000034A40
// __int64 __fastcall scePthreadSetspecific(__int64, __int64)
func libKernel_scePthreadSetspecific(key uint32, value uintptr) uintptr {
	err := libKernel_pthread_setspecific(key, value)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}
