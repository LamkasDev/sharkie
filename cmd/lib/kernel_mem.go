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

// 0x0000000000001C90
// __int64 __fastcall sub_1C90()
func libKernel_mname(addr, length, namePtr uintptr) uintptr {
	// Perform initial pointer checks.
	if namePtr == 0 {
		return SCE_KERNEL_ERROR_EINVAL
	}

	// TODO: actually name the regions.
	name := ReadCString(namePtr)
	fmt.Printf("%-120s %s marked %s address %s as %s.\n",
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
