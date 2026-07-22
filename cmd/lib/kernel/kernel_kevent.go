package kernel

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Kevent(equeueHandle, changelistPtr, nchanges, eventlistPtr, nevents, timestampPtr uintptr) uintptr {
	return libKernel_kevent(equeueHandle, changelistPtr, nchanges, eventlistPtr, nevents, timestampPtr)
}

// 0x00000000000013B0
// __int64 __fastcall kevent()
func libKernel_kevent(equeueHandle, changelistPtr, nchanges, eventlistPtr, nevents, timestampPtr uintptr) uintptr {
	equeue := GetEqueue(equeueHandle)
	if equeue == nil {
		logger.Printf("%-132s %s failed due to unknown equeue %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("kevent"),
			color.Yellow.Sprintf("0x%X", equeueHandle),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}

	if changelistPtr != 0 && nchanges > 0 {
		changes := unsafe.Slice((*KernelEvent)(unsafe.Pointer(changelistPtr)), nchanges)
		for _, event := range changes {
			processKeventChange(equeue, event)
		}
	}

	if eventlistPtr != 0 && nevents > 0 {
		return processKeventWait(equeue, eventlistPtr, nevents, timestampPtr)
	}

	return 0
}

func processKeventChange(equeue *Equeue, event KernelEvent) {
	if (event.Flags&EV_ADD) != 0 || (event.Flags&EV_ENABLE) != 0 {
		switch event.Filter {
		case EVFILT_VIDEO_OUT:
			logger.Printf("%-132s %s ignoring video out event registration for now.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("processKeventChange"),
			)
			return
		}
	}

	logger.Printf("%-132s %s ignored change %s (filter=%s, flags=%s, filterFlags=%s, filterData=%s, userData=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("processKeventChange"),
		color.Yellow.Sprintf("0x%X", event.Id),
		color.Yellow.Sprintf("0x%X", event.Filter),
		color.Yellow.Sprintf("0x%X", event.Flags),
		color.Yellow.Sprintf("0x%X", event.FilterFlags),
		color.Yellow.Sprintf("0x%X", event.FilterData),
		color.Yellow.Sprintf("0x%X", event.UserData),
	)
}

func processKeventWait(equeue *Equeue, eventlistPtr, nevents, timestampPtr uintptr) uintptr {
	timeout := time.Duration(-1)
	if timestampPtr != 0 {
		timestamp := (*Timestamp)(unsafe.Pointer(timestampPtr))
		timeout = time.Duration(timestamp.Seconds)*time.Second +
			time.Duration(timestamp.Nanoseconds)*time.Nanosecond
	}
	logger.Printf("%-132s %s waiting on %s for %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("processKeventWait"),
		color.Blue.Sprint(equeue.Name),
		color.Yellow.Sprint(timeout.String()),
	)

	eventSlice := unsafe.Slice((*KernelEvent)(unsafe.Pointer(eventlistPtr)), nevents)
	switch {
	case timeout == 0:
		// Non-blocking poll.
		select {
		case event := <-equeue.Events:
			eventSlice[0] = event
			logger.Printf("%-132s %s returned event %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("processKeventWait"),
				color.Yellow.Sprintf("0x%X", event.Id),
			)
			return 1
		default:
			logger.Printf("%-132s %s returned no event.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("processKeventWait"),
			)
			return 0
		}
	case timeout > 0:
		// Timeout wait.
		select {
		case event := <-equeue.Events:
			eventSlice[0] = event
			logger.Printf("%-132s %s returned event %s for %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("processKeventWait"),
				color.Yellow.Sprintf("0x%X", event.Id),
				color.Blue.Sprint(equeue.Name),
			)
			return 1
		case <-time.After(timeout):
			logger.Printf("%-132s %s timed out on %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("processKeventWait"),
				color.Blue.Sprint(equeue.Name),
			)
			return 0
		}
	default:
		// Infinite wait.
		event := <-equeue.Events
		eventSlice[0] = event
		logger.Printf("%-132s %s returned event %s for %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("processKeventWait"),
			color.Yellow.Sprintf("0x%X", event.Id),
			color.Blue.Sprint(equeue.Name),
		)
		return 1
	}
}
