package lib

import (
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000231C0
// __int64 __fastcall sceKernelCreateEventFlag(_QWORD *, __int64, unsigned int, __int64, __int64)
func libKernel_sceKernelCreateEventFlag(efHandlePtr, namePtr, attr, initPattern, optParamPtr uintptr) uintptr {
	// This is correct, btw.
	if efHandlePtr == 0 || optParamPtr != 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCreateEventFlag"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	handle := libKernel_sys_evf_create(namePtr, uint32(attr), uint64(initPattern))
	if handle == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	efHandleSlice := unsafe.Slice((*byte)(unsafe.Pointer(efHandlePtr)), 4)
	binary.LittleEndian.PutUint32(efHandleSlice, uint32(handle))

	return 0
}

func libKernel_sys_evf_create(namePtr uintptr, attr uint32, initPattern uint64) uintptr {
	name := "unnamed"
	if namePtr != 0 {
		name = ReadCString(namePtr)
	}
	if len(name) >= EVF_NAME_MAX {
		logger.Printf("%-132s %s failed due to too long name.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_create"),
		)
		SetErrno(ENAMETOOLONG)
		return ERR_PTR
	}

	queueMode := attr & 0xF
	if queueMode != EVF_ATTR_TH_FIFO && queueMode != 0 {
		logger.Printf("%-132s %s requesting unknown queue mode %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_create"),
			color.Yellow.Sprintf("0x%X", queueMode),
		)
	}

	eventFlag := CreateEventFlag(name, attr, initPattern, initPattern)

	logger.Printf("%-132s %s created an event flag %s (name=%s, attr=%s, initPattern=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_evf_create"),
		color.Yellow.Sprintf("0x%X", eventFlag.Handle),
		color.Blue.Sprint(name),
		color.Yellow.Sprintf("0x%X", attr),
		color.Yellow.Sprintf("0x%X", initPattern),
	)
	return eventFlag.Handle
}

// 0x0000000000023370
// __int64 __fastcall sceKernelOpenEventFlag(_QWORD *, __int64)
func libKernel_sceKernelOpenEventFlag(handlePtr uintptr, namePtr uintptr) uintptr {
	if handlePtr == 0 || namePtr == 0 {
		return SCE_KERNEL_ERROR_EINVAL
	}
	name := ReadCString(namePtr)

	var foundEventFlag *EventFlag
	EventFlagLock.RLock()
	for _, eventFlag := range EventFlagRepo {
		if eventFlag.Name == name {
			foundEventFlag = eventFlag
			break
		}
	}
	EventFlagLock.RUnlock()

	if foundEventFlag == nil {
		logger.Printf("%-132s %s failed due to unknown event flag %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelOpenEventFlag"),
			color.Blue.Sprint(name),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	handleSlice := unsafe.Slice((*byte)(unsafe.Pointer(handlePtr)), 4)
	binary.LittleEndian.PutUint32(handleSlice, uint32(foundEventFlag.Handle))

	logger.Printf("%-132s %s opened event flag %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelOpenEventFlag"),
		color.Yellow.Sprintf("0x%X", foundEventFlag.Handle),
		color.Green.Sprint(name),
	)
	return 0
}

