package lib

// 0x00000000000137A0
// __int64 scePthreadCondInit()
func libKernel_scePthreadCondInit(condHandlePtr, attrHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_init(condHandlePtr, attrHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}

// 0x0000000000013840
// __int64 scePthreadCondDestroy()
func libKernel_scePthreadCondDestroy(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_destroy(condHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}

// 0x0000000000013780
// __int64 __fastcall scePthreadCondBroadcast(__int64 *, __int64, int, int, int, int)
func libKernel_scePthreadCondBroadcast(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_broadcast(condHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}

// 0x0000000000013860
// __int64 scePthreadCondSignal()
func libKernel_scePthreadCondSignal(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_signal(condHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}

// 0x00000000000138C0
// __int64 scePthreadCondWait()
func libKernel_scePthreadCondWait(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_wait(condHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}
