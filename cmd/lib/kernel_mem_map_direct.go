package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000017920
// __int64 __fastcall sceKernelMapDirectMemory(__int64 *, __int64, unsigned int, int, __int64, unsigned __int64)
func libKernel_sceKernelMapDirectMemory(addrPtr, length, prot, flags, offset, alignment uintptr) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelMapDirectMemory(addrPtr, length, prot, flags, offset, alignment)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// TODO: make this more robust.
// 0x0000000000018540
// __int64 __fastcall sceKernelMapNamedDirectMemory(__int64 *, int, int, int, int, __int64)
func libKernel_sceKernelMapNamedDirectMemory(addrPtr, length, prot, flags, offset, alignment, namePtr uintptr) uintptr {
	// TODO: pthread_once
	err := libKernel_sys_sceKernelMapDirectMemory(addrPtr, length, prot, flags, offset, alignment)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}
	addrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	addr := uintptr(binary.LittleEndian.Uint64(addrSlice))
	if libKernel_mname(addr, length, namePtr) == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// TODO: make this more robust.
func libKernel_sys_sceKernelMapDirectMemory(addrPtr, length, prot, flags, offset, alignment uintptr) uintptr {
	// Perform initial pointer checks.
	if alignment != 0 {
		if !IsPowerOfTwo(alignment) {
			logger.Printf("%-120s %s failed due to invalid alignment %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelMapDirectMemory"),
				color.Yellow.Sprintf("0x%X", alignment),
			)
			SetErrno(EINVAL)
			return ERR_PTR
		}
		if (offset & (alignment - 1)) != 0 {
			logger.Printf("%-120s %s failed due to invalid offset %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelMapDirectMemory"),
				color.Yellow.Sprintf("0x%X", offset),
			)
			SetErrno(EINVAL)
			return ERR_PTR
		}
	}
	if length < MEMORY_ALIGN || (length&(MEMORY_ALIGN-1)) != 0 {
		logger.Printf("%-120s %s failed due to invalid size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapDirectMemory"),
			color.Yellow.Sprintf("0x%X", length),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}
	if addrPtr == 0 {
		logger.Printf("%-120s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapDirectMemory"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	// Write back offset.
	addrPtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(addrPtr)), 8)
	binary.LittleEndian.PutUint64(addrPtrSlice, uint64(offset))

	if _, err := ProtectKernelMemory(offset, length, prot); err != nil {
		logger.Printf("%-120s %s failed due to memory protection error (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMapDirectMemory"),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-120s %s mapped %s bytes at %s (addrPtr=%s, prot=%s, flags=%s, alignment=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMapDirectMemory"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", offset),
		color.Yellow.Sprintf("0x%X", addrPtr),
		color.Blue.Sprint(MemoryProtName(prot)),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", alignment),
	)
	return 0
}
