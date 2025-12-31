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
		return GetErrno() - 0x7FFE0000
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

	handleSlice := unsafe.Slice((*byte)(unsafe.Pointer(handlePtr)), 8)
	binary.LittleEndian.PutUint64(handleSlice, uint64(equeue.Handle))

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
func libKernel_sceKernelWaitEqueue(handle, eventPtr, num, resultPtr, microsPtr uintptr) uintptr {
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

	micros := uint32(0)
	timeoutSlice := [2]int64{}
	timeoutPtr := uintptr(unsafe.Pointer(&timeoutSlice[0]))
	if microsPtr != 0 {
		microsSlice := unsafe.Slice((*byte)(unsafe.Pointer(microsPtr)), 4)
		micros = binary.LittleEndian.Uint32(microsSlice)

		timeoutSlice[0] = int64(micros / 1_000_000)
		timeoutSlice[1] = int64((micros % 1_000_000) * 1000)
	}

	err := processKeventWait(equeue, eventPtr, num, timeoutPtr)
	if resultPtr != 0 {
		resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 8)
		binary.LittleEndian.PutUint64(resultSlice, uint64(err))
	}
	if err == 0 && microsPtr != 0 {
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
	equeue.UserEvents[eventId] = true
	equeue.Lock.Unlock()

	logger.Printf("%-132s %s added user event %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelAddUserEvent"),
		color.Yellow.Sprintf("0x%X", eventId),
		color.Blue.Sprint(equeue.Name),
	)
	return 0
}
