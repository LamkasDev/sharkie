package lib

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
	"github.com/gookit/color"
)

// 0x0000000000004CA0
// __int64 __fastcall pthread_cond_init(_QWORD *, _DWORD **)
func libKernel_pthread_cond_init(condHandlePtr, attrHandlePtr uintptr) uintptr {
	condAddr := GlobalGoAllocator.Malloc(PthreadCondSize)
	if condAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))
	cond.KernelId = 0
	cond.Flags = 0

	// Copy the pointer back to condHandlePtr.
	WriteAddress(condHandlePtr, condAddr)

	logger.Printf("%-132s %s created cond at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_init"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)
	return 0
}

func libKernel_initStaticCond(condHandlePtr uintptr) uintptr {
	condAddr := GlobalGoAllocator.Malloc(PthreadCondSize)
	if condAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))
	cond.KernelId = 0
	cond.Flags = 0

	// Copy the pointer back to condHandlePtr.
	WriteAddress(condHandlePtr, condAddr)

	logger.Printf("%-132s %s created cond at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("libKernel_initStaticCond"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)
	return 0
}

// 0x0000000000004D60
// __int64 __fastcall pthread_cond_destroy(__int64 *)
func libKernel_pthread_cond_destroy(condHandlePtr uintptr) uintptr {
	// Resolve the handle.
	cond, err := ResolveHandle[PthreadCond](condHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_destroy"),
		)
		return err
	}

	// Free the memory.
	condAddr := uintptr(unsafe.Pointer(cond))
	if !GlobalGoAllocator.Free(condAddr) {
		logger.Printf("%-132s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_destroy"),
		)
		return EFAULT
	}

	logger.Printf("%-132s %s destroyed cond %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_destroy"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)
	return 0
}

// 0x0000000000006150
// __int64 __fastcall pthread_cond_broadcast(__int64 *, __int64, int, int, int, int)
func libKernel_pthread_cond_broadcast(condHandlePtr uintptr) uintptr {
	if condHandlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_broadcast"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
	}

	// Broadcast to it.
	hostCond := GetCond(condAddr)
	hostCond.Broadcast()

	if logger.LogSyncing {
		logger.Printf("%-132s %s broadcasted cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_broadcast"),
			color.Yellow.Sprintf("0x%X", condAddr),
		)
	}
	return 0
}

// 0x0000000000005AD0
// __int64 __fastcall pthread_cond_signal(__int64 *, __m128, __m128, __m128, __m128, __m128, __m128, __m128, __m128, __int64, __int64, __int64, __int64, __int64)
func libKernel_pthread_cond_signal(condHandlePtr uintptr) uintptr {
	if condHandlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_signal"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
	}

	// Signal to it.
	hostCond := GetCond(condAddr)
	hostCond.Signal()

	if logger.LogSyncing {
		logger.Printf("%-132s %s signaled cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_signal"),
			color.Yellow.Sprintf("0x%X", condAddr),
		)
	}
	return 0
}

// 0x0000000000005550
// __int64 __fastcall pthread_cond_wait(__int64 *, unsigned __int64 *, __int64, __int64, __int64, int)
func libKernel_pthread_cond_wait(condHandlePtr uintptr, mutexHandlePtr uintptr) uintptr {
	if condHandlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_wait"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
	}

	// Unlock mutex, wait on condition and relock mutex.
	err := libKernel_pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_wait"),
			color.Yellow.Sprintf("0x%X", condAddr),
		)
	}
	hostCond := GetCond(condAddr)
	hostCond.L.Lock()
	hostCond.Wait()
	hostCond.L.Unlock()
	err = libKernel_pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err
	}

	return 0
}

// 0x00000000000056B0
// __int64 __fastcall pthread_cond_timedwait(__int64 *, unsigned __int64 *, __int64, __int64, __int64, int)
func libKernel_pthread_cond_timedwait(condHandlePtr uintptr, mutexHandlePtr uintptr, timestampPtr uintptr) uintptr {
	if condHandlePtr == 0 || timestampPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_timedwait"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
	}

	// Calculate actual timeout from absolute time.
	timestamp := (*Timestamp)(unsafe.Pointer(timestampPtr))
	targetTime := time.Unix(int64(timestamp.Seconds), int64(timestamp.Nanoseconds))
	timeout := time.Until(targetTime)
	if timeout <= 0 {
		logger.Printf("%-132s %s timed out on cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_timedwait"),
			color.Yellow.Sprintf("0x%X", condAddr),
		)
		return ETIMEDOUT
	}

	// Unlock mutex, perform a timed wait on condition and relock mutex.
	err := libKernel_pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on cond %s for %s microseconds.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_timedwait"),
			color.Yellow.Sprintf("0x%X", condAddr),
			color.Green.Sprintf("%d", timeout.Microseconds()),
		)
	}
	hostCond := GetCond(condAddr)
	hostCond.L.Lock()
	waited := CondWaitTimeout(hostCond, timeout)
	hostCond.L.Unlock()
	err = libKernel_pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if !waited {
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on cond %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_cond_timedwait"),
				color.Yellow.Sprintf("0x%X", condAddr),
			)
		}
		return ETIMEDOUT
	}

	return 0
}

// 0x00000000000057D0
// __int64 __fastcall pthread_cond_reltimedwait_np(__int64 *, unsigned __int64 *, unsigned int, __int64, __int64, int)
func libKernel_pthread_cond_reltimedwait_np(condHandlePtr uintptr, mutexHandlePtr uintptr, micros uintptr) uintptr {
	if condHandlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_reltimedwait_np"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
	}

	// Calculate timeout.
	timeout := time.Duration(micros) * time.Microsecond

	// Unlock mutex, perform a timed wait on condition and relock mutex.
	err := libKernel_pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on cond %s for %s microseconds.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_reltimedwait_np"),
			color.Yellow.Sprintf("0x%X", condAddr),
			color.Green.Sprintf("%d", timeout.Microseconds()),
		)
	}
	hostCond := GetCond(condAddr)
	hostCond.L.Lock()
	waited := CondWaitTimeout(hostCond, timeout)
	hostCond.L.Unlock()
	err = libKernel_pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if !waited {
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on cond %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_cond_reltimedwait_np"),
				color.Yellow.Sprintf("0x%X", condAddr),
			)
		}
		return ETIMEDOUT
	}

	return 0
}
