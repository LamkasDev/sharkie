package lib

import (
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000002FDF0
// __int64 __fastcall pthread_mutex_init(__int64, _QWORD *)
func libKernel_pthread_mutex_init(mutexHandlePtr uintptr, attrHandlePtr uintptr) uintptr {
	mutexAddr := GlobalGoAllocator.Malloc(PthreadMutexSize)
	if mutexAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	mutex.Lock = 0
	mutex.Flags = uint32(PthreadMutexTypeRecursive)
	mutex.Owner = 0
	mutex.Count = 0
	mutex.SpinLoops = 0
	mutex.YieldLoops = 0
	mutex.Protocol = PthreadMutexProtocolNone

	// Apply attributes.
	attr, err := ResolveHandle[PthreadMutexAttr](attrHandlePtr)
	if err == 0 {
		if attr.Type < PthreadMutexTypeErrorCheck || attr.Type > PthreadMutexTypeAdaptiveNp ||
			attr.Protocol > PthreadMutexProtocolProtect {
			tempHandleAddr := mutexAddr
			if err = libKernel_pthread_mutex_destroy(tempHandleAddr); err != 0 {
				return err
			}
			logger.Printf("%-132s %s failed due to invalid attribute.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_init"),
			)
			return EINVAL
		}

		mutex.Flags = uint32(attr.Type)
		mutex.Protocol = attr.Protocol
		if attr.Type == PthreadMutexTypeAdaptiveNp {
			mutex.SpinLoops = 2000
		}
	}

	// Copy the pointer back to mutexHandlePtr.
	WriteAddress(mutexHandlePtr, mutexAddr)

	logger.Printf("%-132s %s created mutex at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_init"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}

func libKernel_initStaticMutex(mutexHandlePtr uintptr, initType uintptr) uintptr {
	mutexAddr := GlobalGoAllocator.Malloc(PthreadMutexSize)
	if mutexAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	mutex.Lock = 0
	if initType == ThrAdaptiveMutexInitializer {
		mutex.Flags = uint32(PthreadMutexTypeAdaptiveNp)
		mutex.SpinLoops = 2000
	} else {
		mutex.Flags = uint32(PthreadMutexTypeRecursive)
		mutex.SpinLoops = 0
	}
	mutex.Owner = 0
	mutex.Count = 0
	mutex.YieldLoops = 0
	mutex.Protocol = PthreadMutexProtocolNone

	// Copy the pointer back to mutexHandlePtr.
	WriteAddress(mutexHandlePtr, mutexAddr)

	logger.Printf("%-132s %s created mutex at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("libKernel_initStaticMutex"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}

// 0x0000000000030CB0
// __int64 __fastcall pthread_mutex_destroy(__int64 *)
func libKernel_pthread_mutex_destroy(mutexHandlePtr uintptr) uintptr {
	// Resolve the handle.
	mutex, err := ResolveHandle[PthreadMutex](mutexHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_destroy"),
		)
		return err
	}

	// Free the memory.
	mutexAddr := uintptr(unsafe.Pointer(mutex))
	if !GlobalGoAllocator.Free(mutexAddr) {
		logger.Printf("%-132s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_destroy"),
		)
		return EFAULT
	}

	logger.Printf("%-132s %s destroyed mutex %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_destroy"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}

// TODO: shouldn't we do the other one too?
// 0x0000000000031B40
// __int64 __fastcall sub_31B40(__int64 *, __int64, int, int, int, int)
func libKernel_pthread_mutex_lock(mutexHandlePtr uintptr) uintptr {
	if mutexHandlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_lock"),
		)
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	if mutexAddr <= ThrMutexDestroyed {
		MutexLock.Lock()
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to lock destroyed mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_lock"),
			)
			MutexLock.Unlock()
			return EINVAL
		}
		if err := libKernel_initStaticMutex(mutexHandlePtr, mutexAddr); err != 0 {
			MutexLock.Unlock()
			return err
		}
		MutexLock.Unlock()
		mutexAddr = *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	}

	// Process special mutex types.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if mutex.Owner == threadPtr {
		mutexType := mutex.Flags & PthreadMutexTypeMask
		switch mutexType {
		case uint32(PthreadMutexTypeAdaptiveNp), uint32(PthreadMutexTypeRecursive):
			mutex.Count++
			if logger.LogSyncing {
				logger.Printf("%-132s %s incremented recursive/adaptive mutex %s (count=%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_lock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Green.Sprintf("%d", mutex.Count),
				)
			}
			return 0
		}
		logger.Printf("%-132s %s tried to lock a mutex %s it already owns.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_lock"),
			GetMutexNameText(mutex, mutexAddr),
		)
		return EDEADLK
	}

	hostMutex := GetMutex(mutexAddr)

	// For adaptive mutexes, spin for a bit.
	if mutex.Protocol == PthreadMutexProtocolNone {
		spinCount := mutex.SpinLoops
		for spinCount > 0 {
			if hostMutex.TryLock() {
				mutex.Owner = threadPtr
				if logger.LogSyncing {
					logger.Printf("%-132s %s locked mutex %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("pthread_mutex_lock"),
						GetMutexNameText(mutex, mutexAddr),
					)
				}
				return 0
			}
			spinCount--
		}

		yieldCount := mutex.YieldLoops
		for yieldCount > 0 {
			runtime.Gosched()
			if hostMutex.TryLock() {
				mutex.Owner = threadPtr
				if logger.LogSyncing {
					logger.Printf("%-132s %s locked mutex %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("pthread_mutex_lock"),
						GetMutexNameText(mutex, mutexAddr),
					)
				}
				return 0
			}
			yieldCount--
		}
	}

	// Fallback to a blocking lock.
	if logger.LogSyncing {
		logger.Printf("%-132s %s locking mutex %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_lock"),
			GetMutexNameText(mutex, mutexAddr),
		)
	}
	hostMutex.Lock()
	mutex.Owner = threadPtr
	mutex.Count = 1

	return 0
}

