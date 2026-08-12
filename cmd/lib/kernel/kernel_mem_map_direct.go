package kernel

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000017920
// __int64 __fastcall sceKernelMapDirectMemory(__int64 *, __int64, unsigned int, int, __int64, unsigned __int64)
func libKernel_sceKernelMapDirectMemory(addrPtr uintptr, length uint64, prot, flags int32, offset, alignment uintptr) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelMapDirectMemory(addrPtr, length, prot, flags, offset, alignment)
	if err != 0 {
		return err
	}

	return 0
}

// TODO: make this more robust.
// 0x0000000000018540
// __int64 __fastcall sceKernelMapNamedDirectMemory(__int64 *, int, int, int, int, __int64)
func libKernel_sceKernelMapNamedDirectMemory(addrPtr uintptr, length uint64, prot, flags int32, offset, alignment uintptr, namePtr Cstring) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelMapDirectMemory(addrPtr, length, prot, flags, offset, alignment)
	if err != 0 {
		return err
	}
	addrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrSlice))
	if posix.Mname(addr, length, namePtr) == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}

// TODO: make this more robust.
func libKernel_sys_sceKernelMapDirectMemory(addrPtr uintptr, length uint64, prot, flags int32, offset, alignment uintptr) uintptr {
	// Perform initial pointer checks.
	if alignment != 0 {
		if (alignment & (alignment - 1)) != 0 {
			logger.Printf("%-132s %s failed due to invalid alignment %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelMapDirectMemory"),
				color.Yellow.Sprintf("0x%X", alignment),
			)
			return SCE_KERNEL_ERROR_EINVAL
		}
		if (offset & (alignment - 1)) != 0 {
			logger.Printf("%-132s %s failed due to invalid offset %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelMapDirectMemory"),
				color.Yellow.Sprintf("0x%X", offset),
			)
			return SCE_KERNEL_ERROR_EINVAL
		}
	}
	if length == 0 || (length%MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to invalid size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapDirectMemory"),
			color.Yellow.Sprintf("0x%X", length),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	if addrPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapDirectMemory"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	// Get virtual address.
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	allocatedAddr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))
	if allocatedAddr == 0 || ((flags&MAP_FIXED) == 0 && !structs.GlobalMemoryManager.IsAddressRangeFree(allocatedAddr, uintptr(length))) {
		allocatedAddr = GlobalAllocator.GetNextAlignedAddress(uint64(alignment), length)
	}

	// Map memory.
	HookMapDirect(allocatedAddr, length, uint64(offset), 0, prot)
	HookMapMemoryVulkan(allocatedAddr, length, offset)
	if _, err := ProtectKernelMemory(allocatedAddr, length, prot); err != nil {
		logger.Printf("%-132s %s failed due to memory protection error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapDirectMemory"),
			err.Error(),
		)
		return SCE_KERNEL_ERROR_EFAULT
	}

	// Write back address.
	WriteAddress(addrPtr, allocatedAddr)

	logger.Printf("%-132s %s mapped %s bytes at %s (addrPtr=%s, offset=%s, prot=%s, flags=%s, alignment=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapDirectMemory"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Yellow.Sprintf("0x%X", offset),
		color.Blue.Sprint(MemoryProtName(prot)),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", alignment),
	)
	return 0
}
