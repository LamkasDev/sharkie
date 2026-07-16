package pad

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pad"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/elokore/glfw/v3.4/glfw"
	"github.com/gookit/color"
)

// 0x0000000000000220
// __int64 __fastcall scePadInit(__m128)
func libScePad_scePadInit() uintptr {
	logger.Printf("scePadInit called\n")
	return 0
}

// 0x0000000000000580
// __int64 __fastcall scePadOpen(unsigned int, unsigned int, unsigned int, __m128)
func libScePad_scePadOpen(userId, padType, index, param uintptr) uintptr {
	handle := GlobalPadEngine.CreateHandle()

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePadOpen"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return uintptr(handle.Id)
}

// 0x00000000000020B0
// __int64 __fastcall scePadRead(__int64, __int64, __int64)
func libScePad_scePadRead(handleId, dataPtr, count uintptr) uintptr {
	if int32(handleId) <= 0 || dataPtr == 0 {
		return 0x802F0001 // SCE_PAD_ERROR_INVALID_HANDLE? some error
	}

	data := (*PadData)(unsafe.Pointer(dataPtr))
	data.Connected = true
	data.ConnectedCount = 1
	data.LeftStick.X = 128
	data.LeftStick.Y = 128
	data.RightStick.X = 128
	data.RightStick.Y = 128

	buttons := uint32(0)

	// We check GLFW keyboard state.
	if app.GlobalApplication != nil && app.GlobalApplication.Window != nil {
		window := app.GlobalApplication.Window
		if window.GetKey(glfw.KeyEnter) == glfw.Press {
			buttons |= PadButtonCross
			logger.Printf("pressed enter\n")
		}
		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			buttons |= PadButtonCircle
		}
		if window.GetKey(glfw.KeyUp) == glfw.Press {
			buttons |= PadButtonUp
		}
		if window.GetKey(glfw.KeyDown) == glfw.Press {
			buttons |= PadButtonDown
		}
		if window.GetKey(glfw.KeyLeft) == glfw.Press {
			buttons |= PadButtonLeft
		}
		if window.GetKey(glfw.KeyRight) == glfw.Press {
			buttons |= PadButtonRight
		}
	}

	data.Buttons = buttons

	return 0
}

// 0x00000000000036A0
// __int64 __fastcall scePadGetControllerInformation(unsigned int, __int64, __m128 _XMM0, __m128 _XMM1)
func libScePad_scePadGetControllerInformation(handleId, infoPtr uintptr) uintptr {
	if int32(handleId) <= 0 || infoPtr == 0 {
		return 0x802F0001
	}

	info := (*PadControllerInformation)(unsafe.Pointer(infoPtr))
	info.Connected = true
	info.ConnectedCount = 1
	info.ConnectionType = 0 // Local
	info.DeviceClass = 0    // Standard
	info.StickInfo.DeadZoneLeft = 0
	info.StickInfo.DeadZoneRight = 0
	info.TouchPadInfo.ResolutionX = 1920
	info.TouchPadInfo.ResolutionY = 900
	info.TouchPadInfo.PixelDensity = 44.0

	return 0
}
