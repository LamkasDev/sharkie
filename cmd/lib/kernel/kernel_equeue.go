package kernel

import (
	"encoding/binary"
	"unsafe"

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
	if handlePtr == nil {
		logger.Printf("%-132s %s failed due to invalid handle pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCreateEqueue"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	err := posix.Kqueue(handlePtr, namePtr)
	if err == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
	}

	// TODO: emulate __sys_namedobj_create?

	return 0
}

// 0x000000000001ACF0
// __int64 __fastcall sceKernelWaitEqueue(unsigned int, __int64, unsigned int, int *, unsigned int *)
func libKernel_sceKernelWaitEqueue(handle, eventPtr, num, resultPtr uintptr, timeout *Timeout) uintptr {
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
	if resultPtr != 0 {
		resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
		binary.LittleEndian.PutUint32(resultSlice, uint32(count))
	}
	if count == 0 {
		return SCE_KERNEL_ERROR_TIMEDOUT
	}

	return 0
}

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
