package gnm_driver

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000019D0
// _BOOL8 sceGnmAreSubmitsAllowed()
func libSceGnmDriver_sceGnmAreSubmitsAllowed() uintptr {
	allowed := GlobalLiverpool.UnfinishedSubmits <= 0
	if !allowed {
		return 0
	}

	return 1
}

// 0x0000000000001720
// __int64 sceGnmSubmitDone()
func libSceGnmDriver_sceGnmSubmitDone() int64 {
	// Wait for work to finish.
	GlobalLiverpool.WaitIdleFinish()

	if logger.LogGraphics {
		logger.Printf("%-132s %s signaled done.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmSubmitDone"),
		)
	}
	return 0
}

// TODO: this isn't right
// 0x0000000000004020
// __int64 __fastcall sceGnmDingDong(unsigned int a1, unsigned int a2)
func libSceGnmDriver_sceGnmDingDong(vqId, nextOffsetsDw uint32) int64 {
	return libSceGnmDriver_sceGnmDingDongForWorkload(vqId, nextOffsetsDw, 0)
}

// TODO: this isn't right
// 0x0000000000003F60
// __int64 __fastcall sceGnmDingDongForWorkload(unsigned int, unsigned int)
func libSceGnmDriver_sceGnmDingDongForWorkload(vqId, nextOffsetsDw uint32, workloadId uintptr) int64 {
	if vqId == 0 {
		logger.Printf("%-132s %s skipped due to invalid ring index.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmDingDongForWorkload"),
		)
		return 0
	}

	if logger.LogGraphics {
		logger.Printf("%-132s %s dinged ring %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceGnmDingDongForWorkload"),
			color.Green.Sprintf("%d", vqId),
		)
	}
	return 0
}
