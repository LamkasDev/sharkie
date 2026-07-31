package kernel

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000016580
// __int64 __fastcall sceKernelMmap(__int64, __int64, __int64, __int64, __int64, __int64, __int64 *)
func libKernel_sceKernelMmap(addr uintptr, length uint64, prot, flags int32, fd FileDescriptor, offset, retAddrPtr uintptr) uintptr {
	allocatedAddr := posix.Mmap(addr, length, prot, flags, fd, offset)
	if allocatedAddr == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
	}

	if retAddrPtr != 0 {
		WriteAddress(retAddrPtr, allocatedAddr)
	}

	return 0
}

// 0x00000000000149E0
// __int64 sceKernelMunmap()
func libKernel_sceKernelMunmap(addr uintptr, length uint64) uintptr {
	err := posix.Munmap(addr, length)
	if err == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
	}

	HookFreeDirectVulkan(addr, length)

	return 0
}

// 0x0000000000018290
// __int64 __fastcall sceKernelSetVirtualRangeName()
func libKernel_sceKernelSetVirtualRangeName(addr uintptr, length uint64, namePtr Cstring) uintptr {
	err := posix.Mname(addr, length, namePtr)
	if err == ERR_PTR {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}

// 0x0000000000016FD0
// __int64 __fastcall sceKernelGetDirectMemorySize()
func libKernel_sceKernelGetDirectMemorySize() uint64 {
	// TODO: pthread_once
	size := GlobalAllocator.Size

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetDirectMemorySize"),
		color.Yellow.Sprintf("0x%X", size),
	)
	return size
}

// 0x00000000000181B0
// __int64 sceKernelVirtualQuery()
func libKernel_sceKernelVirtualQuery(addr uintptr, flags int32, infoPtr uintptr, infoSize uint64) uintptr {
	var info VirtualQueryInfo
	err := structs.GlobalMemoryManager.VirtualQuery(addr, flags, &info)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}
	if infoPtr != 0 {
		*(*VirtualQueryInfo)(unsafe.Pointer(infoPtr)) = info
	}

	logger.Printf("%-132s %s queried %s/%s (start=%s, end=%s, prot=%s, mem_type=%s, bitfield=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelVirtualQuery"),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", info.Start),
		color.Yellow.Sprintf("0x%X", info.End),
		color.Blue.Sprint(MemoryProtName(info.Protection)),
		color.Yellow.Sprintf("0x%X", info.MemoryType),
		color.Yellow.Sprintf("0x%X", info.Bitfield),
	)
	return 0
}

// 0x0000000000018B20
// __int64 __fastcall sceKernelDirectMemoryQuery(double)
func libKernel_sceKernelDirectMemoryQuery(addr uintptr, flags int32, infoPtr uintptr, infoSize uint64) uintptr {
	var info VirtualQueryInfo
	err := structs.GlobalMemoryManager.DirectMemoryQuery(addr, flags, &info)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}
	if infoPtr != 0 {
		*(*VirtualQueryInfo)(unsafe.Pointer(infoPtr)) = info
	}

	logger.Printf("%-132s %s queried %s (flags=%s) returning start=%s, end=%s, prot=%s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelDirectMemoryQuery"),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", info.Start),
		color.Yellow.Sprintf("0x%X", info.End),
		color.Blue.Sprint(MemoryProtName(info.Protection)),
	)
	return 0
}
