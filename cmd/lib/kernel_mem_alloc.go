package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000175D0
// __int64 __fastcall sceKernelAllocateDirectMemory(__int64, __int64, __int64, __int64, int, _QWORD *)
func libKernel_sceKernelAllocateDirectMemory(searchStart, searchEnd uintptr, length, alignment uint64, memType int32, destPtr uintptr) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelAllocateDirectMemory(searchStart, searchEnd, length, alignment, memType, destPtr)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

func libKernel_sys_sceKernelAllocateDirectMemory(searchStart, searchEnd uintptr, length, alignment uint64, memType int32, destPtr uintptr) uintptr {
	// Perform initial pointer checks.
	if length == 0 || destPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid length or pointer.\n",
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
	var directAddr uintptr
	if memType == SCE_KERNEL_MTYPE_WC_GARLIC || memType == SCE_KERNEL_MTYPE_WB_ONION {
		directAddr = GlobalAllocator.GpuMemoryCurrent
	} else {
		directAddr = GlobalAllocator.DirectMemoryCurrent
	}
	if directAddr%uintptr(alignment) != 0 {
		directAddr += uintptr(alignment) - (directAddr % uintptr(alignment))
	}

	// Allocate direct memory and perform alignment check.
	allocatedAddr := libKernel_mmap(directAddr, length, PROT_READ|PROT_WRITE, MAP_ANON|MAP_PRIVATE|MAP_FIXED, ERR_PTRI, 0)
	if allocatedAddr == ERR_PTR {
		return ERR_PTR
	}
	if allocatedAddr%uintptr(alignment) != 0 {
		logger.Printf("%-132s %s failed due to ignored alignment of %s (got=%s, wanted=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelAllocateDirectMemory"),
			color.Yellow.Sprintf("0x%X", alignment),
			color.Yellow.Sprintf("0x%X", allocatedAddr),
			color.Yellow.Sprintf("0x%X", directAddr),
		)
		return ERR_PTR
	}

	// Write back pointer.
	WriteAddress(destPtr, allocatedAddr)
	if memType == SCE_KERNEL_MTYPE_WC_GARLIC || memType == SCE_KERNEL_MTYPE_WB_ONION {
		GlobalAllocator.GpuMemoryCurrent = allocatedAddr + uintptr(length)
	} else {
		GlobalAllocator.DirectMemoryCurrent = allocatedAddr + uintptr(length)
	}

	logger.Printf("%-132s %s stored pointer at %s (type=%s, alignment=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelAllocateDirectMemory"),
		color.Yellow.Sprintf("0x%X", destPtr),
		color.Blue.Sprint(MemoryTypeNames[memType]),
		color.Yellow.Sprintf("0x%X", alignment),
	)
	return 0
}
