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

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
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
	VulkanContext       *vulkan.VulkanContext
	Renderer            *renderer.Renderer
	SwapchainDimensions *backend.SwapchainDimensions
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
	GlobalApplication = &Application{
		Config:        DefaultConfig(),
		VulkanContext: vulkan.NewVulkanContext(),
	}

	// Set up window and monitor.
	var videoMode *glfw.VidMode
	GlobalApplication.Monitor, videoMode = setupMonitor(0)
	GlobalApplication.SwapchainDimensions = getSwapchainDimensions(GlobalApplication.Monitor, videoMode)
	GlobalApplication.Window = createWindow(GlobalApplication.Monitor, videoMode, GlobalApplication.SwapchainDimensions)

	// Setup context.
	cfg := vulkan.ContextConfig{
		ApiVersion:         uint32(GlobalApplication.VulkanAPIVersion()),
		AppVersion:         uint32(GlobalApplication.VulkanAPIVersion()), // Uses same for app right now
		AppName:            "base",
		InstanceExtensions: GlobalApplication.VulkanInstanceExtensions(),
		DeviceExtensions:   GlobalApplication.VulkanDeviceExtensions(),
		ValidationLayers:   GlobalApplication.VulkanLayers(),
		Debug:              GlobalApplication.Config.DebugMode,
		DeviceCreateNext:   GlobalApplication.VulkanDeviceCreateNext(),
		SurfaceFunc:        GlobalApplication.VulkanSurface,
		Dimensions:         GlobalApplication.SwapchainDimensions,
	}

	if err := GlobalApplication.VulkanContext.Init(cfg); err != nil {
		return fmt.Errorf("vulkan context init: %w", err)
	}

	// Setup renderer.
	GlobalApplication.Config = DefaultConfig()
	GlobalApplication.Renderer = renderer.NewRenderer(GlobalApplication.VulkanContext, GlobalApplication.SwapchainDimensions)
	GlobalApplication.Renderer.Backend.AttachToExistingWindow(
		GlobalApplication.Window,
		GlobalApplication.Renderer.Handles.Instance,
		GlobalApplication.Renderer.Handles.Device,
		GlobalApplication.Renderer.Handles.PhysicalDevice,
		GlobalApplication.Renderer.Handles.GraphicsQueue.Queue,
		GlobalApplication.Renderer.PipelineCache,
		GlobalApplication.Renderer.Handles.GraphicsQueueFamilyIndex,
		GlobalApplication.Renderer.Handles.Context.SwapchainImageResources,
		GlobalApplication.Renderer.SwapchainDimensions,
	)

	// Setup overlay.
	GlobalApplication.Renderer.Overlay = renderer.NewImguiOverlay(GlobalApplication.Renderer.Backend)

	return nil
}

func RunApplication() error {
	defer CloseApplication()

	// Start goroutine to consume ring work/frames.
	consumeRingWorkDone := make(chan struct{})
	go pprof.Do(context.Background(), pprof.Labels("name", "ConsumeRingWork"), func(ctx context.Context) {
		GlobalApplication.Renderer.ConsumeRingWork(consumeRingWorkDone)
	})
	consumeFlipsDone := make(chan struct{})
	go pprof.Do(context.Background(), pprof.Labels("name", "ConsumeFlips"), func(ctx context.Context) {
		GlobalApplication.Renderer.ConsumeFlips(consumeFlipsDone)
	})

	// Start the main render loop.
	exitC := make(chan struct{}, 1)

	frameDelay, _ := getRefreshRate(GlobalApplication.Monitor)
	fpsTicker := time.NewTicker(frameDelay)
	defer fpsTicker.Stop()
	for {
		select {
		case <-exitC:
			GlobalApplication.Renderer.RingWorkSource.IsClosing.Store(true)
			close(GlobalApplication.Renderer.RingWorkSource.Channel)
			GlobalApplication.Renderer.FrameSource.IsClosing.Store(true)
			close(GlobalApplication.Renderer.FrameSource.Channel)
			<-consumeRingWorkDone
			<-consumeFlipsDone
			logger.Println("renderer: main loop exited")
			return nil
		case <-fpsTicker.C:
			if GlobalApplication.Window.ShouldClose() {
				exitC <- struct{}{}
				continue
			}
			glfw.PollEvents()

			imageIdx, outdated, err := GlobalApplication.VulkanContext.AcquireNextImage()
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

			// Render UI and record command buffers .
			GlobalApplication.Renderer.Render()
			// cimgui-go-vulkan backend submits to GraphicsQueue manually, we must lock manually.
			GlobalApplication.Renderer.Handles.GraphicsQueue.Lock.Lock()
			GlobalApplication.Renderer.Backend.RenderFrame(imageIdx)
			GlobalApplication.Renderer.Handles.GraphicsQueue.Lock.Unlock()

			if err = GlobalApplication.VulkanContext.Submit(imageIdx); err != nil {
				panic(err)
			}
			_, err = GlobalApplication.VulkanContext.PresentImage(imageIdx)

			if err != nil {
				panic(fmt.Errorf("PresentImage: %w", err))
			}

			imgui.UpdatePlatformWindows()
		}
	}

	return nil
}

