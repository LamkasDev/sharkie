package lib

// 0x0000000000013780
// __int64 __fastcall scePthreadCondBroadcast(__int64 *, __int64, int, int, int, int)
func libKernel_scePthreadCondBroadcast(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_broadcast(condHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}
