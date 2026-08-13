package kernel

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000018610
// __int64 __fastcall sceKernelReserveVirtualRange(__int64 *, __int64, int, unsigned __int64 _RCX)
func libKernel_sceKernelReserveVirtualRange(addrPtr uintptr, length uint64, flags int32, alignment uintptr) uintptr {
	// Perform initial pointer checks.
	if alignment != 0 {
		if (alignment & (alignment - 1)) != 0 {
			logger.Printf("%-132s %s failed due to invalid alignment %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelReserveVirtualRange"),
				color.Yellow.Sprintf("0x%X", alignment),
			)
			emu.SetErrno(EINVAL)
			return ERR_PTR
		}
	}
	if length == 0 || (length%MemoryPageSize) != 0 {
		logger.Printf("%-132s %s failed due to invalid size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelReserveVirtualRange"),
			color.Yellow.Sprintf("0x%X", length),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	if addrPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelReserveVirtualRange"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Check adress alignment, round down hint to page boundary.
	isFixed := (flags & MAP_FIXED) != 0
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))
	if alignment == 0 {
		alignment = uintptr(MemoryPageSize)
	}
	if isFixed && addr != 0 && (addr&(alignment-1)) != 0 {
		logger.Printf("%-132s %s failed due to invalid fixed address alignment %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelReserveVirtualRange"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	alignedAddr := addr & ^(alignment - 1)
	alignedSize := (length + uint64(alignment) - 1) & ^(uint64(alignment) - 1)

	// Get virtual address.
	allocatedAddr := alignedAddr
	if allocatedAddr == 0 {
		allocatedAddr = GlobalAllocator.GetNextAlignedAddress(uint64(alignment), length)
	}
	if !structs.GlobalMemoryManager.IsAddressRangeFree(alignedAddr, uintptr(alignedSize)) {
		allocatedAddr = 0
		if !isFixed {
			allocatedAddr = GlobalAllocator.GetNextAlignedAddress(uint64(alignment), length)
		}
	}
	if allocatedAddr == 0 {
		logger.Printf("%-132s %s failed reserving memory (addrPtr=%s, length=%s, flags=%s, alignment=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelReserveVirtualRange"),
			color.Yellow.Sprintf("0x%X", addrPtr),
			color.Yellow.Sprintf("0x%X", length),
			color.Yellow.Sprintf("0x%X", flags),
			color.Yellow.Sprintf("0x%X", alignment),
		)
		emu.SetErrno(ENOMEM)
		return ERR_PTR
	}
	if !isFixed && alignedAddr != 0 && allocatedAddr != alignedAddr {
		/* logger.Printf("%-132s %s ignored reservation address hint (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelReserveVirtualRange"),
			color.Yellow.Sprintf("0x%X", alignedAddr),
			color.Yellow.Sprintf("0x%X", allocatedAddr),
		) */
	}

	// Reserve memory.
	HookReserve(allocatedAddr, length)

	// Write back address.
	WriteAddress(addrPtr, allocatedAddr)

	logger.Printf("%-132s %s reserved %s bytes at %s (addrPtr=%s, flags=%s, alignment=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelReserveVirtualRange"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", alignment),
	)
	return 0
}
