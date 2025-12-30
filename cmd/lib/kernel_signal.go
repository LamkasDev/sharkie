package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000000C090
// __int64 __fastcall sigprocmask(unsigned int, _QWORD *, __int64)
func libKernel_sigprocmask(op uintptr, maskPtr uintptr, oldMaskPtr uintptr) uintptr {
	thread := emu.GetCurrentThread()

	// Write back old mask.
	if oldMaskPtr != 0 {
		oldSetSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldMaskPtr)), 16)
		binary.LittleEndian.PutUint64(oldSetSlice, thread.SignalMask[0])
		binary.LittleEndian.PutUint64(oldSetSlice[8:], thread.SignalMask[1])
	}
	if maskPtr == 0 {
		return 0
	}

	// Read new mask.
	maskSlice := unsafe.Slice((*byte)(unsafe.Pointer(maskPtr)), 16)
	maskLow := binary.LittleEndian.Uint64(maskSlice)
	maskHigh := binary.LittleEndian.Uint64(maskSlice[8:])
	if op != SIG_UNBLOCK {
		maskLow &^= 0x80000000
	}

	// Perform specified operation and save it.
	thread.Lock.Lock()
	switch op {
	case SIG_BLOCK:
		thread.SignalMask[0] |= maskLow
		thread.SignalMask[1] |= maskHigh
	case SIG_UNBLOCK:
		thread.SignalMask[0] &^= maskLow
		thread.SignalMask[1] &^= maskHigh
	case SIG_SETMASK:
		thread.SignalMask[0] = maskLow
		thread.SignalMask[1] = maskHigh
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
