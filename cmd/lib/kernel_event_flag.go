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
func libKernel_sceKernelCreateEventFlag(handlePtr uintptr, namePtr Cstring, attr, initPattern, optParamPtr uintptr) uintptr {
	// This is correct, btw.
	if handlePtr == 0 || optParamPtr != 0 {
		logger.Printf("%-132s %s failed due to invalid handle pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCreateEventFlag"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	handle := libKernel_sys_evf_create(namePtr, uint32(attr), uint64(initPattern))
	if handle == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	efHandleSlice := unsafe.Slice((*byte)(unsafe.Pointer(handlePtr)), 8)
	binary.LittleEndian.PutUint64(efHandleSlice, uint64(handle))

	return 0
}

func libKernel_sys_evf_create(namePtr Cstring, attr uint32, initPattern uint64) uintptr {
	name := "unnamed"
	if namePtr != nil {
		name = GoString(namePtr)
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
func libKernel_sceKernelOpenEventFlag(handlePtr uintptr, namePtr Cstring) uintptr {
	if handlePtr == 0 || namePtr == nil {
		logger.Printf("%-132s %s failed due to invalid handle pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelOpenEventFlag"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	name := GoString(namePtr)

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

	handleSlice := unsafe.Slice((*byte)(unsafe.Pointer(handlePtr)), 8)
	binary.LittleEndian.PutUint64(handleSlice, uint64(foundEventFlag.Handle))

	logger.Printf("%-132s %s opened event flag %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelOpenEventFlag"),
		color.Yellow.Sprintf("0x%X", foundEventFlag.Handle),
		color.Blue.Sprint(name),
	)
	return 0
}

// 0x0000000000023240
// __int64 __fastcall sceKernelWaitEventFlag(unsigned int, __int64, unsigned int, __int64, __int64)
func libKernel_sceKernelWaitEventFlag(handle uintptr, waitPattern uint64, waitMode uint32, outPatternPtr, timeoutPtr uintptr) uintptr {
	err := libKernel_sys_evf_wait(handle, waitPattern, waitMode, outPatternPtr, timeoutPtr)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

func libKernel_sys_evf_wait(handle uintptr, waitPattern uint64, waitMode uint32, outPatternPtr, timeoutPtr uintptr) uintptr {
	eventFlag := GetEventFlag(handle)
	if eventFlag == nil {
		logger.Printf("%-132s %s failed due to unknown event flag handle %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_wait"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	timeout := time.Duration(-1)
	if timeoutPtr != 0 {
		timeoutObj := (*Timeout)(unsafe.Pointer(timeoutPtr))
		timeout = time.Duration(timeoutObj.Microseconds) * time.Microsecond
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

			if logger.LogSyncing {
				logger.Printf("%-132s %s finished waiting on event flag %s.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sys_evf_wait"),
					GetEventFlagName(eventFlag),
				)
			}
			return 0
		}

		if timeout != -1 {
			if time.Since(start) >= timeout {
				if logger.LogSyncingFail {
					logger.Printf("%-132s %s timed out event flag %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sys_evf_wait"),
						GetEventFlagName(eventFlag),
					)
				}
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}

		// Wait.
		if logger.LogSyncing {
			logger.Printf("%-132s %s waiting on event flag %s for %s microseconds.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_evf_wait"),
				GetEventFlagName(eventFlag),
				color.Yellow.Sprintf("0x%X", timeout.Microseconds()),
			)
		}
		if timeout == -1 {
			eventFlag.Cond.Wait()
		} else {
			waited := CondWaitTimeout(eventFlag.Cond, timeout)
			if !waited {
				if logger.LogSyncingFail {
					logger.Printf("%-132s %s timed out on event flag %s.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sys_evf_wait"),
						GetEventFlagName(eventFlag),
					)
				}
				return SCE_KERNEL_ERROR_TIMEDOUT
			}
		}
	}
}

// 0x00000000000232B0
// __int64 sceKernelPollEventFlag()
func libKernel_sceKernelPollEventFlag(handle uintptr, waitPattern uint64, waitMode uint32, outPatternPtr uintptr) uintptr {
	err := libKernel_sys_evf_trywait(handle, waitPattern, waitMode, outPatternPtr)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

func libKernel_sys_evf_trywait(handle uintptr, waitPattern uint64, waitMode uint32, outPatternPtr uintptr) uintptr {
	eventFlag := GetEventFlag(handle)
	if eventFlag == nil {
		logger.Printf("%-132s %s failed due unknown event flag handle %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_trywait"),
			color.Yellow.Sprintf("0x%X", handle),
		)
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

		if logger.LogSyncing {
			logger.Printf("%-132s %s finished waiting on event flag %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sys_evf_trywait"),
				GetEventFlagName(eventFlag),
			)
		}
		return 0
	}

	if logger.LogSyncingFail {
		logger.Printf("%-132s %s tried waiting on event flag %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_trywait"),
			GetEventFlagName(eventFlag),
		)
	}
	return SCE_KERNEL_ERROR_TIMEDOUT
}

// 0x00000000000232E0
// __int64 sceKernelSetEventFlag()
func libKernel_sceKernelSetEventFlag(handle uintptr, bits uint64) uintptr {
	err := libKernel_sys_evf_set(handle, bits)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

func libKernel_sys_evf_set(handle uintptr, bits uint64) uintptr {
	eventFlag := GetEventFlag(handle)
	if eventFlag == nil {
		logger.Printf("%-132s %s failed due unknown event flag handle %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_set"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	eventFlag.Lock.Lock()
	eventFlag.CurrentPattern |= bits
	eventFlag.Lock.Unlock()
	eventFlag.Cond.Broadcast()

	logger.Printf("%-132s %s set event flag %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_evf_set"),
		GetEventFlagName(eventFlag),
		color.Yellow.Sprintf("0x%X", bits),
	)
	return 0
}
