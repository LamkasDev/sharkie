package posix

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/cond"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func InitStaticCond(condHandlePtr *uintptr) uintptr {
	condAddr := GlobalGoAllocator.Malloc(PthreadCondSize)
	if condAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))
	cond.KernelId = 0
	cond.Flags = 0

	// Copy the pointer back to condHandlePtr.
	*condHandlePtr = condAddr

	logger.Printf("%-132s %s created cond at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("libKernel_initStaticCond"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)
	return 0
}

func Pthread_cond_init(condHandlePtr, attrHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_cond_init(condHandlePtr, attrHandlePtr)
}

func libScePosix_pthread_cond_init(condHandlePtr, attrHandlePtr *uintptr) uintptr {
	condAddr := GlobalGoAllocator.Malloc(PthreadCondSize)
	if condAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))
	cond.KernelId = 0
	cond.Flags = 0

	// Copy the pointer back to condHandlePtr.
	*condHandlePtr = condAddr

	logger.Printf("%-132s %s created cond at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_init"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)
	return 0
}

func Pthread_cond_destroy(condHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_cond_destroy(condHandlePtr)
}

func libScePosix_pthread_cond_destroy(condHandlePtr *uintptr) uintptr {
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

func Pthread_cond_broadcast(condHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_cond_broadcast(condHandlePtr)
}

func libScePosix_pthread_cond_broadcast(condHandlePtr *uintptr) uintptr {
	if condHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_broadcast"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *condHandlePtr
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := InitStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *condHandlePtr
	}
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))

	// Broadcast to it.
	hostCond := GetCond(condAddr)
	hostCond.Broadcast()

	if logger.LogSyncing {
		logger.Printf("%-132s %s broadcasted cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_broadcast"),
			GetCondNameText(cond, condAddr),
		)
	}
	return 0
}

func Pthread_cond_signal(condHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_cond_signal(condHandlePtr)
}

func libScePosix_pthread_cond_signal(condHandlePtr *uintptr) uintptr {
	if condHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_signal"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *condHandlePtr
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := InitStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *condHandlePtr
	}
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))

	// Signal to it.
	hostCond := GetCond(condAddr)
	hostCond.Signal()

	if logger.LogSyncing {
		logger.Printf("%-132s %s signaled cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_signal"),
			GetCondNameText(cond, condAddr),
		)
	}
	return 0
}

func Pthread_cond_wait(condHandlePtr, mutexHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_cond_wait(condHandlePtr, mutexHandlePtr)
}

func libScePosix_pthread_cond_wait(condHandlePtr, mutexHandlePtr *uintptr) uintptr {
	if condHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_wait"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *condHandlePtr
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := InitStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *condHandlePtr
	}
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))

	hostCond := GetCond(condAddr)
	hostCond.Mutex.Lock()
	w := hostCond.WaitChan()
	hostCond.Mutex.Unlock()
	err := Pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_wait"),
			GetCondNameText(cond, condAddr),
		)
	}
	<-w
	err = Pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err
	}

	return 0
}

func Pthread_cond_timedwait(condHandlePtr, mutexHandlePtr *uintptr, timestamp *Timestamp) uintptr {
	return libScePosix_pthread_cond_timedwait(condHandlePtr, mutexHandlePtr, timestamp)
}

func libScePosix_pthread_cond_timedwait(condHandlePtr, mutexHandlePtr *uintptr, timestamp *Timestamp) uintptr {
	if condHandlePtr == nil || timestamp == nil {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_timedwait"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *condHandlePtr
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := InitStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *condHandlePtr
	}
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))

	// Calculate actual timeout from absolute time.
	targetTime := time.Unix(int64(timestamp.Seconds), int64(timestamp.Nanoseconds))
	timeout := time.Until(targetTime)
	if timeout <= 0 {
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on cond %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_cond_timedwait"),
				GetCondNameText(cond, condAddr),
			)
		}
		return ETIMEDOUT
	}

	// Unlock mutex, wait on condition and relock mutex.
	hostCond := GetCond(condAddr)
	hostCond.Mutex.Lock()
	w := hostCond.WaitChan()
	hostCond.Mutex.Unlock()
	err := Pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on cond %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_timedwait"),
			GetCondNameText(cond, condAddr),
		)
	}

	select {
	case <-w:
	case <-time.After(timeout):
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on cond %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_cond_timedwait"),
				GetCondNameText(cond, condAddr),
			)
		}
		Pthread_mutex_lock(mutexHandlePtr)
		return ETIMEDOUT
	}
	err = Pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err
	}

	return 0
}

func Pthread_cond_reltimedwait_np(condHandlePtr, mutexHandlePtr *uintptr, micros uintptr) uintptr {
	return libScePosix_pthread_cond_reltimedwait_np(condHandlePtr, mutexHandlePtr, micros)
}

func libScePosix_pthread_cond_reltimedwait_np(condHandlePtr, mutexHandlePtr *uintptr, micros uintptr) uintptr {
	if condHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid cond pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_reltimedwait_np"),
		)
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *condHandlePtr
	if condAddr == PthreadCondInitializer {
		CondLock.Lock()
		if err := InitStaticCond(condHandlePtr); err != 0 {
			CondLock.Unlock()
			return err
		}
		CondLock.Unlock()
		condAddr = *condHandlePtr
	}
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))

	// Calculate timeout.
	timeout := time.Duration(micros) * time.Microsecond

	// Unlock mutex, perform a timed wait on condition and relock mutex.
	hostCond := GetCond(condAddr)
	hostCond.Mutex.Lock()
	w := hostCond.WaitChan()
	hostCond.Mutex.Unlock()
	err := Pthread_mutex_unlock(mutexHandlePtr)
	if err != 0 {
		return err
	}
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on cond %s for %s microseconds.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cond_reltimedwait_np"),
			GetCondNameText(cond, condAddr),
			color.Green.Sprintf("%d", timeout.Microseconds()),
		)
	}

	select {
	case <-w:
	case <-time.After(timeout):
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on cond %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_cond_reltimedwait_np"),
				GetCondNameText(cond, condAddr),
			)
		}
		Pthread_mutex_lock(mutexHandlePtr)
		return ETIMEDOUT
	}
	err = Pthread_mutex_lock(mutexHandlePtr)
	if err != 0 {
		return err
	}

	return 0
}