// 0x0000000000030FA0
// __int64 __fastcall pthread_mutex_trylock(unsigned __int64 *, __int64, int, int, int, int)
func libKernel_pthread_mutex_trylock(mutexHandlePtr uintptr) uintptr {
	if mutexHandlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_trylock"),
		)
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	if mutexAddr <= ThrMutexDestroyed {
		MutexLock.Lock()
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to lock destroyed mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_trylock"),
			)
			MutexLock.Unlock()
			return EINVAL
		}
		if err := libKernel_initStaticMutex(mutexHandlePtr, mutexAddr); err != 0 {
			MutexLock.Unlock()
			return err
		}
		MutexLock.Unlock()
		mutexAddr = *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	}

	// Process special mutex types.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if mutex.Owner == threadPtr {
		mutexType := mutex.Flags & PthreadMutexTypeMask
		switch mutexType {
		case uint32(PthreadMutexTypeAdaptiveNp), uint32(PthreadMutexTypeRecursive):
			mutex.Count++
			if logger.LogSyncing {
				logger.Printf("%-132s %s incremented recursive/adaptive mutex %s (count=%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_trylock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Green.Sprintf("%d", mutex.Count),
				)
			}
			return 0
		}
		logger.Printf("%-132s %s tried to lock a mutex %s it already owns.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_trylock"),
			GetMutexNameText(mutex, mutexAddr),
		)
		return EBUSY
	}

	hostMutex := GetMutex(mutexAddr)

	// Fallback to a blocking lock.
	if !hostMutex.TryLock() {
		if logger.LogSyncing {
			logger.Printf("%-132s %s tried to lock a mutex %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_trylock"),
				GetMutexNameText(mutex, mutexAddr),
			)
		}
		return EBUSY
	}
	mutex.Owner = threadPtr
	mutex.Count = 1
	if logger.LogSyncing {
		logger.Printf("%-132s %s locked mutex %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_trylock"),
			GetMutexNameText(mutex, mutexAddr),
		)
	}

	return 0
}

// 0x00000000000327F0
// __int64 __fastcall sub_327F0(__int64 *, __int64, __int64, __int64)
func libKernel_pthread_mutex_unlock(mutexHandlePtr uintptr) uintptr {
	if mutexHandlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
		)
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	if mutexAddr <= ThrMutexDestroyed {
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to unlock destroyed mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_unlock"),
			)
			return EINVAL
		}
		logger.Printf("%-132s %s failed trying to unlock uninitialized mutex.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
		)
		return EPERM
	}

	// Check permissions.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if mutex.Owner != threadPtr {
		logger.Printf("%-132s %s failed trying to unlock unowned mutex.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
		)
		return EPERM
	}

	// Handle special mutex types.
	mutexType := mutex.Flags & PthreadMutexTypeMask
	shouldReleaseHost := true
	if (mutexType == uint32(PthreadMutexTypeAdaptiveNp) || mutexType == uint32(PthreadMutexTypeRecursive)) && mutex.Count > 0 {
		mutex.Count--
		if logger.LogSyncing {
			logger.Printf("%-132s %s decremented recursive/adaptive mutex %s (count=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_unlock"),
				GetMutexNameText(mutex, mutexAddr),
				color.Green.Sprintf("%d", mutex.Count),
			)
		}
		if mutex.Count > 0 {
			shouldReleaseHost = false
			return 0
		}
	}
	if !shouldReleaseHost {
		return 0
	}

	// Unlock the mutex.
	if logger.LogSyncing {
		logger.Printf("%-132s %s unlocking mutex %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
			GetMutexNameText(mutex, mutexAddr),
		)
	}
	mutex.Owner = 0
	hostMutex := GetMutex(mutexAddr)
	hostMutex.Unlock()

	return 0
}
