package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/gookit/color"
)

const (
	MAP_FAILED = ^uintptr(0) // -1
	MAP_FIXED  = 0x10
	MAP_ANON   = 0x1000
)

const (
	SCE_KERNEL_ERROR_EINVAL = 0x80020016
	SCE_KERNEL_ERROR_ENOMEM = 0x8002000C

	// Alignment requirement from decompilation (16KB)
	ALIGN_MASK = 0x3FFF
)

// 0x00000000000125A0
// __int64 __fastcall mmap(__int64, __int64, __int64, __int64, __int64, __int64)
func libKernel_mmap(addr, length, prot, flags, fd, offset uintptr) uintptr {
	return libKernel_mmap_0(addr, length, prot, flags, fd, offset)
}

// 0x0000000000002990
// __int64 __fastcall mmap_0()
func libKernel_mmap_0(addr, length, prot, flags, fd, offset uintptr) uintptr {
	if (flags & MAP_FIXED) == 0 {
		addr = 0
	}
	allocatedAddr, err := libKernel_alloc(addr, length, prot, flags)
	if allocatedAddr == 0 {
		fmt.Printf("%-120s %s failed allocating memory.\n%s\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap_0"),
			err.Error(),
		)
		SetErrno(ENOMEM)

		return MAP_FAILED
	}
	fmt.Printf("%-120s %s allocated memory at %s (addr=%s, length=%s, prot=%s, flags=%s, fd=%s, offset=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mmap_0"),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", prot),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", offset),
	)

	// TODO: zero memory?

	return allocatedAddr
}

// 0x0000000000016580
// __int64 __fastcall sceKernelMmap(__int64, __int64, __int64, __int64, __int64, __int64, __int64 *)
func libKernel_sceKernelMmap(addr, length, prot, flags, fd, offset, retAddrPtr uintptr) uintptr {
	err := libKernel_mmap(addr, length, prot, flags, fd, offset)
	if err == MAP_FAILED {
		err = GetErrno()
		return uintptr(uint32(err) - 0x7FFE0000)
	}
	if retAddrPtr != 0 {
		retAddrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(retAddrPtr)), 8)
		binary.LittleEndian.PutUint64(retAddrPtrSlice, uint64(err))
	}

	return 0
}

// 0x0000000000018400
// __int64 __fastcall sceKernelMapNamedSystemFlexibleMemory(__int64 *, unsigned __int64, unsigned int, unsigned int, __int64)
func libKernel_sceKernelMapNamedSystemFlexibleMemory(addrPtr, length, prot, flags, namePtr uintptr) uintptr {
	if length < 0x4000 || (length&ALIGN_MASK) != 0 {
		fmt.Printf("%-120s %s failed due to invalid alignment or size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	if addrPtr == 0 {
		fmt.Printf("%-120s %s failed due to address pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrPtrSlice))

	flags |= 0x3000
	allocatedAddr := libKernel_mmap(addr, length, prot, flags, ^uintptr(0), 0)
	if allocatedAddr == ^uintptr(0) {
		return SCE_KERNEL_ERROR_ENOMEM
	}

	libKernel_mname(allocatedAddr, flags, namePtr)
	binary.LittleEndian.PutUint64(addrPtrSlice, uint64(allocatedAddr))
	fmt.Printf("%-120s %s stored pointer at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapNamedSystemFlexibleMemory"),
		color.Yellow.Sprintf("0x%X", addrPtr),
	)

	return 0
}

// 0x0000000000001C90
// __int64 __fastcall sub_1C90()
func libKernel_mname(addr, length, namePtr uintptr) uintptr {
	if namePtr == 0 {
		return 0
	}

	name := ReadCString(namePtr)
	fmt.Printf("%-120s %s marked address %s as %s (%s bytes).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mname"),
		color.Yellow.Sprintf("0x%X", addr),
		color.Blue.Sprintf(name),
		color.Gray.Sprintf("%d", length),
	)

	return 0
}
