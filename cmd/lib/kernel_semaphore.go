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
func libKernel_sceKernelCreateSema(handlePtr, namePtr, attributes, currentCount, maxCount, optionPtr uintptr) uintptr {
	if handlePtr == 0 || optionPtr != 0 {
		return SCE_KERNEL_ERROR_EINVAL
	}

	semaphore := CreateSemaphore("unnamed", uint32(attributes), int32(currentCount), int32(maxCount))
	if namePtr != 0 {
		semaphore.Name = ReadCString(namePtr)
	} else {
		semaphore.Name = fmt.Sprintf("0x%X", semaphore.Handle)
	}

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
func libKernel_sceKernelOpenSema(handlePtr uintptr, namePtr uintptr) uintptr {
	if handlePtr == 0 || namePtr == 0 {
		return SCE_KERNEL_ERROR_EINVAL
	}
	name := ReadCString(namePtr)

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
		color.Green.Sprint(name),
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
func libKernel_sceKernelWaitSema(handle uintptr, needed uintptr, timeoutPtr uintptr) uintptr {
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
		if semaphore.CurrentCount >= int32(needed) {
			semaphore.CurrentCount--
			logger.Printf("%-132s %s decremented semaphore %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelWaitSema"),
				color.Blue.Sprint(semaphore.Name),
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
