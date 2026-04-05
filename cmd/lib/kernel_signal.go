package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000000C090
// __int64 __fastcall sigprocmask(unsigned int, _QWORD *, __int64)
func libKernel_sigprocmask(op, maskPtr, oldMaskPtr uintptr) uintptr {
	thread := emu.GetCurrentThread()

	// Write back old mask.
	if oldMaskPtr != 0 {
		oldMask := (*ThreadSignalMask)(unsafe.Pointer(oldMaskPtr))
		oldMask.Low = thread.SignalMask.Low
		oldMask.High = thread.SignalMask.High
	}
	if maskPtr == 0 {
		return 0
	}

	// Read new mask.
	mask := (*ThreadSignalMask)(unsafe.Pointer(maskPtr))
	maskLow := mask.Low
	maskHigh := mask.High
	if op != SIG_UNBLOCK {
		mask.Low &^= 0x80000000
	}

	// Perform specified operation and save it.
	thread.Lock.Lock()
	switch op {
	case SIG_BLOCK:
		thread.SignalMask.Low |= maskLow
		thread.SignalMask.High |= maskHigh
	case SIG_UNBLOCK:
		thread.SignalMask.Low &^= maskLow
		thread.SignalMask.High &^= maskHigh
	case SIG_SETMASK:
		thread.SignalMask.Low = maskLow
		thread.SignalMask.High = maskHigh
	default:
		thread.Lock.Unlock()
		logger.Printf("%-132s %s failed due to invalid op %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sigprocmask"),
			color.Yellow.Sprintf("0x%X", op),
		)
		return EINVAL
	}
	thread.Lock.Unlock()

	logger.Printf("%-132s %s set mask to %s %s (op=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sigprocmask"),
		color.Yellow.Sprintf("0x%X", maskLow),
		color.Yellow.Sprintf("0x%X", maskHigh),
		color.Yellow.Sprintf("0x%X", op),
	)
	return 0
}
