package pad

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pad"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000036A0
// __int64 __fastcall scePadGetControllerInformation(unsigned int, __int64, __m128 _XMM0, __m128 _XMM1)
func libScePad_scePadGetControllerInformation(handleId uint32, info *PadControllerInformation) uintptr {
	handle := GlobalPadEngine.GetHandle(handleId)
	if handle == nil || info == nil {
		logger.Printf("%-132s %s failed due to invalid handle or info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePadGetControllerInformation"),
		)
		return 0x80920003
	}
	handle.Device.GetControllerInformation(info)

	return 0
}
