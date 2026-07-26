package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000C090
// __int64 __fastcall sigprocmask(unsigned int, _QWORD *, __int64)
func libKernel_sigprocmask(op uintptr, mask, oldMask *ThreadSignalMask) uintptr {
	thread := emu.GetCurrentThread()

	// Write back old mask.
	if oldMask != nil {
		oldMask.Low = thread.SignalMask.Low
		oldMask.High = thread.SignalMask.High
	}
	if mask == nil {
		return 0
	}

	// Read new mask.
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
