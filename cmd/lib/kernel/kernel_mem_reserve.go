package kernel

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
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

	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))

	allocatedAddr, err := AllocKernelMemory(addr, length, PROT_NONE, flags|MAP_ANON, alignment)
	if allocatedAddr == 0 {
		logger.Printf("%-132s %s failed due to allocation error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelReserveVirtualRange"),
			err.Error(),
		)
		return emu.GetErrno() - SonyErrorOffset
	}

	HookMap(allocatedAddr, length, PROT_NONE)
	WriteAddress(addrPtr, allocatedAddr)

	logger.Printf("%-132s %s mapped %s bytes at %s (addrPtr=%s, flags=%s, alignment=%s).\n",
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
