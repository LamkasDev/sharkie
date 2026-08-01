package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000001A620
// __int64 sceKernelGetTscFrequency()
func libKernel_sceKernelGetTscFrequency() uintptr {
	freq := uintptr(lib_structs.TSC_FREQUENCY)

	if logger.LogMisc {
		logger.Printf("%-132s %s returned frequency %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetTscFrequency"),
			color.Yellow.Sprintf("0x%X", freq),
		)
	}
	return freq
}

// 0x000000000001A690
// unsigned __int64 sceKernelReadTsc()
func libKernel_sceKernelReadTsc() uintptr {
	ticks := lib_structs.ReadTsc()

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s ticks.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelReadTsc"),
			color.Green.Sprint(ticks),
		)
	}
	return ticks
}

// 0x000000000001A720
// __int64 sceKernelGetProcessTimeCounterFrequency()
func libKernel_sceKernelGetProcessTimeCounterFrequency() uintptr {
	freq := uintptr(lib_structs.TSC_FREQUENCY)

	if logger.LogMisc {
		logger.Printf("%-132s %s returned frequency %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetProcessTimeCounterFrequency"),
			color.Yellow.Sprintf("0x%X", freq),
		)
	}
	return freq
}

// 0x000000000001A700
// unsigned __int64 sceKernelGetProcessTimeCounter()
func libKernel_sceKernelGetProcessTimeCounter() uintptr {
	ticks := lib_structs.ReadUptimeTsc()

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s ticks.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetProcessTimeCounter"),
			color.Green.Sprint(ticks),
		)
	}
	return ticks
}
