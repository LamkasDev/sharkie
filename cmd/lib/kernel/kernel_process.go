package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/kernel"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000233E0
// __int64 __fastcall sceKernelGetProcessType()
func libKernel_sceKernelGetProcessType() uintptr {
	processType := uintptr(1)
	logger.Printf("%-132s %s returned process type %s.\n",
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
		logger.Printf("%-132s %s returned process parameters %s (relative=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetProcParam"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", module.ProcessParamSection.POffset),
		)
		return addr
	}

	logger.Printf("%-132s %s failed to return process parameters.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetProcParam"),
	)
	return 0
}

// 0x0000000000015690
// __int64 sceKernelGetCpumode()
func libKernel_sceKernelGetCpumode() uintptr {
	// 67, haha.
	isCpu6, isCpu7 := GlobalPsfAttributes.Has(PsfAttributeSixCpuMode), GlobalPsfAttributes.Has(PsfAttributeSevenCpuMode)
	cpuMode := uintptr(0)
	if isCpu6 && isCpu7 {
		cpuMode = 2
	}
	if isCpu7 {
		cpuMode = 5
	}

	logger.Printf("%-132s %s returned cpu mode %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetCpumode"),
		color.Yellow.Sprintf("0x%X", cpuMode),
	)
	return cpuMode
}

// 0x0000000000014BE0
// __int64 __fastcall sceKernelUsleep(unsigned int)
func libKernel_sceKernelUsleep(micros uint32) uintptr {
	err := posix.Usleep(micros)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014B50
// __int64 __fastcall sceKernelNanosleep(__int128 *, __int64)
func libKernel_sceKernelNanosleep(timestamp, remainingTimestamp *Timestamp) uintptr {
	err := posix.Nanosleep(timestamp, remainingTimestamp)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}
