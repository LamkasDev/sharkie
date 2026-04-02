package app

import (
	"fmt"
	"log"
	"time"

	"github.com/elokore/glfw/v3.4/glfw"
	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
)

func setupMonitor(monitorNum int) (*glfw.Monitor, *glfw.VidMode) {
	if monitorNum <= 0 {
		return nil, nil
	}
	monitors := glfw.GetMonitors()
	if monitorNum > len(monitors) {
		return nil, nil
	}

	monitor := monitors[monitorNum-1]
	return monitor, monitor.GetVideoMode()
}

func getRefreshRate(monitor *glfw.Monitor) (time.Duration, int) {
	var refreshRate int
	var videoMode *glfw.VidMode
	if monitor != nil {
		videoMode = monitor.GetVideoMode()
	} else {
		videoMode = glfw.GetPrimaryMonitor().GetVideoMode()
	}
	if videoMode != nil {
		refreshRate = videoMode.RefreshRate
	}
	if refreshRate <= 0 {
		log.Println("unable to parse monitor refresh rate. falling back to 60fps.")
		refreshRate = 60
	}

	return time.Second / time.Duration(refreshRate), refreshRate
}

func getSwapchainDimensions(monitor *glfw.Monitor, videoMode *glfw.VidMode) *as.SwapchainDimensions {
	// Exclusive fullscreen.
	if monitor != nil {
		return &as.SwapchainDimensions{
			Width: uint32(videoMode.Width), Height: uint32(videoMode.Height), Format: vk.FormatB8g8r8a8Unorm,
		}
	}

	// Windowed mode.
	return &as.SwapchainDimensions{
		Width: 800, Height: 600, Format: vk.FormatB8g8r8a8Unorm,
	}
}

func createWindow(monitor *glfw.Monitor, videoMode *glfw.VidMode, dimensions *as.SwapchainDimensions) *glfw.Window {
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	if monitor != nil {
		// Additional hints for exclusive fullscreen.
		glfw.WindowHint(glfw.RedBits, videoMode.RedBits)
		glfw.WindowHint(glfw.GreenBits, videoMode.GreenBits)
		glfw.WindowHint(glfw.BlueBits, videoMode.BlueBits)
		glfw.WindowHint(glfw.RefreshRate, videoMode.RefreshRate)

		// Ensure no decorations.
		glfw.WindowHint(glfw.Decorated, glfw.False)
		glfw.WindowHint(glfw.Floating, glfw.True)
	} else {
		// Normal window decorations.
		glfw.WindowHint(glfw.Decorated, glfw.True)
		glfw.WindowHint(glfw.Floating, glfw.False)
	}

	window, err := glfw.CreateWindow(int(dimensions.Width), int(dimensions.Height), "sharkie", monitor, nil)
	if err != nil {
		panic(fmt.Errorf("glfw.CreateWindow: %w", err))
	}

	// Additional fullscreen setup.
	if monitor != nil {
		// Force window to front and focus.
		window.Show()
		window.Focus()

		// Disable cursor in fullscreen.
		window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	}

	return window
}
