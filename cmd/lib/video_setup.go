package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/dce"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
	"github.com/gookit/color"
)

// 0x000000000000AAD0
// __int64 __fastcall sceVideoOutOpen(unsigned int, unsigned int, unsigned int, _DWORD *, __m128 _XMM0)
func libSceVideoOut_sceVideoOutOpen() uintptr {
	handleId := 1
	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutOpen"),
		color.Yellow.Sprintf("0x%X", handleId),
	)

	return uintptr(handleId)
}

// 0x000000000000BDE0
// __int64 __fastcall sceVideoOutSetFlipRate(int, unsigned int)
func libSceVideoOut_sceVideoOutSetFlipRate(rawHandle, flipRate uintptr) uintptr {
	handle := GlobalDisplayCoreEngine.GetHandleById(int(rawHandle))
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutSetFlipRate"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}

	handle.FlipRate = uint32(flipRate)

	logger.Printf("%-132s %s set %s's flip rate to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutSetFlipRate"),
		color.Yellow.Sprintf("0x%X", handle.Id),
		color.Green.Sprintf("%d", flipRate),
	)
	return 0
}

// 0x000000000000BB80
// __int64 __fastcall sceVideoOutGetBufferLabelAddress(int, _QWORD *)
func libSceVideoOut_sceVideoOutGetBufferLabelAddress(rawHandle, resultLabelBufferAddressPtr uintptr) uintptr {
	if resultLabelBufferAddressPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid result label buffer address pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetBufferLabelAddress"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_VALUE
	}
	handle := GlobalDisplayCoreEngine.GetHandleById(int(rawHandle))
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceVideoOutGetBufferLabelAddress"),
		)
		return SCE_VIDEO_OUT_ERROR_INVALID_HANDLE
	}

	*(*uintptr)(unsafe.Pointer(resultLabelBufferAddressPtr)) = handle.LabelBufferAddress

	logger.Printf("%-132s %s wrote %s's label buffer address %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutGetBufferLabelAddress"),
		color.Yellow.Sprintf("0x%X", handle.Id),
		color.Yellow.Sprintf("0x%X", handle.LabelBufferAddress),
		color.Yellow.Sprintf("0x%X", resultLabelBufferAddressPtr),
	)
	return 0
}
