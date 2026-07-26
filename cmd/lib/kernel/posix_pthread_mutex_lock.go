package kernel

import (
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/mutex"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000030FA0
// __int64 __fastcall pthread_mutex_trylock(unsigned __int64 *, __int64, int, int, int, int)
func libKernel_pthread_mutex_trylock(mutexHandlePtr *uintptr) uintptr {
	if mutexHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_trylock"),
		)
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *mutexHandlePtr
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
		if err := posix.InitStaticMutex(mutexHandlePtr, mutexAddr); err != 0 {
			MutexLock.Unlock()
			return err
		}
		MutexLock.Unlock()
		mutexAddr = *mutexHandlePtr
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
					color.Magenta.Sprint("pthread_mutex_trylock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Green.Sprintf("%d", mutex.Count),
				)
			}
			return 0
		case uint32(PthreadMutexTypeAdaptiveNp), uint32(PthreadMutexTypeErrorCheck):
			if logger.LogSyncingFail {
				logger.Printf("%-132s %s tried to lock a mutex %s it already owns.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_trylock"),
					GetMutexNameText(mutex, mutexAddr),
				)
			}
			return EDEADLK
		}
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

// 0x0000000000030070
// __int64 __fastcall pthread_mutex_timedlock(__int64 *, __int64, int, int, int, int)
func libKernel_pthread_mutex_timedlock(mutexHandlePtr *uintptr, timestamp *Timestamp) uintptr {
	if mutexHandlePtr == nil || timestamp == nil {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_timedlock"),
		)
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *mutexHandlePtr
	if mutexAddr <= ThrMutexDestroyed {
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to unlock destroyed mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_timedlock"),
			)
			return EINVAL
		}
		logger.Printf("%-132s %s failed trying to unlock uninitialized mutex.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_timedlock"),
		)
		return EPERM
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
					color.Magenta.Sprint("pthread_mutex_timedlock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Green.Sprintf("%d", mutex.Count),
				)
			}
			return 0
		case uint32(PthreadMutexTypeAdaptiveNp), uint32(PthreadMutexTypeErrorCheck):
			if logger.LogSyncingFail {
				logger.Printf("%-132s %s tried to lock a mutex %s it already owns.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_timedlock"),
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
						color.Magenta.Sprint("pthread_mutex_timedlock"),
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
						color.Magenta.Sprint("pthread_mutex_timedlock"),
						GetMutexNameText(mutex, mutexAddr),
					)
				}
				return 0
			}
			yieldCount--
		}
	}

	// Calculate actual timeout from absolute time.
	targetTime := time.Unix(int64(timestamp.Seconds), int64(timestamp.Nanoseconds))
	timeout := time.Until(targetTime)
	if timeout <= 0 {
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on mutex %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_timedlock"),
				GetMutexNameText(mutex, mutexAddr),
			)
		}
		return ETIMEDOUT
	}

	// Try locking mutex with timeout.
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on mutex %s for %s microseconds.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_timedlock"),
			GetMutexNameText(mutex, mutexAddr),
			color.Green.Sprintf("%d", timeout.Microseconds()),
		)
	}
	waited := MutexLockTimeout(hostMutex, timeout)
	if !waited {
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on mutex %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_timedlock"),
				GetMutexNameText(mutex, mutexAddr),
			)
		}
		return ETIMEDOUT
	}
	mutex.Owner = threadPtr
	mutex.Count = 1

	return 0
}

// 0x0000000000031FF0
// __int64 __fastcall pthread_mutex_reltimedlock_np(__int64 *, unsigned int, int, int, int, int)
func libKernel_pthread_mutex_reltimedlock_np(mutexHandlePtr *uintptr, micros uintptr) uintptr {
	if mutexHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
		)
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *mutexHandlePtr
	if mutexAddr <= ThrMutexDestroyed {
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to unlock destroyed mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
			)
			return EINVAL
		}
		logger.Printf("%-132s %s failed trying to unlock uninitialized mutex.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
		)
		return EPERM
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
					color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
					GetMutexNameText(mutex, mutexAddr),
					color.Green.Sprintf("%d", mutex.Count),
				)
			}
			return 0
		case uint32(PthreadMutexTypeAdaptiveNp), uint32(PthreadMutexTypeErrorCheck):
			if logger.LogSyncingFail {
				logger.Printf("%-132s %s tried to lock a mutex %s it already owns.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
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
						color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
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
						color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
						GetMutexNameText(mutex, mutexAddr),
					)
				}
				return 0
			}
			yieldCount--
		}
	}

	// Calculate timeout.
	timeout := time.Duration(micros) * time.Microsecond

	// Try locking mutex with timeout.
	if logger.LogSyncing {
		logger.Printf("%-132s %s waiting on mutex %s for %s microseconds.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
			GetMutexNameText(mutex, mutexAddr),
			color.Green.Sprintf("%d", timeout.Microseconds()),
		)
	}
	waited := MutexLockTimeout(hostMutex, timeout)
	if !waited {
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s timed out on mutex %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_reltimedlock_np"),
				GetMutexNameText(mutex, mutexAddr),
			)
		}
		return ETIMEDOUT
	}
	mutex.Owner = threadPtr
	mutex.Count = 1

	return 0
}

func MutexLockTimeout(mutex *sync.Mutex, timeout time.Duration) bool {
	if mutex.TryLock() {
		return true
	}
	start := time.Now()
	for time.Since(start) < timeout {
		runtime.Gosched()
		if mutex.TryLock() {
			return true
		}
	}

	return false
}
