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

// 0x00000000000182C0
// __int64 __fastcall sceKernelMapNamedFlexibleMemory(__int64 *, int, int, int, int, __int64)
func libKernel_sceKernelMapNamedFlexibleMemory(addrPtr uintptr, length uint64, prot, flags int32, namePtr Cstring) uintptr {
	err := libKernel_sceKernelMapFlexibleMemory(addrPtr, length, prot, flags)
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

// 0x0000000000017330
// __int64 __fastcall sceKernelMapFlexibleMemory(__int64 *, unsigned __int64, unsigned int, unsigned int)
func libKernel_sceKernelMapFlexibleMemory(addrPtr uintptr, length uint64, prot, flags int32) uintptr {
	if length == 0 || (length%MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to invalid size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
			color.Yellow.Sprintf("0x%X", length),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	if addrPtr == 0 {
		logger.Printf("%-132s %s failed due to address pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	// Get virtual address.
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	allocatedAddr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))
	if allocatedAddr == 0 || ((flags&0x10) == 0 && !structs.GlobalMemoryManager.IsAddressRangeFree(allocatedAddr, uintptr(length))) {
		allocatedAddr = GlobalAllocator.GetNextAlignedAddress(MemoryPageSize, length)
	}

	// Map memory.
	HookAllocateMemoryVulkan(allocatedAddr, length)
	HookMapMemoryVulkan(allocatedAddr, length, allocatedAddr)
	HookMap(allocatedAddr, length, prot)
	if _, err := ProtectKernelMemory(allocatedAddr, length, prot); err != nil {
		logger.Printf("%-132s %s failed due to memory protection error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
			err.Error(),
		)
		return SCE_KERNEL_ERROR_EFAULT
	}

	// Write back address.
	WriteAddress(addrPtr, allocatedAddr)

	logger.Printf("%-132s %s mapped %s bytes at %s (addrPtr=%s, prot=%s, flags=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Blue.Sprint(MemoryProtName(prot)),
		color.Yellow.Sprintf("0x%X", flags),
	)
	return 0
}

// 0x0000000000018400
// __int64 __fastcall sceKernelMapNamedSystemFlexibleMemory(__int64 *, unsigned __int64, unsigned int, unsigned int, __int64)
func libKernel_sceKernelMapNamedSystemFlexibleMemory(addrPtr uintptr, length uint64, prot, flags int32, namePtr Cstring) uintptr {
	if length == 0 || (length%MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to invalid size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
			color.Yellow.Sprintf("0x%X", length),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	if addrPtr == 0 {
		logger.Printf("%-132s %s failed due to address pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	// Get virtual address.
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	allocatedAddr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))
	if allocatedAddr == 0 || ((flags&0x10) == 0 && !structs.GlobalMemoryManager.IsAddressRangeFree(allocatedAddr, uintptr(length))) {
		allocatedAddr = GlobalAllocator.GetNextAlignedAddress(MemoryPageSize, length)
	}

	// Map memory.
	HookAllocateMemoryVulkan(allocatedAddr, length)
	HookMap(allocatedAddr, length, prot)
	HookMapMemoryVulkan(allocatedAddr, length, allocatedAddr)
	if _, err := ProtectKernelMemory(allocatedAddr, length, prot); err != nil {
		logger.Printf("%-132s %s failed due to memory protection error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
			err.Error(),
		)
		return SCE_KERNEL_ERROR_EFAULT
	}
	if posix.Mname(allocatedAddr, length, namePtr) == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
	}

	// Write back address.
	WriteAddress(addrPtr, allocatedAddr)

	logger.Printf("%-132s %s mapped %s bytes at %s (addrPtr=%s, prot=%s, flags=%s, name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Blue.Sprint(MemoryProtName(prot)),
		color.Yellow.Sprintf("0x%X", flags),
		color.Blue.Sprint(GoString(namePtr)),
	)
	return 0
}
