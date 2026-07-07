package kernel

import (
	"time"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000001A620
// __int64 sceKernelGetTscFrequency()
func libKernel_sceKernelGetTscFrequency() uintptr {
	freq := uintptr(lib_structs.TSC_FREQUENCY)
	logger.Printf("%-132s %s returned frequency %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetTscFrequency"),
		color.Yellow.Sprintf("0x%X", freq),
	)
	return freq
}

// 0x000000000001A690
// unsigned __int64 sceKernelReadTsc()
func libKernel_sceKernelReadTsc() uintptr {
	ticks := readTsc()
	logger.Printf("%-132s %s returned %s ticks.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelReadTsc"),
		color.Green.Sprint(ticks),
	)
	return ticks
}

func readTsc() uintptr {
	elapsed := time.Since(lib_structs.TscStartTime)
	return uintptr((elapsed.Nanoseconds() * int64(lib_structs.TSC_FREQUENCY)) / 1_000_000_000)
}
