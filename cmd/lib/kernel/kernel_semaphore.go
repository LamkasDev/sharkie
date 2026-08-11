package kernel

import (
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/semaphore"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000023410
// __int64 __fastcall sceKernelCreateSema(_QWORD *, __int64, unsigned int, unsigned int, unsigned int, __int64)
func libKernel_sceKernelCreateSema(handlePtr uintptr, namePtr Cstring, attributes uint32, currentCount, maxCount int32, optionPtr uintptr) uintptr {
	if handlePtr == 0 || attributes > 2 || currentCount < 0 || maxCount <= 0 || currentCount > maxCount {
		logger.Printf("%-132s %s failed due to invalid parameters.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCreateSema"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	semaphore := CreateSemaphore("unnamed", attributes, currentCount, maxCount)
	var name string
	if namePtr != nil {
		name = GoString(namePtr)
	}
	if name == "" {
		name = fmt.Sprintf("0x%X", semaphore.Handle)
	}
	semaphore.Name = name

	handleSlice := unsafe.Slice((*byte)(unsafe.Pointer(handlePtr)), 4)
	binary.LittleEndian.PutUint32(handleSlice, uint32(semaphore.Handle))

	logger.Printf("%-132s %s created semaphore %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelCreateSema"),
		color.Yellow.Sprintf("0x%X", semaphore.Handle),
		color.Blue.Sprint(semaphore.Name),
	)
	return 0
}

// 0x0000000000023580
// __int64 __fastcall sceKernelOpenSema(_QWORD *, __int64)
func libKernel_sceKernelOpenSema(handlePtr uintptr, namePtr Cstring) uintptr {
	if handlePtr == 0 || namePtr == nil {
		logger.Printf("%-132s %s failed due to handle or name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelOpenSema"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	name := GoString(namePtr)

	var foundSemaphore *Semaphore
	SemaphoreLock.RLock()
	for _, semaphore := range SemaphoreRepo {
		if semaphore.Name == name {
			foundSemaphore = semaphore
			break
		}
	}
	SemaphoreLock.RUnlock()

	if foundSemaphore == nil {
		logger.Printf("%-132s %s failed due to unknown semaphore %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelOpenSema"),
			color.Blue.Sprint(name),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	handleSlice := unsafe.Slice((*byte)(unsafe.Pointer(handlePtr)), 4)
	binary.LittleEndian.PutUint32(handleSlice, uint32(foundSemaphore.Handle))

	logger.Printf("%-132s %s opened semaphore %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelOpenSema"),
		color.Yellow.Sprintf("0x%X", foundSemaphore.Handle),
		color.Blue.Sprint(name),
	)
	return 0
}

// 0x0000000000023550
// __int64 sceKernelCancelSema()
func libKernel_sceKernelCancelSema(handle uint32, setCount int32, numWaitThreadsPtr uintptr) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCancelSema"),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	semaphore.Cond.Mutex.Lock()

	// Output the number of waiting threads.
	if numWaitThreadsPtr != 0 {
		waitersCount := atomic.LoadInt32(&semaphore.Cond.Waiters)
		numWaitThreadsSlice := unsafe.Slice((*byte)(unsafe.Pointer(numWaitThreadsPtr)), 4)
		binary.LittleEndian.PutUint32(numWaitThreadsSlice, uint32(waitersCount))
	}

	// Set the new token count (if < 0, reset to initial count).
	if setCount < 0 {
		semaphore.CurrentCount = semaphore.InitCount
	} else {
		semaphore.CurrentCount = setCount
	}

	// Cancel current waiters by bumping the generation and waking them up.
	semaphore.CancelGeneration++
	semaphore.Cond.Mutex.Unlock()
	semaphore.Cond.Broadcast()
	if logger.LogSyncing {
		logger.Printf("%-132s %s canceled semaphore %s (newCount=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCancelSema"),
			color.Blue.Sprint(semaphore.Name),
			color.Yellow.Sprintf("0x%X", semaphore.CurrentCount),
		)
	}

	return 0
}

// 0x0000000000023460
// __int64 sceKernelDeleteSema()
func libKernel_sceKernelDeleteSema(handle uint32) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelDeleteSema"),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	// Safely check for active waiters before deleting.
	if atomic.LoadInt32(&semaphore.Cond.Waiters) > 0 {
		logger.Printf("%-132s %s failed deleting semaphore %s (busy).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelDeleteSema"),
			color.Blue.Sprint(semaphore.Name),
		)
		return SCE_KERNEL_ERROR_EBUSY
	}
	DeleteSemaphore(handle)

	logger.Printf("%-132s %s deleted semaphore %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelDeleteSema"),
		color.Blue.Sprint(semaphore.Name),
	)
	return 0
}

