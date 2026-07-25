package pad

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pad"
	"github.com/elokore/glfw/v3.4/glfw"
)

type KeyboardDevice struct {
	Window *glfw.Window
}

func (k *KeyboardDevice) Read(data *PadData) {
	data.Connected = true
	data.ConnectedCount = 1
	data.LeftStick.X = 128
	data.LeftStick.Y = 128
	data.RightStick.X = 128
	data.RightStick.Y = 128

	buttons := uint32(0)
	if k.Window != nil {
		// Face Buttons.
		if k.Window.GetKey(glfw.KeyEnter) == glfw.Press {
			buttons |= PadButtonCross
		}
		if k.Window.GetKey(glfw.KeyEscape) == glfw.Press {
			buttons |= PadButtonCircle
		}
		if k.Window.GetKey(glfw.KeyQ) == glfw.Press {
			buttons |= PadButtonSquare
		}
		if k.Window.GetKey(glfw.KeyE) == glfw.Press {
			buttons |= PadButtonTriangle
		}

		// D-Pad.
		if k.Window.GetKey(glfw.KeyUp) == glfw.Press {
			buttons |= PadButtonUp
		}
		if k.Window.GetKey(glfw.KeyDown) == glfw.Press {
			buttons |= PadButtonDown
		}
		if k.Window.GetKey(glfw.KeyLeft) == glfw.Press {
			buttons |= PadButtonLeft
		}
		if k.Window.GetKey(glfw.KeyRight) == glfw.Press {
			buttons |= PadButtonRight
		}

		// Left stick.
		if k.Window.GetKey(glfw.KeyW) == glfw.Press {
			data.LeftStick.Y = 0
		} else if k.Window.GetKey(glfw.KeyS) == glfw.Press {
			data.LeftStick.Y = 255
		}
		if k.Window.GetKey(glfw.KeyA) == glfw.Press {
			data.LeftStick.X = 0
		} else if k.Window.GetKey(glfw.KeyD) == glfw.Press {
			data.LeftStick.X = 255
		}

		// Right stick.
		if k.Window.GetKey(glfw.KeyI) == glfw.Press {
			data.RightStick.Y = 0
		} else if k.Window.GetKey(glfw.KeyK) == glfw.Press {
			data.RightStick.Y = 255
		}
		if k.Window.GetKey(glfw.KeyJ) == glfw.Press {
			data.RightStick.X = 0
		} else if k.Window.GetKey(glfw.KeyL) == glfw.Press {
			data.RightStick.X = 255
		}
	}
	data.Buttons = buttons
}

func (k *KeyboardDevice) GetControllerInformation(info *PadControllerInformation) {
	info.Connected = true
	info.ConnectedCount = 1
	info.ConnectionType = 0 // Local
	info.DeviceClass = 0    // Standard
	info.StickInfo.DeadZoneLeft = 0
	info.StickInfo.DeadZoneRight = 0
	info.TouchPadInfo.ResolutionX = 1920
	info.TouchPadInfo.ResolutionY = 900
	info.TouchPadInfo.PixelDensity = 44.0
}

type ControllerDevice struct {
	Joystick glfw.Joystick
}

func (c *ControllerDevice) Read(data *PadData) {
	state := c.Joystick.GetGamepadState()
	if state == nil {
		data.Connected = false
		return
	}
	data.Connected = true
	data.ConnectedCount = 1

	// Map axes to stick data (GLFW is -1.0 to 1.0).
	data.LeftStick.X = uint8((state.Axes[glfw.AxisLeftX] + 1.0) * 127.5)
	data.LeftStick.Y = uint8((state.Axes[glfw.AxisLeftY] + 1.0) * 127.5)
	data.RightStick.X = uint8((state.Axes[glfw.AxisRightX] + 1.0) * 127.5)
	data.RightStick.Y = uint8((state.Axes[glfw.AxisRightY] + 1.0) * 127.5)

	// Map triggers to analog buttons (GLFW is -1.0 to 1.0).
	data.AnalogButtons.L2 = uint8((state.Axes[glfw.AxisLeftTrigger] + 1.0) * 127.5)
	data.AnalogButtons.R2 = uint8((state.Axes[glfw.AxisRightTrigger] + 1.0) * 127.5)

	// Map buttons.
	buttons := uint32(0)
	if state.Buttons[glfw.ButtonA] == glfw.Press {
		buttons |= PadButtonCross
	}
	if state.Buttons[glfw.ButtonB] == glfw.Press {
		buttons |= PadButtonCircle
	}
	if state.Buttons[glfw.ButtonX] == glfw.Press {
		buttons |= PadButtonSquare
	}
	if state.Buttons[glfw.ButtonY] == glfw.Press {
		buttons |= PadButtonTriangle
	}
	if state.Buttons[glfw.ButtonDpadUp] == glfw.Press {
		buttons |= PadButtonUp
	}
	if state.Buttons[glfw.ButtonDpadDown] == glfw.Press {
		buttons |= PadButtonDown
	}
	if state.Buttons[glfw.ButtonDpadLeft] == glfw.Press {
		buttons |= PadButtonLeft
	}
	if state.Buttons[glfw.ButtonDpadRight] == glfw.Press {
		buttons |= PadButtonRight
	}
	if state.Buttons[glfw.ButtonLeftBumper] == glfw.Press {
		buttons |= PadButtonL1
	}
	if state.Buttons[glfw.ButtonRightBumper] == glfw.Press {
		buttons |= PadButtonR1
	}
	if state.Buttons[glfw.ButtonLeftThumb] == glfw.Press {
		buttons |= PadButtonL3
	}
	if state.Buttons[glfw.ButtonRightThumb] == glfw.Press {
		buttons |= PadButtonR3
	}
	if state.Buttons[glfw.ButtonStart] == glfw.Press {
		buttons |= PadButtonOptions
	}
	if state.Buttons[glfw.ButtonBack] == glfw.Press {
		buttons |= PadButtonTouchPad
	}
	if data.AnalogButtons.L2 > 0 {
		buttons |= PadButtonL2
	}
	if data.AnalogButtons.R2 > 0 {
		buttons |= PadButtonR2
	}
	data.Buttons = buttons
}

func (c *ControllerDevice) GetControllerInformation(info *PadControllerInformation) {
	info.Connected = true
	info.ConnectedCount = 1
	info.ConnectionType = 0 // Local
	info.DeviceClass = 0    // Standard
	info.StickInfo.DeadZoneLeft = 0
	info.StickInfo.DeadZoneRight = 0
	info.TouchPadInfo.ResolutionX = 1920
	info.TouchPadInfo.ResolutionY = 900
	info.TouchPadInfo.PixelDensity = 44.0
}
