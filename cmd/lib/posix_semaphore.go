package lib

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000000FC70
// __int64 __fastcall sem_init(__int64, int, int)
func libKernel_sem_init(semPtr uintptr, pShared uintptr, value uintptr) uintptr {
	if semPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_init"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	// Initialize to defaults.
	semaphore := (*PSemaphore)(unsafe.Pointer(semPtr))
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
		color.Yellow.Sprintf("0x%X", semPtr),
		color.Yellow.Sprintf("0x%X", value),
	)
	return 0
}

// 0x00000000000109F0
// __int64 __fastcall sem_wait(__int64)
func libKernel_sem_wait(semPtr uintptr) uintptr {
	return libKernel_sem_timedwait(semPtr, 0)
}

// 0x00000000000104D0
// __int64 __fastcall sem_timedwait(__int64, __int64)
func libKernel_sem_timedwait(semPtr uintptr, timestampPtr uintptr) uintptr {
	if semPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_timedwait"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	semaphore := (*PSemaphore)(unsafe.Pointer(semPtr))
	if semaphore.Magic != PSemaphoreMagic {
		logger.Printf("%-132s %s failed due to invalid sem magic.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_timedwait"),
		)
		SetErrno(EINVAL)
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
					color.Yellow.Sprintf("0x%X", semPtr),
				)
			}
			return 0
		}
	}

	// Calculate actual timeout from absolute time.
	timeout := time.Duration(-1)
	if timestampPtr != 0 {
		timestamp := (*Timestamp)(unsafe.Pointer(timestampPtr))
		targetTime := time.Unix(int64(timestamp.Seconds), int64(timestamp.Nanoseconds))
		timeout = time.Until(targetTime)
		if timeout <= 0 {
			if logger.LogSyncingFail {
				logger.Printf("%-132s %s timed out on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sem_timedwait"),
					color.Yellow.Sprintf("0x%X", semPtr),
				)
			}
			SetErrno(ETIMEDOUT)
			return ERR_PTR
		}
	}

	// Lock semaphore.
	hostSemaphore := GetPSemaphore(semPtr)
	hostSemaphore.L.Lock()
	defer hostSemaphore.L.Unlock()

	for {
		// Check value again (holding lock this time).
		if semaphore.Value > 0 {
			semaphore.Value--
			if logger.LogSyncing {
				logger.Printf("%-132s %s waited on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sem_timedwait"),
					color.Yellow.Sprintf("0x%X", semPtr),
				)
			}
			return 0
		}

		// Wait.
		if logger.LogSyncing {
			logger.Printf("%-132s %s waiting on semaphore %s for %s microseconds.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sem_timedwait"),
				color.Yellow.Sprintf("0x%X", semPtr),
				color.Green.Sprintf("%d", timeout.Microseconds()),
			)
		}
		if timeout == -1 {
			hostSemaphore.Wait()
		} else {
			waited := CondWaitTimeout(hostSemaphore, timeout)
			if !waited {
				if logger.LogSyncingFail {
					logger.Printf("%-132s %s timed out on semaphore %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sem_timedwait"),
						color.Yellow.Sprintf("0x%X", semPtr),
					)
				}
				SetErrno(ETIMEDOUT)
				return ERR_PTR
			}
		}
	}
}

// 0x0000000000010A00
// __int64 __fastcall sem_post(__int64)
func libKernel_sem_post(semPtr uintptr) uintptr {
	if semPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid sem pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_post"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	semaphore := (*PSemaphore)(unsafe.Pointer(semPtr))
	if semaphore.Magic != PSemaphoreMagic {
		logger.Printf("%-132s %s failed due to invalid sem magic.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_post"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	// Increment semaphore for fast-path.
	atomic.AddInt32(&semaphore.Value, 1)

	// Signal slow-path.
	hostSemaphore := GetPSemaphore(semPtr)
	hostSemaphore.L.Lock()
	hostSemaphore.Signal()
	if logger.LogSyncing {
		logger.Printf("%-132s %s signaled semaphore %s (value=%d).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sem_post"),
			color.Yellow.Sprintf("0x%X", semPtr),
			semaphore.Value,
		)
	}
	hostSemaphore.L.Unlock()

	return 0
}
