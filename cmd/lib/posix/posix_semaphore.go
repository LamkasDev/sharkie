package posix

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/semaphore"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func libScePosix_sem_init(semaphore *PSemaphore, pShared, value uintptr) uintptr {
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_init"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Initialize to defaults.
	semaphore.Magic = PSemaphoreMagic
	semaphore.Flags = 0
	semaphore.WaitAddress = 0
	semaphore.Value = int32(value)
	semaphore.Pshared = 0
	if pShared != 0 {
		semaphore.Pshared = 1
	}

	logger.Printf("%-132s %s created semaphore at %s (value=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sem_init"),
		color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
		color.Yellow.Sprintf("0x%X", value),
	)
	return 0
}

func libScePosix_sem_destroy(semaphore *PSemaphore) uintptr {
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_destroy"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	if semaphore.Magic != PSemaphoreMagic {
		logger.Printf("%-132s %s failed due to invalid sem magic.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_destroy"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	semAddress := uintptr(unsafe.Pointer(semaphore))
	hostSemaphore := GetPSemaphore(semAddress)

	// Safely check for active waiters and remove from repo.
	PSemaphoreLock.Lock()
	hostSemaphore, exists := PSemaphoreRepo[semAddress]
	if exists {
		if atomic.LoadInt32(&hostSemaphore.Waiters) > 0 {
			logger.Printf("%-132s %s failed destroying semaphore %s (busy).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sem_destroy"),
				color.Yellow.Sprintf("0x%X", semAddress),
			)
			PSemaphoreLock.Unlock()
			emu.SetErrno(EBUSY)
			return ERR_PTR
		}
		delete(PSemaphoreRepo, semAddress)
	}
	PSemaphoreLock.Unlock()

	// Invalidate the magic.
	semaphore.Magic = 0

	logger.Printf("%-132s %s destroyed semaphore %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sem_destroy"),
		color.Yellow.Sprintf("0x%X", semAddress),
	)
	return 0
}

func libScePosix_sem_trywait(semaphore *PSemaphore) uintptr {
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_trywait"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	if semaphore.Magic != PSemaphoreMagic {
		logger.Printf("%-132s %s failed due to invalid sem magic.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_trywait"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Try decrement semaphore, otherwise return.
	for {
		value := atomic.LoadInt32(&semaphore.Value)
		if value <= 0 {
			if logger.LogSyncingFail {
				logger.Printf("%-132s %s tried waiting on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sem_timedwait"),
					color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
				)
			}
			emu.SetErrno(EAGAIN)
			return ERR_PTR
		}
		if atomic.CompareAndSwapInt32(&semaphore.Value, value, value-1) {
			if logger.LogSyncing {
				logger.Printf("%-132s %s waited on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sem_timedwait"),
					color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
				)
			}
			return 0
		}
	}
}

func libScePosix_sem_wait(semaphore *PSemaphore) uintptr {
	return libScePosix_sem_timedwait(semaphore, nil)
}

func libScePosix_sem_timedwait(semaphore *PSemaphore, timestamp *Timestamp) uintptr {
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_timedwait"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	if semaphore.Magic != PSemaphoreMagic {
		logger.Printf("%-132s %s failed due to invalid sem magic.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_timedwait"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Try decrement semaphore without host sync primitives.
	for {
		value := atomic.LoadInt32(&semaphore.Value)
		if value <= 0 {
			break
		}
		if atomic.CompareAndSwapInt32(&semaphore.Value, value, value-1) {
			if logger.LogSyncing {
				logger.Printf("%-132s %s waited on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sem_timedwait"),
					color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
				)
			}
			return 0
		}
	}

	// Calculate actual timeout from absolute time.
	timeout := time.Duration(-1)
	if timestamp != nil {
		targetTime := time.Unix(int64(timestamp.Seconds), int64(timestamp.Nanoseconds))
		timeout = time.Until(targetTime)
		if timeout <= 0 {
			if logger.LogSyncingFail {
				logger.Printf("%-132s %s timed out on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sem_timedwait"),
					color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
				)
			}
			emu.SetErrno(ETIMEDOUT)
			return ERR_PTR
		}
	}

	// Lock semaphore.
	start := time.Now()
	hostSemaphore := GetPSemaphore(uintptr(unsafe.Pointer(semaphore)))

	for {
		// Check value again (holding lock this time).
		hostSemaphore.Mutex.Lock()
		if semaphore.Value > 0 {
			semaphore.Value--
			if logger.LogSyncing {
				logger.Printf("%-132s %s waited on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sem_timedwait"),
					color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
				)
			}
			hostSemaphore.Mutex.Unlock()
			return 0
		}
		w := hostSemaphore.WaitChanNoLock()
		atomic.AddInt32(&hostSemaphore.Waiters, 1)
		hostSemaphore.Mutex.Unlock()

		var remaining time.Duration
		if timeout != -1 {
			remaining = timeout - time.Since(start)
			if remaining <= 0 {
				if logger.LogSyncingFail {
					logger.Printf("%-132s %s timed out on semaphore %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sem_timedwait"),
						color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
					)
				}
				emu.SetErrno(ETIMEDOUT)
				atomic.AddInt32(&hostSemaphore.Waiters, -1)
				return ERR_PTR
			}
		}

		// Wait.
		if logger.LogSyncing {
			logger.Printf("%-132s %s waiting on semaphore %s for %s microseconds.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sem_timedwait"),
				color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
				color.Green.Sprintf("%d", timeout.Microseconds()),
			)
		}
		if timeout == -1 {
			<-w
		} else {
			select {
			case <-w:
			case <-time.After(remaining):
				if logger.LogSyncingFail {
					logger.Printf("%-132s %s timed out on semaphore %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sem_timedwait"),
						color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
					)
				}
				emu.SetErrno(ETIMEDOUT)
				atomic.AddInt32(&hostSemaphore.Waiters, -1)
				return ERR_PTR
			}
		}
		atomic.AddInt32(&hostSemaphore.Waiters, -1)
	}
}

func libScePosix_sem_post(semaphore *PSemaphore) uintptr {
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_post"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	if semaphore.Magic != PSemaphoreMagic {
		logger.Printf("%-132s %s failed due to invalid sem magic.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_post"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Increment semaphore for fast-path.
	atomic.AddInt32(&semaphore.Value, 1)

	// Signal slow-path.
	hostSemaphore := GetPSemaphore(uintptr(unsafe.Pointer(semaphore)))
	hostSemaphore.Signal()
	if logger.LogSyncing {
		logger.Printf("%-132s %s signaled semaphore %s (value=%d).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_post"),
			color.Yellow.Sprintf("0x%X", uintptr(unsafe.Pointer(semaphore))),
			semaphore.Value,
		)
	}

	return 0
}
