package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/gookit/color"
)

// 0x00000000000006F0
// __int64 __fastcall getpid()
func libKernel_getpid() uintptr {
	processId := uintptr(1001)
	fmt.Printf("%-120s %s returned process id %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("getpid"),
		color.Blue.Sprintf("0x%X", processId),
	)

	return processId
}

// 0x00000000000233E0
// __int64 __fastcall sceKernelGetProcessType()
func libKernel_sceKernelGetProcessType() uintptr {
	processType := uintptr(1)
	fmt.Printf("%-120s %s returned process type %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetProcessType"),
		color.Blue.Sprintf("0x%X", processType),
	)

	return processType
}