func CloseApplication() error {
	GlobalApplication.Renderer.Overlay.Destroy(GlobalApplication.Renderer.Backend)
	GlobalApplication.Renderer.Destroy()
	GlobalApplication.VulkanContext.Destroy()
	GlobalApplication.Window.Destroy()
	glfw.Terminate()
	closer.Close()

	return nil
}

func (app *Application) VulkanSwapchainDimensions() *backend.SwapchainDimensions {
	return app.SwapchainDimensions
}

func (app *Application) SetSwapchainDimensions(dimensions *backend.SwapchainDimensions) {
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
		"VK_KHR_dynamic_rendering",
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
	vulkanDynamicRenderingFeatures := &vk.PhysicalDeviceDynamicRenderingFeatures{
		SType:            vk.StructureTypePhysicalDeviceDynamicRenderingFeatures,
		DynamicRendering: vk.True,
	}
	vulkan12Features := &vk.PhysicalDeviceVulkan12Features{
		SType:                           vk.StructureTypePhysicalDeviceVulkan12Features,
		PNext:                           unsafe.Pointer(vulkanDynamicRenderingFeatures),
		RuntimeDescriptorArray:          vk.True,
		DescriptorBindingPartiallyBound: vk.True,
		DescriptorBindingSampledImageUpdateAfterBind:       vk.True,
		DescriptorBindingStorageImageUpdateAfterBind:       vk.True,
		DescriptorBindingUniformTexelBufferUpdateAfterBind: vk.True,
		DescriptorBindingStorageTexelBufferUpdateAfterBind: vk.True,
		ShaderSampledImageArrayNonUniformIndexing:          vk.True,
		ScalarBlockLayout:                 vk.True,
		BufferDeviceAddress:               vk.True,
		StorageBuffer8BitAccess:           vk.True,
		UniformAndStorageBuffer8BitAccess: vk.True,
		StoragePushConstant8:              vk.True,
		ShaderInt8:                        vk.True,
	}
	features2 := &vk.PhysicalDeviceFeatures2{
		SType: vk.StructureTypePhysicalDeviceFeatures2,
		PNext: unsafe.Pointer(vulkan12Features),
		Features: vk.PhysicalDeviceFeatures{
			ShaderInt16:                          vk.True,
			ShaderInt64:                          vk.True,
			SampleRateShading:                    vk.True,
			IndependentBlend:                     vk.True,
			GeometryShader:                       vk.True,
			TessellationShader:                   vk.True,
			FragmentStoresAndAtomics:             vk.True,
			ShaderStorageImageReadWithoutFormat:  vk.True,
			ShaderStorageImageWriteWithoutFormat: vk.True,
			DepthClamp:                           vk.True,
			TextureCompressionBC:                 vk.True,
			LogicOp:                              vk.True,
			SamplerAnisotropy:                    vk.True,
			WideLines:                            vk.True,
		},
	}
	pageableDeviceLocalMemoryFeatures := &vulkan.VkPhysicalDevicePageableDeviceLocalMemoryFeaturesEXT{
		SType:                     vulkan.StructureTypePhysicalDevicePageableDeviceLocalMemoryFeaturesExt,
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
	memoryPriorityFeatures := &vulkan.VkPhysicalDeviceMemoryPriorityFeaturesEXT{
		SType:          vulkan.StructureTypePhysicalDeviceMemoryPriorityFeaturesExt,
		PNext:          unsafe.Pointer(subgroupSizeControlFeatures),
		MemoryPriority: vk.True,
	}

	return unsafe.Pointer(memoryPriorityFeatures)
}

func (app *Application) VulkanInstanceExtensions() []string {
	extensions := app.Window.GetRequiredInstanceExtensions()
	if app.Config.DebugMode {
		extensions = append(extensions, "VK_EXT_debug_report")
	}

	return extensions
}
