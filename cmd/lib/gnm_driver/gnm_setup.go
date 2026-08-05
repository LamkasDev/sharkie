package gnm_driver

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000038B0
// __int64 sceGnmGetGpuCoreClockFrequency()
func libSceGnmDriver_sceGnmGetGpuCoreClockFrequency() uintptr {
	// TODO: neo check.
	freq := uintptr(800_000_000)

	if logger.LogGraphics {
		logger.Printf("%-132s %s returned frequency.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmGetGpuCoreClockFrequency"),
		)
	}
	return freq
}
