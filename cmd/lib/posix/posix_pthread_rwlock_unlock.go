package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/rwlock"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_rwlock_unlock(rwlockHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_rwlock_unlock(rwlockHandlePtr)
}

func libScePosix_pthread_rwlock_unlock(rwlockHandlePtr *uintptr) uintptr {
	if rwlockHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid rwlock pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_unlock"),
		)
		return EINVAL
	}

	// Try initializing a rwlock, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	rwlockAddr := *rwlockHandlePtr
	if rwlockAddr <= ThrRwlockDestroyed {
		if rwlockAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to unlock destroyed rwlock.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_rwlock_unlock"),
			)
			return EINVAL
		}
		logger.Printf("%-132s %s failed trying to unlock uninitialized rwlock.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_unlock"),
		)
		return EPERM
	}

	// Unlock the rwlock.
	rwLock := (*PthreadRwlock)(unsafe.Pointer(rwlockAddr))
	if logger.LogSyncing {
		logger.Printf("%-132s %s unlocking rwlock %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_unlock"),
			GetRwlockNameText(rwLock, rwlockAddr),
		)
	}
	hostRwlock := rwlock.GetRwlock(rwlockAddr)
	if hostRwlock.Owner == threadPtr {
		hostRwlock.Owner = 0
		hostRwlock.Mu.Unlock()
	} else {
		hostRwlock.Mu.RUnlock()
	}

	return 0
}
