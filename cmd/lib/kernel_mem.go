package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

func libKernel_fake() uintptr {
	return 1
}

// 0x0000000000018290
// __int64 __fastcall sceKernelSetVirtualRangeName()
func libKernel_sceKernelSetVirtualRangeName(addr, length, namePtr uintptr) uintptr {
	if libKernel_mname(addr, length, namePtr) == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// 0x0000000000001C90
// __int64 __fastcall sub_1C90()
func libKernel_mname(addr, length, namePtr uintptr) uintptr {
	return libKernel_sys_mname(addr, length, namePtr)
}

func libKernel_sys_mname(addr, length, namePtr uintptr) uintptr {
	// Perform initial pointer checks.
	if addr == 0 {
		SetErrno(EINVAL)
		return ERR_PTR
	}

	name := "unnamed"
	if namePtr != 0 {
		name = ReadCString(namePtr)
	}

	// TODO: actually name the regions.
	fmt.Printf("%-120s %s marked %s bytes at %s as %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("mname"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", addr),
		color.Blue.Sprintf(name),
	)

	return 0
}

// 0x0000000000016FD0
// __int64 __fastcall sceKernelGetDirectMemorySize()
func libKernel_sceKernelGetDirectMemorySize() uintptr {
	// TODO: pthread_once
	size := uintptr(0x100000000) // 4GB

	fmt.Printf("%-120s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetDirectMemorySize"),
		color.Yellow.Sprintf("0x%X", size),
	)
	return size
}
