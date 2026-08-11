package pad

import (
	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pad"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/elokore/glfw/v3.4/glfw"
	"github.com/gookit/color"
)

var wasF11Pressed bool

// 0x0000000000000220
// __int64 __fastcall scePadInit(__m128)
func libScePad_scePadInit() uintptr {
	logger.Printf("scePadInit called\n")
	return 0
}

// 0x0000000000000580
// __int64 __fastcall scePadOpen(unsigned int, unsigned int, unsigned int, __m128)
func libScePad_scePadOpen(userId UserId, padType, index, param uintptr) uintptr {
	handle := GlobalPadEngine.CreateHandle()
	if config.GlobalConfig != nil && config.GlobalConfig.InputMode == "controller" {
		joystick := glfw.Joystick1
		if joystick.Present() {
			handle.Device = &ControllerDevice{
				Joystick: joystick,
			}
		} else {
			handle.Device = &KeyboardDevice{Window: app.GlobalApplication.Window}
		}
	} else {
		handle.Device = &KeyboardDevice{Window: app.GlobalApplication.Window}
	}

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePadOpen"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return uintptr(handle.Id)
}

// 0x00000000000036A0
// __int64 __fastcall scePadGetControllerInformation(unsigned int, __int64, __m128 _XMM0, __m128 _XMM1)
func libScePad_scePadGetControllerInformation(handleId uint32, info *PadControllerInformation) uintptr {
	handle := GlobalPadEngine.GetHandle(handleId)
	if handle == nil || info == nil {
		logger.Printf("%-132s %s failed due to invalid handle or info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePadGetControllerInformation"),
		)
		return 0x802F0001
	}
	handle.Device.GetControllerInformation(info)

	return 0
}
