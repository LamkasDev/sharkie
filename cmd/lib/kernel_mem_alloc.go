package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000175D0
// __int64 __fastcall sceKernelAllocateDirectMemory(__int64, __int64, __int64, __int64, int, _QWORD *)
func libKernel_sceKernelAllocateDirectMemory(searchStart, searchEnd, length, alignment, memType, destPtr uintptr) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelAllocateDirectMemory(searchStart, searchEnd, length, alignment, memType, destPtr)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

func libKernel_sys_sceKernelAllocateDirectMemory(searchStart, searchEnd, length, alignment, memType, destPtr uintptr) uintptr {
	// Perform initial pointer checks.
	if length == 0 || destPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid length or pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelAllocateDirectMemory"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}
	if alignment == 0 {
		alignment = MEMORY_ALIGN
	}

	// Get the direct memory address.
	directAddr := GlobalAllocator.DirectMemoryCurrent
	if directAddr%alignment != 0 {
		directAddr += alignment - (directAddr % alignment)
	}

	// Allocate direct memory and perform alignment check.
	allocatedAddr := libKernel_mmap(directAddr, length, PROT_READ|PROT_WRITE, MAP_ANON|MAP_PRIVATE|MAP_FIXED, ERR_PTR, 0)
	if allocatedAddr == ERR_PTR {
		return ERR_PTR
	}
	if allocatedAddr%alignment != 0 {
		fmt.Printf("%-120s %s failed due to ignored alignment of %s (got=%s, wanted=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelAllocateDirectMemory"),
			color.Yellow.Sprintf("0x%X", alignment),
			color.Yellow.Sprintf("0x%X", allocatedAddr),
			color.Yellow.Sprintf("0x%X", directAddr),
		)
		return ERR_PTR
	}

	// Write back pointer.
	destPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(destPtr)), 8)
	binary.LittleEndian.PutUint64(destPtrSlice, uint64(allocatedAddr))
	GlobalAllocator.DirectMemoryCurrent = allocatedAddr + length

	memTypeName := fmt.Sprintf("unknown 0x%X", memType)
	if name, ok := MemoryTypeNames[memType]; ok {
		memTypeName = name
	}
	fmt.Printf("%-120s %s stored pointer at %s (type=%s, alignment=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelAllocateDirectMemory"),
		color.Yellow.Sprintf("0x%X", destPtr),
		color.Blue.Sprint(memTypeName),
		color.Yellow.Sprintf("0x%X", alignment),
	)
	return 0
}
