package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000006F0
// __int64 __fastcall getpid()
func libKernel_getpid() uintptr {
	processId := uintptr(1001)
	logger.Printf("%-120s %s returned process id %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("getpid"),
		color.Green.Sprintf("%d", processId),
	)

	return processId
}

// 0x00000000000233E0
// __int64 __fastcall sceKernelGetProcessType()
func libKernel_sceKernelGetProcessType() uintptr {
	processType := uintptr(1)
	logger.Printf("%-120s %s returned process type %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetProcessType"),
		color.Blue.Sprintf("0x%X", processType),
	)

	return processType
}

// 0x000000000001A790
// __int64 sceKernelGetProcParam()
func libKernel_sceKernelGetProcParam() uintptr {
	module := emu.GlobalModuleManager.CurrentModule
	if module.ProcessParamSection != nil {
		addr := module.BaseAddress + uintptr(module.ProcessParamSection.PVaddr)
		logger.Printf("%-120s %s returning process parameters %s (relative=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetProcParam"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", module.ProcessParamSection.PVaddr),
		)
		return addr
	}

	logger.Printf("%-120s %s failed to return process parameters.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetProcParam"),
	)
	return 0
}
