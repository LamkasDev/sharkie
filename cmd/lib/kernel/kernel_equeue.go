package kernel

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000001AC00
// __int64 __fastcall sceKernelCreateEqueue(__int64 *, __int64)
func libKernel_sceKernelCreateEqueue(handlePtr *uintptr, namePtr Cstring) uintptr {
	if handlePtr == nil || namePtr == nil {
		logger.Printf("%-132s %s failed due to invalid handle or name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCreateEqueue"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	handle := posix.Kqueue()
	if handle == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
	}
	equeue := GetEqueue(handle)
	equeue.Name = fmt.Sprintf("0x%X", equeue.Handle)
	*handlePtr = handle

	// TODO: emulate __sys_namedobj_create?

	return 0
}

// 0x000000000001ACF0
// __int64 __fastcall sceKernelWaitEqueue(unsigned int, __int64, unsigned int, int *, unsigned int *)
func libKernel_sceKernelWaitEqueue(handle, eventPtr, num uintptr, resultPtr *int32, timeout *Timeout) uintptr {
	equeue := GetEqueue(handle)
	if equeue == nil {
		logger.Printf("%-132s %s failed due to unknown equeue %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelWaitEqueue"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}

	var timestamp *Timestamp
	if timeout != nil {
		timestamp = &Timestamp{
			Seconds:     int64(timeout.Microseconds / 1_000_000),
			Nanoseconds: int64((timeout.Microseconds % 1_000_000) * 1000),
		}
	}

	count := posix.ProcessKeventWait(equeue, eventPtr, num, timestamp)
	if resultPtr != nil {
		*resultPtr = int32(count)
	}
	if count == 0 {
		return SCE_KERNEL_ERROR_TIMEDOUT
	}

	return 0
}
