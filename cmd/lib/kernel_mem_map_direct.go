package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// TODO: make this more robust.
// 0x0000000000018540
// __int64 __fastcall sceKernelMapNamedDirectMemory(__int64 *, int, int, int, int, __int64)
func libKernel_sceKernelMapNamedDirectMemory(addrPtr, length, prot, flags, offset, alignment, namePtr uintptr) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelMapNamedDirectMemory(addrPtr, length, prot, flags, offset, alignment)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}
	addrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrSlice))
	if libKernel_mname(addr, length, namePtr) == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// TODO: make this more robust.
func libKernel_sys_sceKernelMapNamedDirectMemory(addrPtr, length, prot, flags, offset, alignment uintptr) uintptr {
	// Perform initial pointer checks.
	if length < MEMORY_ALIGN || (length&MEMORY_ALIGN_MASK) != 0 {
		fmt.Printf("%-120s %s failed due to invalid alignment or size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedDirectMemory"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}
	if addrPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedDirectMemory"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	// Write back offset.
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	binary.LittleEndian.PutUint64(addrPtrSlice, uint64(offset))

	fmt.Printf("%-120s %s mapped %s bytes at %s. (addrPtr=%s)\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapNamedDirectMemory"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", offset),
		color.Yellow.Sprintf("0x%X", addrPtr),
	)
	return 0
}
