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

// 0x0000000000018400
// __int64 __fastcall sceKernelMapNamedSystemFlexibleMemory(__int64 *, unsigned __int64, unsigned int, unsigned int, __int64)
func libKernel_sceKernelMapNamedSystemFlexibleMemory(addrPtr uintptr, length uint64, prot, flags int32, namePtr Cstring) uintptr {
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

	// Check adress alignment, round down hint to page boundary.
	isFixed := (flags & MAP_FIXED) != 0
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))
	if isFixed && addr != 0 && (addr&(uintptr(MemoryPageSize)-1)) != 0 {
		logger.Printf("%-132s %s failed due to invalid fixed address alignment %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	alignedAddr := addr & ^(uintptr(MemoryPageSize) - 1)
	alignedSize := (length + MemoryPageSize - 1) & ^(MemoryPageSize - 1)

	// Get virtual address.
	allocatedAddr := alignedAddr
	if allocatedAddr == 0 {
		allocatedAddr = GlobalAllocator.GetNextAlignedAddress(MemoryPageSize, length)
	}
	if !structs.GlobalMemoryManager.IsAddressRangeUnmapped(alignedAddr, uintptr(alignedSize)) {
		allocatedAddr = 0
		if !isFixed {
			allocatedAddr = GlobalAllocator.GetNextAlignedAddress(MemoryPageSize, length)
		}
	}
	if allocatedAddr == 0 {
		logger.Printf("%-132s %s failed allocating memory (addrPtr=%s, length=%s, prot=%s, flags=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
			color.Yellow.Sprintf("0x%X", addrPtr),
			color.Yellow.Sprintf("0x%X", length),
			color.Blue.Sprint(MemoryProtName(prot)),
			color.Yellow.Sprintf("0x%X", flags),
		)
		emu.SetErrno(ENOMEM)
		return ERR_PTR
	}
	if !isFixed && alignedAddr != 0 && allocatedAddr != alignedAddr {
		logger.Printf("%-132s %s ignored allocation address hint (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
			color.Yellow.Sprintf("0x%X", alignedAddr),
			color.Yellow.Sprintf("0x%X", allocatedAddr),
		)
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
