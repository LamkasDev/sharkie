package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/rwlock"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_rwlock_rdlock(rwlockHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_rwlock_rdlock(rwlockHandlePtr)
}

func libScePosix_pthread_rwlock_rdlock(rwlockHandlePtr *uintptr) uintptr {
	if rwlockHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid rwlock pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_rdlock"),
		)
		return EINVAL
	}

	// Try initializing a rwlock, if it wasn't initialized yet.
	rwlockAddr := *rwlockHandlePtr
	if rwlockAddr <= ThrRwlockDestroyed {
		RwlockLock.Lock()
		if rwlockAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to lock destroyed rwlock.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_rwlock_rdlock"),
			)
			RwlockLock.Unlock()
			return EINVAL
		}
		if err := InitStaticRwlock(rwlockHandlePtr); err != 0 {
			RwlockLock.Unlock()
			return err
		}
		RwlockLock.Unlock()
		rwlockAddr = *rwlockHandlePtr
	}

	// Lock the rwlock.
	rwLock := (*PthreadRwlock)(unsafe.Pointer(rwlockAddr))
	hostRwlock := GetRwlock(rwlockAddr)
	hostRwlock.Mu.RLock()
	if logger.LogSyncing {
		logger.Printf("%-132s %s locked rwlock %s for reading.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_rdlock"),
			GetRwlockNameText(rwLock, rwlockAddr),
		)
	}

	return 0
}

func Pthread_rwlock_wrlock(rwlockHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_rwlock_wrlock(rwlockHandlePtr)
}

func libScePosix_pthread_rwlock_wrlock(rwlockHandlePtr *uintptr) uintptr {
	if rwlockHandlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid rwlock pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_wrlock"),
		)
		return EINVAL
	}

	// Try initializing a rwlock, if it wasn't initialized yet.
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	rwlockAddr := *rwlockHandlePtr
	if rwlockAddr <= ThrRwlockDestroyed {
		RwlockLock.Lock()
		if rwlockAddr == ThrMutexDestroyed {
			logger.Printf("%-132s %s failed trying to lock destroyed rwlock.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_rwlock_wrlock"),
			)
			RwlockLock.Unlock()
			return EINVAL
		}
		if err := InitStaticRwlock(rwlockHandlePtr); err != 0 {
			RwlockLock.Unlock()
			return err
		}
		RwlockLock.Unlock()
		rwlockAddr = *rwlockHandlePtr
	}

	// Lock the rwlock.
	rwLock := (*PthreadRwlock)(unsafe.Pointer(rwlockAddr))
	hostRwlock := GetRwlock(rwlockAddr)
	hostRwlock.Mu.Lock()
	hostRwlock.Owner = threadPtr
	if logger.LogSyncing {
		logger.Printf("%-132s %s locked rwlock %s for writing.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_wrlock"),
			GetRwlockNameText(rwLock, rwlockAddr),
		)
	}

	return 0
}