// 0x0000000000023240
// __int64 __fastcall sceKernelWaitEventFlag(unsigned int, __int64, unsigned int, __int64, __int64)
func libKernel_sceKernelWaitEventFlag(handle, waitPattern, waitMode, outPatternPtr, timeoutPtr uintptr) uintptr {
	err := libKernel_sys_evf_wait(handle, uint64(waitPattern), uint32(waitMode), outPatternPtr, timeoutPtr)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

func libKernel_sys_evf_wait(handle uintptr, waitPattern uint64, waitMode uint32, outPatternPtr uintptr, timeoutPtr uintptr) uintptr {
	eventFlag := GetEventFlag(handle)
	if eventFlag == nil {
		return SCE_KERNEL_ERROR_ENOENT
	}

	timeout := time.Duration(-1)
	if timeoutPtr != 0 {
		timeoutSlice := unsafe.Slice((*byte)(unsafe.Pointer(timeoutPtr)), 4)
		micros := binary.LittleEndian.Uint32(timeoutSlice)
		timeout = time.Duration(micros) * time.Microsecond
	}

	eventFlag.Lock.Lock()
	defer eventFlag.Lock.Unlock()

	start := time.Now()
	for {
		// Check condition.
		if CheckEventFlagCondition(eventFlag.CurrentPattern, waitPattern, waitMode) {
			if outPatternPtr != 0 {
				outPatternSlice := unsafe.Slice((*byte)(unsafe.Pointer(outPatternPtr)), 8)
				binary.LittleEndian.PutUint64(outPatternSlice, eventFlag.CurrentPattern)
			}

			if (waitMode & EVF_WAITMODE_CLEAR_ALL) != 0 {
				eventFlag.CurrentPattern = 0
			}
			if (waitMode & EVF_WAITMODE_CLEAR_PAT) != 0 {
				eventFlag.CurrentPattern &= ^waitPattern
			}

			logger.Printf("%-132s %s finished waiting on event flag %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_evf_wait"),
				color.Blue.Sprint(eventFlag.Name),
			)
			return 0
		}

		if timeout != -1 {
			if time.Since(start) >= timeout {
				logger.Printf("%-132s %s timed out event flag %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sys_evf_wait"),
					color.Blue.Sprint(eventFlag.Name),
				)
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}

		// Wait.
		logger.Printf("%-132s %s waiting on event flag %s for %s microseconds.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_wait"),
			color.Blue.Sprint(eventFlag.Name),
			color.Yellow.Sprintf("0x%X", timeout.Microseconds()),
		)
		if timeout == -1 {
			eventFlag.Cond.Wait()
		} else {
			waited := CondWaitTimeout(eventFlag.Cond, timeout)
			if !waited {
				logger.Printf("%-132s %s timed out on event flag %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sys_evf_wait"),
					color.Blue.Sprint(eventFlag.Name),
				)
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}
	}
}

// 0x00000000000232B0
// __int64 sceKernelPollEventFlag()
func libKernel_sceKernelPollEventFlag(handle, waitPattern, waitMode, outPatternPtr uintptr) uintptr {
	err := libKernel_sys_evf_trywait(handle, uint64(waitPattern), uint32(waitMode), outPatternPtr)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

func libKernel_sys_evf_trywait(handle uintptr, waitPattern uint64, waitMode uint32, outPatternPtr uintptr) uintptr {
	eventFlag := GetEventFlag(handle)
	if eventFlag == nil {
		return SCE_KERNEL_ERROR_ENOENT
	}

	eventFlag.Lock.Lock()
	defer eventFlag.Lock.Unlock()

	// Check condition.
	if CheckEventFlagCondition(eventFlag.CurrentPattern, waitPattern, waitMode) {
		if outPatternPtr != 0 {
			outPatternSlice := unsafe.Slice((*byte)(unsafe.Pointer(outPatternPtr)), 8)
			binary.LittleEndian.PutUint64(outPatternSlice, eventFlag.CurrentPattern)
		}

		if (waitMode & EVF_WAITMODE_CLEAR_ALL) != 0 {
			eventFlag.CurrentPattern = 0
		}
		if (waitMode & EVF_WAITMODE_CLEAR_PAT) != 0 {
			eventFlag.CurrentPattern &= ^waitPattern
		}

		logger.Printf("%-132s %s finished waiting on event flag %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_trywait"),
			color.Blue.Sprint(eventFlag.Name),
		)
		return 0
	}

	logger.Printf("%-132s %s tried waiting on event flag %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_evf_trywait"),
		color.Blue.Sprint(eventFlag.Name),
	)
	return SCE_KERNEL_ERROR_TIMEDOUT
}

// 0x00000000000232E0
// __int64 sceKernelSetEventFlag()
func libKernel_sceKernelSetEventFlag(handle, bits uintptr) uintptr {
	err := libKernel_sys_evf_set(handle, uint64(bits))
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

func libKernel_sys_evf_set(handle uintptr, bits uint64) uintptr {
	eventFlag := GetEventFlag(handle)
	if eventFlag == nil {
		return SCE_KERNEL_ERROR_ENOENT
	}

	eventFlag.Lock.Lock()
	eventFlag.CurrentPattern |= bits
	eventFlag.Lock.Unlock()
	eventFlag.Cond.Broadcast()

	logger.Printf("%-132s %s set event flag %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_evf_set"),
		color.Blue.Sprint(eventFlag.Name),
		color.Yellow.Sprintf("0x%X", bits),
	)
	return 0
}
