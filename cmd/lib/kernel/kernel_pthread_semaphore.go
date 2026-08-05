package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/semaphore"
)

// 0x0000000000013F70
// __int64 __fastcall scePthreadSemInit(__int64, int, __int64, __int64)
func libKernel_scePthreadSemInit(semaphore *PSemaphore, flag, value uintptr, name Cstring) uintptr {
	err := posix.Sem_init(semaphore, 0, value)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013FF0
// __int64 __fastcall scePthreadSemDestroy(__int64)
func libKernel_scePthreadSemDestroy(semaphore *PSemaphore) uintptr {
	err := posix.Sem_destroy(semaphore)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000140B0
// __int64 __fastcall scePthreadSemTrywait(__int64)
func libKernel_scePthreadSemTrywait(semaphore *PSemaphore) uintptr {
	err := posix.Sem_trywait(semaphore)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000140E0
// __int64 __fastcall scePthreadSemWait(__int64)
func libKernel_scePthreadSemWait(semaphore *PSemaphore) uintptr {
	err := posix.Sem_wait(semaphore)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014020
// __int64 scePthreadSemTimedwait()
func libKernel_scePthreadSemTimedwait(semaphore *PSemaphore, micros uint32) uintptr {
	err := posix.Sem_reltimedwait_np(semaphore, micros)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014080
// __int64 __fastcall scePthreadSemPost(__int64)
func libKernel_scePthreadSemPost(semaphore *PSemaphore) uintptr {
	err := posix.Sem_post(semaphore)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}
