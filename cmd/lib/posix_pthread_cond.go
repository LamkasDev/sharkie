package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

func libKernel_initStaticCond(condHandlePtr uintptr) uintptr {
	condAddr, _ := sys_struct.AllocReadWriteMemory(unsafe.Sizeof(PthreadCond{}))
	if condAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))
	cond.KernelId = 0
	cond.Flags = 0

	// Copy the pointer back to condHandlePtr.
	condHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(condHandlePtr))), 8)
	binary.LittleEndian.PutUint64(condHandlePtrSlice, uint64(condAddr))
	fmt.Printf("%-120s %s created structs at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("libKernel_initStaticCond"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)

	return 0
}

// 0x0000000000006150
// __int64 __fastcall pthread_cond_broadcast(__int64 *, __int64, int, int, int, int)
func libKernel_pthread_cond_broadcast(condHandlePtr uintptr) uintptr {
	if condHandlePtr == 0 {
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(uintptr(condHandlePtr)))
	if condAddr == PthreadCondInitializer {
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			return err
		}
		condAddr = *(*uintptr)(unsafe.Pointer(uintptr(condHandlePtr)))
	}

	// Broadcast to it.
	hostCond := GetCond(condAddr)
	hostCond.Broadcast()
	fmt.Printf("%-120s %s broadcasted cond %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_broadcast"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)

	return 0
}
