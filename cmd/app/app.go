// Package app handles GLFW, Vulkan and application setup.
package app

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/renderer"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	"github.com/elokore/glfw/v3.4/glfw"
	vk "github.com/goki/vulkan"
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
	go pprof.Do(context.Background(), pprof.Labels("name", "ConsumeFrames"), func(ctx context.Context) {
		GlobalApplication.Renderer.ConsumeFrames(consumeFramesDone)
	})

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

			GlobalApplication.Renderer.QueueMutex.Lock()
			imageIdx, outdated, err := GlobalApplication.Renderer.Handles.Context.AcquireNextImage()
			GlobalApplication.Renderer.QueueMutex.Unlock()
			if err != nil {
				panic(err)
			}
			if outdated {
				panic(fmt.Errorf("AcquireNextImage: %w", err))
			}

			GlobalApplication.Renderer.Backend.NewFrame(imageIdx)
			select {
			case <-GlobalApplication.Renderer.FrameReady:
			default:
			}
			GlobalApplication.Renderer.QueueMutex.Lock()
			GlobalApplication.Renderer.Render()
			GlobalApplication.Renderer.Backend.RenderFrame(imageIdx)
			GlobalApplication.Renderer.QueueMutex.Unlock()
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

func (app *Application) VulkanAPIVersion() vk.Version {
	return vk.Version(vk.MakeVersion(1, 2, 0))
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
	extensions := []string{
		"VK_KHR_swapchain",
		"VK_KHR_shader_non_semantic_info",
		"VK_EXT_pageable_device_local_memory",
		"VK_EXT_memory_priority",
		"VK_EXT_shader_subgroup_ballot",
		"VK_EXT_subgroup_size_control",
		"VK_EXT_provoking_vertex",
	}
	if runtime.GOOS == "linux" {
		extensions = append(extensions, "VK_KHR_external_memory_fd")
		extensions = append(extensions, "VK_EXT_external_memory_dma_buf")
	}
	if runtime.GOOS == "windows" {
		extensions = append(extensions, "VK_KHR_external_memory_win32")
	}

	return extensions
}

func (app *Application) VulkanDeviceCreateNext() unsafe.Pointer {
	vulkan12Features := &vk.PhysicalDeviceVulkan12Features{
		SType:                           vk.StructureTypePhysicalDeviceVulkan12Features,
		RuntimeDescriptorArray:          vk.True,
		DescriptorBindingPartiallyBound: vk.True,
		DescriptorBindingSampledImageUpdateAfterBind: vk.True,
		DescriptorBindingStorageImageUpdateAfterBind: vk.True,
		ShaderSampledImageArrayNonUniformIndexing:    vk.True,
		ScalarBlockLayout:   vk.True,
		BufferDeviceAddress: vk.True,
	}
	features2 := &vk.PhysicalDeviceFeatures2{
		SType: vk.StructureTypePhysicalDeviceFeatures2,
		PNext: unsafe.Pointer(vulkan12Features),
		Features: vk.PhysicalDeviceFeatures{
			ShaderInt64:                          vk.True,
			SampleRateShading:                    vk.True,
			IndependentBlend:                     vk.True,
			GeometryShader:                       vk.True,
			TessellationShader:                   vk.True,
			FragmentStoresAndAtomics:             vk.True,
			ShaderStorageImageReadWithoutFormat:  vk.True,
			ShaderStorageImageWriteWithoutFormat: vk.True,
		},
	}
	pageableDeviceLocalMemoryFeatures := &as.VkPhysicalDevicePageableDeviceLocalMemoryFeaturesEXT{
		SType:                     as.StructureTypePhysicalDevicePageableDeviceLocalMemoryFeaturesExt,
		PNext:                     unsafe.Pointer(features2),
		PageableDeviceLocalMemory: uint32(vk.True),
	}
	provokingVertexFeatures := &vk.PhysicalDeviceProvokingVertexFeatures{
		SType:               vk.StructureTypePhysicalDeviceProvokingVertexFeatures,
		PNext:               unsafe.Pointer(pageableDeviceLocalMemoryFeatures),
		ProvokingVertexLast: vk.True,
	}
	subgroupSizeControlFeatures := &vulkan.VkPhysicalDeviceSubgroupSizeControlFeaturesEXT{
		SType:                vulkan.StructureTypePhysicalDeviceSubgroupSizeControlFeaturesExt,
		PNext:                unsafe.Pointer(provokingVertexFeatures),
		SubgroupSizeControl:  vk.True,
		ComputeFullSubgroups: vk.True,
	}

	return unsafe.Pointer(subgroupSizeControlFeatures)
}

func (app *Application) VulkanInstanceExtensions() []string {
	extensions := app.Window.GetRequiredInstanceExtensions()
	if app.Config.DebugMode {
		extensions = append(extensions, "VK_EXT_debug_report")
	}

	return extensions
}
