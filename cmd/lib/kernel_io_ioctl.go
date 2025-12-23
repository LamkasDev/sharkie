package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000000970
// __int64 __fastcall ioctl()
func libKernel_ioctl(pathPtr uintptr, flags uintptr, mode uintptr) uintptr {
	return libKernel_sys_ioctl(pathPtr, flags, mode)
}

func libKernel_sys_ioctl(fd, request, argPtr uintptr) uintptr {
	file, ok := GlobalFilesystem.Descriptors[int32(fd)]
	if !ok {
		fmt.Printf("%-120s %s requested %s with argument at %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprintf("[ioctl on unknown %d]", fd),
			color.Yellow.Sprintf("0x%X", request),
			color.Yellow.Sprintf("0x%X", argPtr),
		)
		return 0
	}

	fmt.Printf("%-120s %s requested %s with argument at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("[ioctl on %s]", file.Path),
		color.Yellow.Sprintf("0x%X", request),
		color.Yellow.Sprintf("0x%X", argPtr),
	)
	return 0
}
