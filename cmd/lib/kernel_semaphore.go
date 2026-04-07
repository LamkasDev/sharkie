package lib

import (
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000023410
// __int64 __fastcall sceKernelCreateSema(_QWORD *, __int64, unsigned int, unsigned int, unsigned int, __int64)
func libKernel_sceKernelCreateSema(handlePtr uintptr, namePtr Cstring, attributes uint32, currentCount, maxCount int32, optionPtr uintptr) uintptr {
	if handlePtr == 0 || optionPtr != 0 {
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

// 0x0000000000023460
// __int64 sceKernelDeleteSema()
func libKernel_sceKernelDeleteSema(handle uintptr) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		return SCE_KERNEL_ERROR_ENOENT
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
func libKernel_sceKernelWaitSema(handle uintptr, needed int32, timeoutPtr uintptr) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		return SCE_KERNEL_ERROR_ENOENT
	}

	timeout := time.Duration(-1)
	if timeoutPtr != 0 {
		timeoutObj := (*Timeout)(unsafe.Pointer(timeoutPtr))
		timeout = time.Duration(timeoutObj.Microseconds) * time.Microsecond
	}

	semaphore.Lock.Lock()
	defer semaphore.Lock.Unlock()

	start := time.Now()
	for {
		// Check value.
		if semaphore.CurrentCount >= needed {
			semaphore.CurrentCount -= needed
			logger.Printf("%-132s %s decremented semaphore %s to %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelWaitSema"),
				color.Blue.Sprint(semaphore.Name),
				color.Green.Sprint(semaphore.CurrentCount),
			)
			return 0
		}

		if timeout != -1 {
			if time.Since(start) >= timeout {
				logger.Printf("%-132s %s timed out semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceKernelWaitSema"),
					color.Blue.Sprint(semaphore.Name),
				)
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}

		// Wait.
		logger.Printf("%-132s %s waiting on semaphore %s for %s microseconds.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelWaitSema"),
			color.Blue.Sprint(semaphore.Name),
			color.Yellow.Sprintf("0x%X", timeout.Microseconds()),
		)
		if timeout == -1 {
			semaphore.Cond.Wait()
		} else {
			waited := CondWaitTimeout(semaphore.Cond, timeout)
			if !waited {
				logger.Printf("%-132s %s timed out on semaphore %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceKernelWaitSema"),
					color.Blue.Sprint(semaphore.Name),
				)
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}
	}
}

// 0x00000000000234F0
// __int64 sceKernelPollSema()
func libKernel_sceKernelPollSema(handle uintptr, needed int32) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		return SCE_KERNEL_ERROR_ENOENT
	}

	semaphore.Lock.Lock()
	defer semaphore.Lock.Unlock()

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
func libKernel_sceKernelSignalSema(handle uintptr, signalCount int32) uintptr {
	semaphore := GetSemaphore(handle)
	if semaphore == nil {
		return SCE_KERNEL_ERROR_ENOENT
	}

	semaphore.Lock.Lock()
	defer semaphore.Lock.Unlock()

	if semaphore.CurrentCount+signalCount > semaphore.MaxCount {
		return SCE_KERNEL_ERROR_EINVAL
	}

	semaphore.CurrentCount += signalCount
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
