package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000001B470
// __int64 __fastcall sceKernelAddUserEvent(__m128 _XMM0)
func libKernel_sceKernelAddUserEvent(handle, eventId uintptr) uintptr {
	equeue := GetEqueue(handle)
	if equeue == nil {
		logger.Printf("%-132s %s failed due to unknown equeue %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelAddUserEvent"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}
	equeue.Lock.Lock()
	defer equeue.Lock.Unlock()
	equeue.UserEvents[eventId] = true

	logger.Printf("%-132s %s registered user event %s on %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelAddUserEvent"),
		color.Yellow.Sprintf("0x%X", eventId),
		color.Blue.Sprint(equeue.Name),
	)
	return 0
}

// 0x000000000001B4F0
// __int64 __fastcall sceKernelAddUserEventEdge(__int64, int, __m128 _XMM0)
func libKernel_sceKernelAddUserEventEdge(handle, eventId uintptr) uintptr {
	equeue := GetEqueue(handle)
	if equeue == nil {
		logger.Printf("%-132s %s failed due to unknown equeue %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelAddUserEventEdge"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}
	equeue.Lock.Lock()
	defer equeue.Lock.Unlock()
	equeue.UserEvents[eventId] = true

	logger.Printf("%-132s %s registered user event %s on %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelAddUserEventEdge"),
		color.Yellow.Sprintf("0x%X", eventId),
		color.Blue.Sprint(equeue.Name),
	)
	return 0
}

// 0x000000000001B5F0
// __int64 __fastcall sceKernelTriggerUserEvent(__int64, int, __int64)
func libKernel_sceKernelTriggerUserEvent(handle, eventId, userData uintptr) uintptr {
	equeue := GetEqueue(handle)
	if equeue == nil {
		logger.Printf("%-132s %s failed due to unknown equeue %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelTriggerUserEvent"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}
	kevent := KernelEvent{
		Id:       uint64(eventId),
		Filter:   KernelEventFilterUser,
		UserData: userData,
	}
	select {
	case equeue.Events <- kevent:
	default:
	}

	logger.Printf("%-132s %s triggered user event %s on %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelTriggerUserEvent"),
		color.Yellow.Sprintf("0x%X", eventId),
		color.Blue.Sprint(equeue.Name),
	)
	return 0
}

// 0x000000000001ABB0
// __int64 __fastcall sceKernelGetEventId(__int64)
func libKernel_sceKernelGetEventId(event *KernelEvent) uintptr {
	if logger.LogGraphics {
		logger.Printf("%-132s %s returned event id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetEventId"),
		)
	}
	return uintptr(event.Id)
}

// 0x000000000001ABA0
// __int64 __fastcall sceKernelGetEventFilter(__int64)
func libKernel_sceKernelGetEventFilter(event *KernelEvent) uintptr {
	if logger.LogGraphics {
		logger.Printf("%-132s %s returned event filter.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetEventFilter"),
		)
	}
	return uintptr(event.Filter)
}

// 0x000000000001ABC0
// __int64 __fastcall sceKernelGetEventData(__int64)
func libKernel_sceKernelGetEventData(event *KernelEvent) uintptr {
	if logger.LogGraphics {
		logger.Printf("%-132s %s returned event data.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetEventData"),
		)
	}
	return uintptr(event.FilterData)
}

// 0x000000000001ABF0
// __int64 __fastcall sceKernelGetEventUserData(__int64)
func libKernel_sceKernelGetEventUserData(event *KernelEvent) uintptr {
	if logger.LogGraphics {
		logger.Printf("%-132s %s returned event user data.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetEventUserData"),
		)
	}
	return uintptr(event.UserData)
}