// 0x0000000000023490
// __int64 __fastcall sceKernelWaitSema(unsigned int, unsigned int, __int64)
func libKernel_sceKernelWaitSema(handle uint32, needed int32, timeout *Timeout) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelWaitSema"),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	timeoutDuration := time.Duration(-1)
	if timeout != nil {
		timeoutDuration = time.Duration(timeout.Microseconds) * time.Microsecond
	}

	// Snapshot the cancel generation.
	semaphore.Cond.Mutex.Lock()
	startCancelGen := semaphore.CancelGeneration
	semaphore.Cond.Mutex.Unlock()

	start := time.Now()
	for {
		// Check if this semaphore wait was canceled.
		if semaphore.CancelGeneration != startCancelGen {
			semaphore.Cond.Mutex.Unlock()
			return 0x80020055
		}

		// Check value.
		semaphore.Cond.Mutex.Lock()
		if semaphore.CurrentCount >= needed {
			semaphore.CurrentCount -= needed
			if logger.LogSyncing {
				logger.Printf("%-132s %s decremented semaphore %s to %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceKernelWaitSema"),
					color.Blue.Sprint(semaphore.Name),
					color.Green.Sprint(semaphore.CurrentCount),
				)
			}
			semaphore.Cond.Mutex.Unlock()
			return 0
		}
		w := semaphore.Cond.WaitChanNoLock()
		atomic.AddInt32(&semaphore.Cond.Waiters, 1)
		semaphore.Cond.Mutex.Unlock()

		var remaining time.Duration
		if timeoutDuration != -1 {
			remaining = timeoutDuration - time.Since(start)
			if remaining <= 0 {
				if logger.LogSyncingFail {
					logger.Printf("%-132s %s timed out semaphore %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sceKernelWaitSema"),
						color.Blue.Sprint(semaphore.Name),
					)
				}
				atomic.AddInt32(&semaphore.Cond.Waiters, -1)
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}

		// Wait.
		if logger.LogSyncing {
			logger.Printf("%-132s %s waiting on semaphore %s for %s microseconds.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelWaitSema"),
				color.Blue.Sprint(semaphore.Name),
				color.Yellow.Sprintf("0x%X", timeoutDuration.Microseconds()),
			)
		}
		if timeoutDuration == -1 {
			<-w
		} else {
			select {
			case <-w:
			case <-time.After(remaining):
				if logger.LogSyncingFail {
					logger.Printf("%-132s %s timed out on semaphore %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sceKernelWaitSema"),
						color.Blue.Sprint(semaphore.Name),
					)
				}
				atomic.AddInt32(&semaphore.Cond.Waiters, -1)
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}
		atomic.AddInt32(&semaphore.Cond.Waiters, -1)
	}
}

// 0x00000000000234F0
// __int64 sceKernelPollSema()
func libKernel_sceKernelPollSema(handle uint32, needed int32) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelPollSema"),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	semaphore.Cond.Mutex.Lock()
	defer semaphore.Cond.Mutex.Unlock()

	if semaphore.CurrentCount >= needed {
		semaphore.CurrentCount -= needed
		if logger.LogSyncing {
			logger.Printf("%-132s %s decremented semaphore %s to %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelPollSema"),
				color.Blue.Sprint(semaphore.Name),
				color.Green.Sprint(semaphore.CurrentCount),
			)
		}
		return 0
	}

	return SCE_KERNEL_ERROR_EBUSY
}

// 0x0000000000023520
// __int64 sceKernelSignalSema()
func libKernel_sceKernelSignalSema(handle uint32, signalCount int32) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelSignalSema"),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	semaphore.Cond.Mutex.Lock()
	if semaphore.CurrentCount+signalCount > semaphore.MaxCount {
		semaphore.Cond.Mutex.Unlock()
		return SCE_KERNEL_ERROR_EINVAL
	}
	semaphore.CurrentCount += signalCount
	semaphore.Cond.Mutex.Unlock()

	semaphore.Cond.Broadcast()
	if logger.LogSyncing {
		logger.Printf("%-132s %s incremented semaphore %s to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelSignalSema"),
			color.Blue.Sprint(semaphore.Name),
			color.Green.Sprint(semaphore.CurrentCount),
		)
	}

	return 0
}
