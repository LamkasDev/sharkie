package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000001AC00
// __int64 __fastcall sceKernelCreateEqueue(__int64 *, __int64)
func libKernel_sceKernelCreateEqueue(handlePtr uintptr, namePtr uintptr) uintptr {
	if namePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid handle pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCreateEqueue"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	err := libKernel_kqueue(handlePtr, namePtr)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	// TODO: emulate __sys_namedobj_create?

	return 0
}

// 0x0000000000000EB0
// __int64 __fastcall _sys_kqueueex()
func libKernel___sys_kqueueex(knlistPtr uintptr, count uintptr, flags uintptr) uintptr {
	var handlePtr uintptr
	libKernel_kqueue((uintptr)(unsafe.Pointer(&handlePtr)), 0)

	return handlePtr
}

// 0x0000000000001390
// __int64 __fastcall kqueue()
func libKernel_kqueue(handlePtr uintptr, namePtr uintptr) uintptr {
	if handlePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid handle pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("kqueue"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	equeue := CreateEqueue("unnamed")
	if namePtr != 0 {
		equeue.Name = ReadCString(namePtr)
	} else {
		equeue.Name = fmt.Sprintf("0x%X", equeue.Handle)
	}
	WriteAddress(handlePtr, equeue.Handle)

	logger.Printf("%-132s %s created equeue %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("kqueue"),
		color.Yellow.Sprintf("0x%X", equeue.Handle),
		color.Blue.Sprint(equeue.Name),
	)
	return 0
}

// 0x000000000001ACF0
// __int64 __fastcall sceKernelWaitEqueue(unsigned int, __int64, unsigned int, int *, unsigned int *)
func libKernel_sceKernelWaitEqueue(handle, eventPtr, num, resultPtr, timeoutPtr uintptr) uintptr {
	equeue := GetEqueue(handle)
	if equeue == nil {
		logger.Printf("%-132s %s failed due to unknown equeue %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelWaitEqueue"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	timestampPtr := uintptr(0)
	if timeoutPtr != 0 {
		timeout := (*Timeout)(unsafe.Pointer(timeoutPtr))
		timestamp := Timestamp{
			Seconds:     uint64(timeout.Microseconds / 1_000_000),
			Nanoseconds: uint64((timeout.Microseconds % 1_000_000) * 1000),
		}
		timestampPtr = uintptr(unsafe.Pointer(&timestamp))
	}

	count := processKeventWait(equeue, eventPtr, num, timestampPtr)
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
		SetErrno(EFAULT)
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
