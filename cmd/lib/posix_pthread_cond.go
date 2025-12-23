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

// 0x0000000000004CA0
// __int64 __fastcall pthread_cond_init(_QWORD *, _DWORD **)
func libKernel_pthread_cond_init(condHandlePtr, attrHandlePtr uintptr) uintptr {
	condAddr, _ := sys_struct.AllocReadWriteMemory(unsafe.Sizeof(PthreadCond{}))
	if condAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))
	cond.KernelId = 0
	cond.Flags = 0

	// Copy the pointer back to condHandlePtr.
	condHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(condHandlePtr)), 8)
	binary.LittleEndian.PutUint64(condHandlePtrSlice, uint64(condAddr))
	fmt.Printf("%-120s %s created struct at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_init"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)

	return 0
}

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
	fmt.Printf("%-120s %s created struct at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("libKernel_initStaticCond"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)

	return 0
}

// 0x0000000000004D60
// __int64 __fastcall pthread_cond_destroy(__int64 *)
func libKernel_pthread_cond_destroy(condHandlePtr uintptr) uintptr {
	if condHandlePtr == 0 {
		return EINVAL
	}

	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		return 0
	}

	// TODO: actually destroy it.

	fmt.Printf("%-120s %s destroyed cond %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_destroy"),
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
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			return err
		}
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
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

// 0x0000000000005AD0
// __int64 __fastcall pthread_cond_signal(__int64 *, __m128, __m128, __m128, __m128, __m128, __m128, __m128, __m128, __int64, __int64, __int64, __int64, __int64)
func libKernel_pthread_cond_signal(condHandlePtr uintptr) uintptr {
	if condHandlePtr == 0 {
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			return err
		}
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
	}

	// Signal to it.
	hostCond := GetCond(condAddr)
	hostCond.Signal()

	fmt.Printf("%-120s %s signaled cond %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_signal"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)
	return 0
}

// 0x0000000000005550
// __int64 __fastcall pthread_cond_wait(__int64 *, unsigned __int64 *, __int64, __int64, __int64, int)
func libKernel_pthread_cond_wait(condHandlePtr uintptr) uintptr {
	if condHandlePtr == 0 {
		return EINVAL
	}

	// Try initializing a cond, if it wasn't initialized yet.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	if condAddr == PthreadCondInitializer {
		if err := libKernel_initStaticCond(condHandlePtr); err != 0 {
			return err
		}
		condAddr = *(*uintptr)(unsafe.Pointer(condHandlePtr))
	}

	// Wait on it.
	hostCond := GetCond(condAddr)
	hostCond.Wait()

	fmt.Printf("%-120s %s waited on cond %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cond_signal"),
		color.Yellow.Sprintf("0x%X", condAddr),
	)
	return 0
}
