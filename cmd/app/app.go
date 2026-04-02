package app

import (
	"fmt"
	"log"
	"time"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/renderer"
	"github.com/elokore/cimgui-go-vulkan/imgui"
	"github.com/elokore/glfw/v3.4/glfw"
	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
	"github.com/xlab/closer"
)

var GlobalApplication *Application

type Application struct {
	as.BaseVulkanApp

	Renderer            *renderer.Renderer
	SwapchainDimensions *as.SwapchainDimensions
	Config              Config

	Monitor *glfw.Monitor
	Window  *glfw.Window
}

func SetupApplication() error {
	// Initialize GLFW and Vulkan.
	if err := glfw.Init(); err != nil {
		return fmt.Errorf("glfw init: %w", err)
	}
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	if err := vk.Init(); err != nil {
		return fmt.Errorf("vulkan init: %w", err)
	}
	GlobalApplication = &Application{}

	// Set up window and monitor.
	var videoMode *glfw.VidMode
	GlobalApplication.Monitor, videoMode = setupMonitor(0)
	GlobalApplication.SwapchainDimensions = getSwapchainDimensions(GlobalApplication.Monitor, videoMode)
	GlobalApplication.Window = createWindow(GlobalApplication.Monitor, videoMode, GlobalApplication.SwapchainDimensions)

	// Setup platform.
	if _, err := as.NewPlatform(GlobalApplication); err != nil {
		return fmt.Errorf("asche platform: %w", err)
	}

	// Setup renderer.
	GlobalApplication.Config = DefaultConfig()
	GlobalApplication.Renderer = renderer.NewRenderer(GlobalApplication.Context(), GlobalApplication.SwapchainDimensions)
	GlobalApplication.Renderer.Backend.AttachToExistingWindow(
		GlobalApplication.Window,
		GlobalApplication.Renderer.Handles.Instance,
		GlobalApplication.Renderer.Handles.Device,
		GlobalApplication.Renderer.Handles.PhysicalDevice,
		GlobalApplication.Renderer.Handles.GraphicsQueue,
		GlobalApplication.Renderer.PipelineCache,
		GlobalApplication.Renderer.Handles.GraphicsQueueFamilyIndex,
		GlobalApplication.Renderer.Handles.Context.SwapchainImageResources(),
		GlobalApplication.Renderer.SwapchainDimensions,
	)

	// Setup overlay.
	GlobalApplication.Renderer.Overlay = renderer.NewImguiOverlay(GlobalApplication.Renderer.Backend)

	return nil
}

func RunApplication() error {
	defer CloseApplication()

	// Start goroutine to consume new frames.
	consumeFramesDone := make(chan struct{})
	go GlobalApplication.Renderer.ConsumeFrames(consumeFramesDone)

	// Start the main render loop.
	exitC := make(chan struct{}, 1)

	frameDelay, _ := getRefreshRate(GlobalApplication.Monitor)
	fpsTicker := time.NewTicker(frameDelay)
	defer fpsTicker.Stop()
	for {
		select {
		case <-exitC:
			GlobalApplication.Renderer.FrameSource.IsClosing.Store(true)
			close(GlobalApplication.Renderer.FrameSource.Channel)
			<-consumeFramesDone
			logger.Println("renderer: main loop exited")
			return nil
		case <-fpsTicker.C:
			if GlobalApplication.Window.ShouldClose() {
				exitC <- struct{}{}
				continue
			}
			glfw.PollEvents()

			imageIdx, outdated, err := GlobalApplication.Renderer.Handles.Context.AcquireNextImage()
			if err != nil {
				panic(err)
			}
			if outdated {
				panic(fmt.Errorf("AcquireNextImage: %w", err))
			}

			if GlobalApplication.Renderer.FramebufferTexture != nil {
				GlobalApplication.Renderer.FramebufferTexture.UploadPending(&GlobalApplication.Renderer.Handles)
			}
			GlobalApplication.Renderer.Backend.NewFrame(imageIdx)
			GlobalApplication.Renderer.Render()
			GlobalApplication.Renderer.Backend.RenderFrame(imageIdx)
			imgui.UpdatePlatformWindows()

			_, err = GlobalApplication.Renderer.Handles.Context.PresentImage(imageIdx)
			if err != nil {
				panic(fmt.Errorf("PresentImage: %w", err))
			}
		}
	}

	return nil
}

func CloseApplication() error {
	GlobalApplication.Renderer.Overlay.Destroy(GlobalApplication.Renderer.Backend)
	GlobalApplication.Renderer.Destroy()
	GlobalApplication.Window.Destroy()
	glfw.Terminate()
	closer.Close()

	return nil
}

func (app *Application) VulkanSwapchainDimensions() *as.SwapchainDimensions {
	return app.SwapchainDimensions
}

func (app *Application) SetSwapchainDimensions(dimensions *as.SwapchainDimensions) {
	// TODO: this
	app.SwapchainDimensions = dimensions
}

func (app *Application) VulkanSurface(instance vk.Instance) (surface vk.Surface) {
	surfPtr, err := app.Window.CreateWindowSurface(instance, nil)
	if err != nil {
		log.Printf("renderer: CreateWindowSurface: %v", err)
		return vk.NullSurface
	}
	return vk.SurfaceFromPointer(surfPtr)
}

func (app *Application) VulkanLayers() []string {
	validationLayers := []string{}
	if app.Config.DebugMode {
		validationLayers = append(validationLayers, "VK_LAYER_KHRONOS_validation")
	} else {
		log.Println("vulkan: debug mode is off, not using validation layers")
	}

	return validationLayers
}

func (app *Application) VulkanDeviceExtensions() []string {
	return []string{
		"VK_KHR_swapchain",
	}
}

func (app *Application) VulkanInstanceExtensions() []string {
	extensions := app.Window.GetRequiredInstanceExtensions()
	if app.Config.DebugMode {
		extensions = append(extensions, "VK_EXT_debug_report")
	}

	return extensions
}
