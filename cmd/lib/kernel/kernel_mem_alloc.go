package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000175D0
// __int64 __fastcall sceKernelAllocateDirectMemory(__int64, __int64, __int64, __int64, int, _QWORD *)
func libKernel_sceKernelAllocateDirectMemory(searchStart, searchEnd uintptr, length, alignment uint64, memType int32, destPtr uintptr) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelAllocateDirectMemory(searchStart, searchEnd, length, alignment, memType, destPtr)
	if err == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
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
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}

	// Get the allocator.
	var allocator *Allocator
	if memType == SCE_KERNEL_MTYPE_WC_GARLIC || memType == SCE_KERNEL_MTYPE_WB_GARLIC {
		allocator = GlobalGpuAllocator
	} else {
		allocator = GlobalAllocator
	}

	// Get the next memory address from an already allocated buffer.
	allocatedAddr := allocator.GetNextAlignedAddress(alignment, length)
	if allocatedAddr == ERR_PTR {
		return ERR_PTR
	}

	// Write back pointer.
	WriteAddress(destPtr, allocatedAddr)

	HookAllocateDirect(allocatedAddr, length, memType)
	HookAllocateMemoryVulkan(allocatedAddr, length)

	logger.Printf("%-132s %s allocated %s bytes at %s (searchStart=%s, searchEnd=%s, alignment=%s, memType=%s, destPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelAllocateDirectMemory"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", searchStart),
		color.Yellow.Sprintf("0x%X", searchEnd),
		color.Yellow.Sprintf("0x%X", alignment),
		color.Blue.Sprint(MemoryTypeNames[memType]),
		color.Yellow.Sprintf("0x%X", destPtr),
	)
	return 0
}
