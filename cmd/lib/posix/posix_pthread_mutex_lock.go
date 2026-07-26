package posix

import (
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/mutex"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_mutex_lock(mutexHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_mutex_lock(mutexHandlePtr)
}

func libScePosix_pthread_mutex_lock(mutexHandlePtr *uintptr) uintptr {
	if mutexHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_lock"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	if mutexAddr <= ThrMutexDestroyed {
		MutexLock.Lock()
		if mutexAddr == ThrMutexDestroyed {
			a := emu.SprintStackTrace()
			_ = a
			logger.Printf("%-132s %s failed trying to lock destroyed mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_lock"),
			)
			MutexLock.Unlock()
			emu.SetErrno(EINVAL)
			return EINVAL
		}
		if err := InitStaticMutex(mutexHandlePtr, mutexAddr); err != 0 {
			MutexLock.Unlock()
			emu.SetErrno(err)
			return ERR_PTR
		}
		MutexLock.Unlock()
		mutexAddr = *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	}

	// Process special mutex types.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if mutex.Owner == threadPtr {
		mutexType := mutex.Flags & PthreadMutexTypeMask
		switch mutexType {
		case uint32(PthreadMutexTypeRecursive):
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
		case uint32(PthreadMutexTypeAdaptiveNp), uint32(PthreadMutexTypeErrorCheck):
			if logger.LogSyncingFail {
				logger.Printf("%-132s %s tried to lock a mutex %s it already owns.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_lock"),
					GetMutexNameText(mutex, mutexAddr),
				)
			}
			return EDEADLK
		}
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
