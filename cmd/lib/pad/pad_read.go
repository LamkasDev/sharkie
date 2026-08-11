package pad

import (
	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pad"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/elokore/glfw/v3.4/glfw"
	"github.com/gookit/color"
)

// 0x00000000000020B0
// __int64 __fastcall scePadRead(__int64, __int64, __int64)
func libScePad_scePadRead(handleId uint32, data *PadData, count uintptr) uintptr {
	handle := GlobalPadEngine.GetHandle(handleId)
	if handle == nil || data == nil {
		logger.Printf("%-132s %s failed due to invalid handle or data pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePadRead"),
		)
		return 0x802F0001
	}
	handle.Device.Read(data)

	// Global F11 fullscreen logic.
	if app.GlobalApplication != nil && app.GlobalApplication.Window != nil {
		window := app.GlobalApplication.Window
		isF11Pressed := window.GetKey(glfw.KeyF11) == glfw.Press
		if isF11Pressed && !wasF11Pressed {
			logger.Printf("pressed F11 (toggling fullscreen)\n")
			app.GlobalApplication.ToggleFullscreen()
		}
		wasF11Pressed = isF11Pressed
	}

	return 1
}

// 0x00000000000020A0
// __int64 __fastcall scePadReadState(__int64, __int64)
func libScePad_scePadReadState(handleId uint32, data *PadData) uintptr {
	handle := GlobalPadEngine.GetHandle(handleId)
	if handle == nil || data == nil {
		logger.Printf("%-132s %s failed due to invalid handle or data pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("scePadReadState"),
		)
		return 0x802F0001
	}
	handle.Device.Read(data)

	// Global F11 fullscreen logic.
	if app.GlobalApplication != nil && app.GlobalApplication.Window != nil {
		window := app.GlobalApplication.Window
		isF11Pressed := window.GetKey(glfw.KeyF11) == glfw.Press
		if isF11Pressed && !wasF11Pressed {
			logger.Printf("pressed F11 (toggling fullscreen)\n")
			app.GlobalApplication.ToggleFullscreen()
		}
		wasF11Pressed = isF11Pressed
	}

	return 0
}
