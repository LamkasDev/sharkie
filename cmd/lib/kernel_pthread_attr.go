package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000134A0
// __int64 __fastcall scePthreadAttrInit(__int64 *)
func libKernel_scePthreadAttrInit(attrHandlePtr uintptr) uintptr {
	err := libKernel_pthread_attr_init(attrHandlePtr)
	if err != 0 {
		return uintptr(uint32(err) - 0x7FFE0000)
	}

	return 0
}

// 0x0000000000014480
// __int64 __fastcall scePthreadAttrGet(volatile signed __int32 *, __int64 *)
func libKernel_scePthreadAttrGet(attrPtr uintptr, addrPtr uintptr, sizePtr uintptr) uintptr {
	if addrPtr != 0 {
		addrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
		binary.LittleEndian.PutUint64(addrSlice, uint64(emu.GlobalModuleManager.Stack.Address))
	}

	if sizePtr != 0 {
		sizeSlice := unsafe.Slice((*byte)(unsafe.Pointer(sizePtr)), 8)
		binary.LittleEndian.PutUint64(sizeSlice, uint64(StackDefaultSize))
	}

	fmt.Printf("%-120s %s returned thread attributes (attrPtr=%s, addrPtr=%s, sizePtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGet"),
		color.Yellow.Sprintf("0x%X", attrPtr),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Yellow.Sprintf("0x%X", sizePtr),
	)
	return 0
}

// 0x00000000000144A0
// __int64 __fastcall scePthreadAttrGetaffinity(__int64, _QWORD *)
func libKernel_scePthreadAttrGetaffinity(attrPtr uintptr, cpuSetSize uintptr, cpuSetPtr uintptr) uintptr {
	if cpuSetPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid cpu set pointer.\n",
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

	fmt.Printf("%-120s %s returned thread affinity (attrPtr=%s, cpuSetSize=%s, cpuSetPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadAttrGetaffinity"),
		color.Yellow.Sprintf("0x%X", attrPtr),
		color.Yellow.Sprintf("0x%X", cpuSetSize),
		color.Yellow.Sprintf("0x%X", cpuSetPtr),
	)
	return 0
}
