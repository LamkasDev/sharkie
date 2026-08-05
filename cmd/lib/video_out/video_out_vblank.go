package video_out

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000C7C0
// __int64 __fastcall sceVideoOutAddVblankEvent(unsigned int, int, __int64, double)
func libSceVideoOut_sceVideoOutAddVblankEvent(equeueHandle, rawHandle, userData uintptr) uintptr {
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutAddVblankEvent"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}

	handle.VblankEvents = append(handle.VblankEvents, VideoOutEvent{
		EqueueHandle: equeueHandle,
		UserData:     userData,
	})

	logger.Printf("%-132s %s added vblank event to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutAddVblankEvent"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return 0
}

// 0x000000000000BAD0
// __int64 __fastcall sceVideoOutGetVblankStatus(int, __int64)
func libSceVideoOut_sceVideoOutGetVblankStatus(rawHandle uintptr, vblankStatus *VideoOutVblankStatus) uintptr {
	if vblankStatus == nil {
		logger.Printf("%-132s %s failed due to invalid v-blank status pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetVblankStatus"),
		)
		return 0x80290002
	}
	handle, ok := GlobalDisplayCoreEngine.Handles[uint32(rawHandle)]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetVblankStatus"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}
	*vblankStatus = handle.VblankStatus

	if logger.LogGraphics {
		logger.Printf("%-132s %s returned %s's v-blank status.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetVblankStatus"),
			color.Yellow.Sprintf("0x%X", handle.Id),
		)
	}
	return 0
}
