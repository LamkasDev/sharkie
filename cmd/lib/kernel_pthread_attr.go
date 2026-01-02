package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000134A0
// __int64 __fastcall scePthreadAttrInit(__int64 *)
func libKernel_scePthreadAttrInit(attrHandlePtr uintptr) uintptr {
	err := libKernel_pthread_attr_init(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000133E0
// __int64 scePthreadAttrSetstacksize()
func libKernel_scePthreadAttrSetstacksize(attrHandlePtr uintptr, stackSize uintptr) uintptr {
	err := libKernel_pthread_attr_setstacksize(attrHandlePtr, stackSize)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000143E0
// __int64 scePthreadAttrSetschedpolicy()
func libKernel_scePthreadAttrSetschedpolicy(attrHandlePtr uintptr, schedulingPolicy uintptr) uintptr {
	err := libKernel_pthread_attr_setschedpolicy(attrHandlePtr, schedulingPolicy)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000143A0
// __int64 scePthreadAttrSetinheritsched()
func libKernel_scePthreadAttrSetinheritsched(attrHandlePtr uintptr, inheritScheduling uintptr) uintptr {
	err := libKernel_pthread_attr_setinheritsched(attrHandlePtr, inheritScheduling)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000143C0
// __int64 scePthreadAttrSetschedparam()
func libKernel_scePthreadAttrSetschedparam(attrHandlePtr uintptr, schedulingParameterPtr uintptr) uintptr {
	err := libKernel_pthread_attr_setschedparam(attrHandlePtr, schedulingParameterPtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000134E0
// __int64 scePthreadAttrSetguardsize()
func libKernel_scePthreadAttrSetguardsize(attrHandlePtr uintptr, guardSize uintptr) uintptr {
	err := libKernel_pthread_attr_setguardsize(attrHandlePtr, guardSize)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013540
// __int64 scePthreadAttrSetdetachstate()
func libKernel_scePthreadAttrSetdetachstate(attrHandlePtr uintptr, detachState uintptr) uintptr {
	err := libKernel_pthread_attr_setdetachstate(attrHandlePtr, detachState)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014400
// __int64 scePthreadAttrSetscope()
func libKernel_scePthreadAttrSetscope(attrHandlePtr uintptr, scope uintptr) uintptr {
	err := libKernel_pthread_attr_setscope(attrHandlePtr, scope)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000133E0
// __int64 __fastcall scePthreadAttrDestroy(__int64 *)
func libKernel_scePthreadAttrDestroy(attrHandlePtr uintptr) uintptr {
	err := libKernel_pthread_attr_destroy(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014480
// __int64 __fastcall scePthreadAttrGet(volatile signed __int32 *, __int64 *)
func libKernel_scePthreadAttrGet(threadPtr uintptr, attrHandlePtr uintptr) uintptr {
	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePthreadAttrGet"),
		)
		return err
	}

	thread := emu.GetCurrentThread()
	attr.StackAddress = thread.Stack.Address
	attr.StackSize = uintptr(len(thread.Stack.Contents))
	attr.GuardSize = GuardPageSize

	logger.Printf("%-132s %s assigned thread attributes (threadPtr=%s, attrHandlePtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGet"),
		color.Yellow.Sprintf("0x%X", threadPtr),
		color.Yellow.Sprintf("0x%X", attrHandlePtr),
	)
	return 0
}

// 0x0000000000013400
// __int64 scePthreadAttrGetstack()
func libKernel_scePthreadAttrGetstack(attrPtr uintptr, addrPtr uintptr, sizePtr uintptr) uintptr {
	thread := emu.GetCurrentThread()
	if addrPtr != 0 {
		WriteAddress(addrPtr, thread.Stack.Address)
	}

	if sizePtr != 0 {
		sizeSlice := unsafe.Slice((*byte)(unsafe.Pointer(sizePtr)), 8)
		binary.LittleEndian.PutUint64(sizeSlice, uint64(StackDefaultSize))
	}

	logger.Printf("%-132s %s returned thread attributes (attrPtr=%s, addrPtr=%s, sizePtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGetstack"),
		color.Yellow.Sprintf("0x%X", attrPtr),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Yellow.Sprintf("0x%X", sizePtr),
	)
	return 0
}

// 0x00000000000144A0
// __int64 __fastcall scePthreadAttrGetaffinity(__int64, _QWORD *)
func libKernel_scePthreadAttrGetaffinity(attrPtr uintptr, outMaskPtr uintptr) uintptr {
	var cpuSet [16]byte
	err := libKernel_pthread_attr_getaffinity_np(
		attrPtr,
		16,
		uintptr(unsafe.Pointer(&cpuSet[0])),
	)
	if err != 0 {
		return err - SonyErrorOffset
	}

	if outMaskPtr != 0 {
		outMask := unsafe.Slice((*byte)(unsafe.Pointer(outMaskPtr)), 8)
		binary.LittleEndian.PutUint64(outMask, *(*uint64)(unsafe.Pointer(&cpuSet[0])))
	}

	return 0
}

// 0x0000000000003F60
// __int64 __fastcall pthread_attr_getaffinity_np(__int64 *, unsigned __int64, __int64)
func libKernel_pthread_attr_getaffinity_np(attrPtr uintptr, cpuSetSize uintptr, cpuSetPtr uintptr) uintptr {
	if cpuSetPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid cpu set pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePthreadAttrGetaffinity"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	// We enable all cores in the mask for now.
	if cpuSetSize > 1024 {
		cpuSetSize = 1024
	}
	cpuSet := unsafe.Slice((*byte)(unsafe.Pointer(cpuSetPtr)), cpuSetSize)
	for i := range cpuSet {
		cpuSet[i] = 0xFF
	}

	logger.Printf("%-132s %s returned thread affinity (attrPtr=%s, cpuSetSize=%s, cpuSetPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGetaffinity"),
		color.Yellow.Sprintf("0x%X", attrPtr),
		color.Yellow.Sprintf("0x%X", cpuSetSize),
		color.Yellow.Sprintf("0x%X", cpuSetPtr),
	)
	return 0
}
