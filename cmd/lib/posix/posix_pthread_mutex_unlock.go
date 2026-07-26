package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/mutex"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_mutex_unlock(mutexHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_mutex_unlock(mutexHandlePtr)
}

func libScePosix_pthread_mutex_unlock(mutexHandlePtr *uintptr) uintptr {
	if mutexHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	mutexAddr := *mutexHandlePtr
	if mutexAddr <= ThrMutexDestroyed {
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to unlock destroyed mutex.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_unlock"),
			)
			emu.SetErrno(EINVAL)
			return ERR_PTR
		}
		logger.Printf("%-132s %s failed trying to unlock uninitialized mutex.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
		)
		emu.SetErrno(EPERM)
		return ERR_PTR
	}

	// Check permissions.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if mutex.Owner != threadPtr {
		if logger.LogSyncingFail {
			logger.Printf("%-132s %s failed trying to unlock unowned mutex %s (owner=%s, caller=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_unlock"),
				GetMutexNameText(mutex, mutexAddr),
				color.Yellow.Sprintf("0x%X", mutex.Owner),
				color.Yellow.Sprintf("0x%X", threadPtr),
			)
		}
		emu.SetErrno(EPERM)
		return ERR_PTR
	}

	// Handle special mutex types.
	mutexType := mutex.Flags & PthreadMutexTypeMask
	shouldReleaseHost := true
	if mutexType == uint32(PthreadMutexTypeRecursive) && mutex.Count > 0 {
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
