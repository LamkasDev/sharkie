package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
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

	return 0
}
