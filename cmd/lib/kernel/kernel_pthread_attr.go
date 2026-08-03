package kernel

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000134A0
// __int64 __fastcall scePthreadAttrInit(__int64 *)
func libKernel_scePthreadAttrInit(attrHandlePtr *uintptr) uintptr {
	err := posix.Pthread_attr_init(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000133E0
// __int64 scePthreadAttrSetstacksize()
func libKernel_scePthreadAttrSetstacksize(attrHandlePtr *uintptr, stackSize uint64) uintptr {
	err := posix.Pthread_attr_setstacksize(attrHandlePtr, stackSize)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000143E0
// __int64 scePthreadAttrSetschedpolicy()
func libKernel_scePthreadAttrSetschedpolicy(attrHandlePtr *uintptr, schedulingPolicy PthreadSchedulingPolicy) uintptr {
	err := posix.Pthread_attr_setschedpolicy(attrHandlePtr, schedulingPolicy)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000143A0
// __int64 scePthreadAttrSetinheritsched()
func libKernel_scePthreadAttrSetinheritsched(attrHandlePtr *uintptr, inheritScheduling PthreadInheritScheduling) uintptr {
	err := posix.Pthread_attr_setinheritsched(attrHandlePtr, inheritScheduling)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014340
// __int64 scePthreadAttrGetschedparam()
func libKernel_scePthreadAttrGetschedparam(attrHandlePtr *uintptr, schedulingParameterPtr *int32) uintptr {
	err := posix.Pthread_attr_getschedparam(attrHandlePtr, schedulingParameterPtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000143C0
// __int64 scePthreadAttrSetschedparam()
func libKernel_scePthreadAttrSetschedparam(attrHandlePtr *uintptr, schedulingParameterPtr *int32) uintptr {
	err := posix.Pthread_attr_setschedparam(attrHandlePtr, schedulingParameterPtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000134E0
// __int64 scePthreadAttrSetguardsize()
func libKernel_scePthreadAttrSetguardsize(attrHandlePtr *uintptr, guardSize uint64) uintptr {
	err := posix.Pthread_attr_setguardsize(attrHandlePtr, guardSize)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013540
// __int64 scePthreadAttrSetdetachstate()
func libKernel_scePthreadAttrSetdetachstate(attrHandlePtr *uintptr, detachState uintptr) uintptr {
	err := posix.Pthread_attr_setdetachstate(attrHandlePtr, detachState)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014400
// __int64 scePthreadAttrSetscope()
func libKernel_scePthreadAttrSetscope(attrHandlePtr *uintptr, scope uintptr) uintptr {
	err := posix.Pthread_attr_setscope(attrHandlePtr, scope)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000133E0
// __int64 __fastcall scePthreadAttrDestroy(__int64 *)
func libKernel_scePthreadAttrDestroy(attrHandlePtr *uintptr) uintptr {
	err := posix.Pthread_attr_destroy(attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014480
// __int64 __fastcall scePthreadAttrGet(volatile signed __int32 *, __int64 *)
func libKernel_scePthreadAttrGet(threadPtr uintptr, attrHandlePtr *uintptr) uintptr {
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
	attr.StackSize = thread.Stack.Size
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
func libKernel_scePthreadAttrGetstack(attrHandlePtr *uintptr, addrPtr, sizePtr uintptr) uintptr {
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePthreadAttrGetstack"),
		)
		return err
	}
	if addrPtr != 0 {
		WriteAddress(addrPtr, attr.StackAddress)
	}
	if sizePtr != 0 {
		sizeSlice := unsafe.Slice((*byte)(unsafe.Pointer(sizePtr)), 8)
		binary.LittleEndian.PutUint64(sizeSlice, attr.StackSize)
	}

	logger.Printf("%-132s %s returned thread stack attributes (attrHandlePtr=%s, addrPtr=%s, sizePtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGetstack"),
		color.Yellow.Sprintf("0x%X", attrHandlePtr),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Yellow.Sprintf("0x%X", sizePtr),
	)
	return 0
}

// 0x0000000000013460
// __int64 scePthreadAttrGetstackaddr()
func libKernel_scePthreadAttrGetstackaddr(attrHandlePtr *uintptr, addrPtr uintptr) uintptr {
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePthreadAttrGetstackaddr"),
		)
		return err
	}
	if addrPtr != 0 {
		WriteAddress(addrPtr, attr.StackAddress)
	}

	logger.Printf("%-132s %s returned thread stack address (attrHandlePtr=%s, addrPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGetstackaddr"),
		color.Yellow.Sprintf("0x%X", attrHandlePtr),
		color.Yellow.Sprintf("0x%X", addrPtr),
	)
	return 0
}

// 0x0000000000013420
// __int64 scePthreadAttrGetstacksize()
func libKernel_scePthreadAttrGetstacksize(attrHandlePtr *uintptr, sizePtr uintptr) uintptr {
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePthreadAttrGetstacksize"),
		)
		return err
	}
	if sizePtr != 0 {
		sizeSlice := unsafe.Slice((*byte)(unsafe.Pointer(sizePtr)), 8)
		binary.LittleEndian.PutUint64(sizeSlice, attr.StackSize)
	}

	logger.Printf("%-132s %s returned thread stack size (attrHandlePtr=%s, sizePtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGetstacksize"),
		color.Yellow.Sprintf("0x%X", attrHandlePtr),
		color.Yellow.Sprintf("0x%X", sizePtr),
	)
	return 0
}

// 0x00000000000144A0
// __int64 __fastcall scePthreadAttrGetaffinity(__int64, _QWORD *)
func libKernel_scePthreadAttrGetaffinity(attrHandlePtr *uintptr, outMaskPtr uintptr) uintptr {
	var cpuSet [16]byte
	err := posix.Pthread_attr_getaffinity_np(attrHandlePtr, 16, uintptr(unsafe.Pointer(&cpuSet[0])))
	if err != 0 {
		return err - SonyErrorOffset
	}
	if outMaskPtr != 0 {
		outMask := unsafe.Slice((*byte)(unsafe.Pointer(outMaskPtr)), 8)
		binary.LittleEndian.PutUint64(outMask, *(*uint64)(unsafe.Pointer(&cpuSet[0])))
	}

	return 0
}
