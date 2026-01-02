package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// TODO: make this more robust.
// 0x00000000000182C0
// __int64 __fastcall sceKernelMapNamedFlexibleMemory(__int64 *, int, int, int, int, __int64)
func libKernel_sceKernelMapNamedFlexibleMemory(addrPtr, length, prot, flags, namePtr uintptr) uintptr {
	// TODO: this doing other silly stuff
	err := libKernel_sceKernelMapFlexibleMemory(addrPtr, length, prot, flags)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}
	addrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrSlice))
	if libKernel_mname(addr, length, namePtr) == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

// TODO: make this more robust.
// 0x0000000000017330
// __int64 __fastcall sceKernelMapFlexibleMemory(__int64 *, unsigned __int64, unsigned int, unsigned int)
func libKernel_sceKernelMapFlexibleMemory(addrPtr, length, prot, flags uintptr) uintptr {
	// Perform initial alignment and pointer checks.
	if length < MEMORY_ALIGN || (length&MEMORY_ALIGN_MASK) != 0 {
		logger.Printf("%-132s %s failed due to invalid alignment or size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
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

	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))

	if (flags&MAP_FIXED) != 0 && addr == 0 {
		logger.Printf("%-132s %s cleared incorrect MAP_FIXED flag.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
		)
		flags &= ^uintptr(MAP_FIXED)
	}

	if addr == 0 {
		addr = 0x880000000
	}

	allocatedAddr := libKernel_mmap(addr, length, prot, flags|MAP_ANON, ERR_PTR, 0)
	if allocatedAddr == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	WriteAddress(addrPtr, allocatedAddr)
	logger.Printf("%-132s %s stored pointer at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapFlexibleMemory"),
		color.Yellow.Sprintf("0x%X", addrPtr),
	)

	return 0
}

// TODO: make this more robust.
// 0x0000000000018400
// __int64 __fastcall sceKernelMapNamedSystemFlexibleMemory(__int64 *, unsigned __int64, unsigned int, unsigned int, __int64)
func libKernel_sceKernelMapNamedSystemFlexibleMemory(addrPtr, length, prot, flags, namePtr uintptr) uintptr {
	// Perform initial alignment and pointer checks.
	if length < MEMORY_ALIGN || (length&MEMORY_ALIGN_MASK) != 0 {
		logger.Printf("%-132s %s failed due to invalid alignment or size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
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

	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))

	if (flags&MAP_FIXED) != 0 && addr == 0 {
		logger.Printf("%-132s %s cleared incorrect MAP_FIXED flag.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
		)
		flags &= ^uintptr(MAP_FIXED)
	}

	allocatedAddr := libKernel_mmap(addr, length, prot, flags|MAP_ANON|MAP_SYSTEM, ERR_PTR, 0)
	if allocatedAddr == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	WriteAddress(addrPtr, allocatedAddr)
	if libKernel_mname(allocatedAddr, length, namePtr) == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}
	logger.Printf("%-132s %s stored pointer at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
		color.Yellow.Sprintf("0x%X", addrPtr),
	)

	return 0
}
