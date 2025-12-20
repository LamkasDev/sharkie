package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000125A0
// __int64 __fastcall mmap(__int64, __int64, __int64, __int64, __int64, __int64)
func libKernel_mmap(addr, length, prot, flags, fd, offset uintptr) uintptr {
	return libKernel_mmap_0(addr, length, prot, flags, fd, offset)
}

// 0x0000000000002990
// __int64 __fastcall mmap_0()
func libKernel_mmap_0(addr, length, prot, flags, fd, offset uintptr) uintptr {
	// Allocate memory and check error.
	allocatedAddr, err := libKernel_alloc(addr, length, prot, flags)
	if allocatedAddr == 0 {
		// If we're not required to return a fixed address, let's try again and let Windows choose.
		if (flags&MAP_FIXED) == 0 && addr != 0 {
			allocatedAddr, err = libKernel_alloc(0, length, prot, flags)
		}
	}
	if allocatedAddr == 0 {
		fmt.Printf("%-120s %s failed allocating memory.\n%s\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap_0"),
			err.Error(),
		)
		SetErrno(ENOMEM)
		return ERR_PTR
	}
	if addr != 0 && allocatedAddr != addr {
		fmt.Printf("%-120s %s ignored allocation address (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mmap_0"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", allocatedAddr),
		)
	}

	// TODO: zero memory?

	fmt.Printf("%-120s %s allocated %s bytes at %s (addr=%s, prot=%s, flags=%s, fd=%s, offset=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mmap_0"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", allocatedAddr),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", prot),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", offset),
	)
	return allocatedAddr
}

// 0x0000000000016580
// __int64 __fastcall sceKernelMmap(__int64, __int64, __int64, __int64, __int64, __int64, __int64 *)
func libKernel_sceKernelMmap(addr, length, prot, flags, fd, offset, retAddrPtr uintptr) uintptr {
	// Map memory.
	allocatedAddr := libKernel_mmap(addr, length, prot, flags, fd, offset)
	if allocatedAddr == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	// Write back pointer.
	if retAddrPtr != 0 {
		retAddrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(retAddrPtr)), 8)
		binary.LittleEndian.PutUint64(retAddrPtrSlice, uint64(allocatedAddr))
	}

	return 0
}

// 0x0000000000018400
// __int64 __fastcall sceKernelMapNamedSystemFlexibleMemory(__int64 *, unsigned __int64, unsigned int, unsigned int, __int64)
func libKernel_sceKernelMapNamedSystemFlexibleMemory(addrPtr, length, prot, flags, namePtr uintptr) uintptr {
	// Perform initial alignment and pointer checks.
	if length < MEMORY_ALIGN || (length&MEMORY_ALIGN_MASK) != 0 {
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
	if allocatedAddr == ERR_PTR {
		return GetErrno() - 0x7FFE0000
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

// 0x0000000000018540
// __int64 __fastcall sceKernelMapNamedDirectMemory(__int64 *, int, int, int, int, __int64)
func libKernel_sceKernelMapNamedDirectMemory(addrPtr, length, prot, flags, offset, alignment, namePtr uintptr) uintptr {
	// TODO: pthread_once
	// TODO: this doing other silly stuff
	err := libKernel_sys_sceKernelMapNamedDirectMemory(addrPtr, length, prot, flags, offset, alignment, namePtr)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

func libKernel_sys_sceKernelMapNamedDirectMemory(addrPtr, length, prot, flags, offset, alignment, namePtr uintptr) uintptr {
	// Perform initial pointer checks.
	if addrPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapNamedDirectMemory"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	// Write back offset.
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	binary.LittleEndian.PutUint64(addrPtrSlice, uint64(offset))

	name := "unnamed"
	if namePtr != 0 {
		name = ReadCString(namePtr)
	}
	fmt.Printf("%-120s %s mapped %s bytes at %s as %s. (addrPtr=%s)\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapNamedDirectMemory"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", offset),
		color.Blue.Sprintf(name),
		color.Yellow.Sprintf("0x%X", addrPtr),
	)
	return 0
}
