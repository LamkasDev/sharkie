package vulkan

import (
	"errors"
	"log"
	"unsafe"

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	vk "github.com/goki/vulkan"
)

type VulkanContext struct {
	Instance       vk.Instance
	PhysicalDevice vk.PhysicalDevice
	Device         vk.Device
	Surface        vk.Surface
	DebugCallback  vk.DebugReportCallback

	GraphicsQueueIndex uint32
	PresentQueueIndex  uint32
	GraphicsQueue      vk.Queue
	PresentQueue       vk.Queue

	MemoryProperties vk.PhysicalDeviceMemoryProperties
	GpuProperties    vk.PhysicalDeviceProperties

	Swapchain               vk.Swapchain
	SwapchainDimensions     *backend.SwapchainDimensions
	SwapchainImageResources []*backend.SwapchainImageResources

	FrameLag                 int
	FrameIndex               int
	ImageAcquiredSemaphores  []vk.Semaphore
	DrawCompleteSemaphores   []vk.Semaphore
	ImageOwnershipSemaphores []vk.Semaphore
	FrameFences              []vk.Fence

	CmdPool        vk.CommandPool
	PresentCmdPool vk.CommandPool
}

func NewVulkanContext() *VulkanContext {
	return &VulkanContext{
		FrameLag: 3,
	}
}

type ContextConfig struct {
	ApiVersion         uint32
	AppVersion         uint32
	AppName            string
	InstanceExtensions []string
	DeviceExtensions   []string
	ValidationLayers   []string
	Debug              bool
	DeviceCreateNext   unsafe.Pointer
	SurfaceFunc        func(vk.Instance) vk.Surface
	Dimensions         *backend.SwapchainDimensions
}

func (c *VulkanContext) Init(cfg ContextConfig) error {
	actualInstanceExtensions, err := InstanceExtensions()
	if err != nil {
		return err
	}
	instanceExtensions, missing := checkExisting(actualInstanceExtensions, safeStrings(cfg.InstanceExtensions))
	if missing > 0 {
		log.Println("vulkan warning: missing", missing, "required instance extensions during init")
	}

	actualValidationLayers, err := ValidationLayers()
	if err != nil {
		return err
	}
	validationLayers, missing := checkExisting(actualValidationLayers, safeStrings(cfg.ValidationLayers))
	if missing > 0 {
		log.Println("vulkan warning: missing", missing, "required validation layers during init")
	}

	var instance vk.Instance
	ret := vk.CreateInstance(&vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &vk.ApplicationInfo{
			SType:              vk.StructureTypeApplicationInfo,
			ApiVersion:         cfg.ApiVersion,
			ApplicationVersion: cfg.AppVersion,
			PApplicationName:   safeString(cfg.AppName),
			PEngineName:        "vulkango.com\x00",
		},
		EnabledExtensionCount:   uint32(len(instanceExtensions)),
		PpEnabledExtensionNames: instanceExtensions,
		EnabledLayerCount:       uint32(len(validationLayers)),
		PpEnabledLayerNames:     validationLayers,
	}, nil, &instance)
	if ret != vk.Success {
		return NewError(ret)
	}
	c.Instance = instance
	vk.InitInstance(instance)

	if cfg.Debug {
		ret := vk.CreateDebugReportCallback(instance, &vk.DebugReportCallbackCreateInfo{
			SType:       vk.StructureTypeDebugReportCallbackCreateInfo,
			Flags:       vk.DebugReportFlags(vk.DebugReportErrorBit | vk.DebugReportWarningBit),
			PfnCallback: dbgCallbackFunc,
		}, nil, &c.DebugCallback)
		if ret != vk.Success {
			return NewError(ret)
		}
	}

	var gpuCount uint32
	ret = vk.EnumeratePhysicalDevices(c.Instance, &gpuCount, nil)
	if ret != vk.Success {
		return NewError(ret)
	}
	if gpuCount == 0 {
		return errors.New("vulkan error: no GPU devices found")
	}
	gpus := make([]vk.PhysicalDevice, gpuCount)
	ret = vk.EnumeratePhysicalDevices(c.Instance, &gpuCount, gpus)
	if ret != vk.Success {
		return NewError(ret)
	}
	c.PhysicalDevice = gpus[0]
	vk.GetPhysicalDeviceProperties(c.PhysicalDevice, &c.GpuProperties)
	c.GpuProperties.Deref()
	vk.GetPhysicalDeviceMemoryProperties(c.PhysicalDevice, &c.MemoryProperties)
	c.MemoryProperties.Deref()

	actualDeviceExtensions, err := DeviceExtensions(c.PhysicalDevice)
	if err != nil {
		return err
	}
	deviceExtensions, missing := checkExisting(actualDeviceExtensions, safeStrings(cfg.DeviceExtensions))
	if missing > 0 {
		log.Println("vulkan warning: missing", missing, "required device extensions during init")
	}

	c.Surface = cfg.SurfaceFunc(c.Instance)
	if c.Surface == vk.NullSurface {
		return errors.New("vulkan error: surface required but not provided")
	}

	var queueCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(c.PhysicalDevice, &queueCount, nil)
	queueProperties := make([]vk.QueueFamilyProperties, queueCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(c.PhysicalDevice, &queueCount, queueProperties)
	if queueCount == 0 {
		return errors.New("vulkan error: no queue families found on GPU 0")
	}

	var graphicsFound, presentFound, separateQueue bool
	for i := uint32(0); i < queueCount; i++ {
		var supportsPresent vk.Bool32
		if graphicsFound {
			separateQueue = true
			vk.GetPhysicalDeviceSurfaceSupport(c.PhysicalDevice, i, c.Surface, &supportsPresent)
			if supportsPresent.B() {
				c.PresentQueueIndex = i
				presentFound = true
				break
			}
		}
		required := vk.QueueFlags(vk.QueueGraphicsBit | vk.QueueComputeBit)
		vk.GetPhysicalDeviceSurfaceSupport(c.PhysicalDevice, i, c.Surface, &supportsPresent)

		queueProperties[i].Deref()
		if queueProperties[i].QueueFlags&required != 0 {
			if supportsPresent.B() {
				c.GraphicsQueueIndex = i
				c.PresentQueueIndex = i
				graphicsFound = true
				break
			} else {
				c.GraphicsQueueIndex = i
				graphicsFound = true
			}
		}
	}
	if separateQueue && !presentFound {
		return errors.New("vulkan error: could not find separate queue with present capabilities")
	}
	if !graphicsFound {
		return errors.New("vulkan error: could not find a suitable queue family")
	}

	queueInfos := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: c.GraphicsQueueIndex,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	if separateQueue {
		queueInfos = append(queueInfos, vk.DeviceQueueCreateInfo{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: c.PresentQueueIndex,
			QueueCount:       1,
			PQueuePriorities: []float32{1.0},
		})
	}

	var device vk.Device
	ret = vk.CreateDevice(c.PhysicalDevice, &vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		PNext:                   cfg.DeviceCreateNext,
		QueueCreateInfoCount:    uint32(len(queueInfos)),
		PQueueCreateInfos:       queueInfos,
		EnabledExtensionCount:   uint32(len(deviceExtensions)),
		PpEnabledExtensionNames: deviceExtensions,
		EnabledLayerCount:       uint32(len(validationLayers)),
		PpEnabledLayerNames:     validationLayers,
	}, nil, &device)
	if ret != vk.Success {
		return NewError(ret)
	}
	c.Device = device

	var queue vk.Queue
	vk.GetDeviceQueue(c.Device, c.GraphicsQueueIndex, 0, &queue)
	c.GraphicsQueue = queue

	if separateQueue {
		var presentQueue vk.Queue
		vk.GetDeviceQueue(c.Device, c.PresentQueueIndex, 0, &presentQueue)
		c.PresentQueue = presentQueue
	} else {
		c.PresentQueue = c.GraphicsQueue
	}

	c.PreparePresent()
	c.PrepareSwapchain(cfg.Dimensions)
	c.Prepare()

	return nil
}

func (c *VulkanContext) Destroy() {
	if c.Device != nil {
		vk.DeviceWaitIdle(c.Device)
	}

	for i := range c.ImageAcquiredSemaphores {
		vk.DestroySemaphore(c.Device, c.ImageAcquiredSemaphores[i], nil)
		vk.DestroySemaphore(c.Device, c.DrawCompleteSemaphores[i], nil)
		if c.PresentQueueIndex != c.GraphicsQueueIndex {
			vk.DestroySemaphore(c.Device, c.ImageOwnershipSemaphores[i], nil)
		}
	}
	for i := range c.FrameFences {
		vk.DestroyFence(c.Device, c.FrameFences[i], nil)
	}
	for i := 0; i < len(c.SwapchainImageResources); i++ {
		c.SwapchainImageResources[i].Destroy(c.Device, c.CmdPool)
	}
	c.SwapchainImageResources = nil
	if c.Swapchain != vk.NullSwapchain {
		vk.DestroySwapchain(c.Device, c.Swapchain, nil)
		c.Swapchain = vk.NullSwapchain
	}
	vk.DestroyCommandPool(c.Device, c.CmdPool, nil)
	if c.PresentQueueIndex != c.GraphicsQueueIndex {
		vk.DestroyCommandPool(c.Device, c.PresentCmdPool, nil)
	}

	if c.Surface != vk.NullSurface {
		vk.DestroySurface(c.Instance, c.Surface, nil)
		c.Surface = vk.NullSurface
	}
	if c.Device != nil {
		vk.DestroyDevice(c.Device, nil)
		c.Device = nil
	}
	if c.DebugCallback != vk.NullDebugReportCallback {
		vk.DestroyDebugReportCallback(c.Instance, c.DebugCallback, nil)
	}
	if c.Instance != nil {
		vk.DestroyInstance(c.Instance, nil)
		c.Instance = nil
	}
}

func (c *VulkanContext) PreparePresent() {
	semaphoreCreateInfo := &vk.SemaphoreCreateInfo{
		SType: vk.StructureTypeSemaphoreCreateInfo,
	}
	fenceCreateInfo := &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}
	c.ImageAcquiredSemaphores = make([]vk.Semaphore, c.FrameLag)
	c.DrawCompleteSemaphores = make([]vk.Semaphore, c.FrameLag)
	c.ImageOwnershipSemaphores = make([]vk.Semaphore, c.FrameLag)
	c.FrameFences = make([]vk.Fence, c.FrameLag)
	for i := 0; i < c.FrameLag; i++ {
		ret := vk.CreateSemaphore(c.Device, semaphoreCreateInfo, nil, &c.ImageAcquiredSemaphores[i])
		orPanic(NewError(ret))
		ret = vk.CreateSemaphore(c.Device, semaphoreCreateInfo, nil, &c.DrawCompleteSemaphores[i])
		orPanic(NewError(ret))
		if c.PresentQueueIndex != c.GraphicsQueueIndex {
			ret = vk.CreateSemaphore(c.Device, semaphoreCreateInfo, nil, &c.ImageOwnershipSemaphores[i])
			orPanic(NewError(ret))
		}
		ret = vk.CreateFence(c.Device, fenceCreateInfo, nil, &c.FrameFences[i])
		orPanic(NewError(ret))
	}
}

func (c *VulkanContext) PrepareSwapchain(dimensions *backend.SwapchainDimensions) {
	var surfaceCapabilities vk.SurfaceCapabilities
	ret := vk.GetPhysicalDeviceSurfaceCapabilities(c.PhysicalDevice, c.Surface, &surfaceCapabilities)
	orPanic(NewError(ret))
	surfaceCapabilities.Deref()

	var formatCount uint32
	vk.GetPhysicalDeviceSurfaceFormats(c.PhysicalDevice, c.Surface, &formatCount, nil)
	formats := make([]vk.SurfaceFormat, formatCount)
	vk.GetPhysicalDeviceSurfaceFormats(c.PhysicalDevice, c.Surface, &formatCount, formats)

	var format vk.SurfaceFormat
	if formatCount == 1 {
		formats[0].Deref()
		if formats[0].Format == vk.FormatUndefined {
			format = formats[0]
			format.Format = vk.Format(dimensions.Format)
		} else {
			format = formats[0]
		}
	} else if formatCount == 0 {
		orPanic(errors.New("vulkan error: surface has no pixel formats"))
	} else {
		formats[0].Deref()
		format = formats[0]
	}

	var swapchainSize vk.Extent2D
	surfaceCapabilities.CurrentExtent.Deref()
	if surfaceCapabilities.CurrentExtent.Width == vk.MaxUint32 {
		swapchainSize.Width = dimensions.Width
		swapchainSize.Height = dimensions.Height
	} else {
		swapchainSize = surfaceCapabilities.CurrentExtent
	}

	swapchainPresentMode := vk.PresentModeFifo

	desiredSwapchainImages := surfaceCapabilities.MinImageCount + 1
	if surfaceCapabilities.MaxImageCount > 0 && desiredSwapchainImages > surfaceCapabilities.MaxImageCount {
		desiredSwapchainImages = surfaceCapabilities.MaxImageCount
	}

	var preTransform vk.SurfaceTransformFlagBits
	requiredTransforms := vk.SurfaceTransformIdentityBit
	supportedTransforms := surfaceCapabilities.SupportedTransforms
	if vk.SurfaceTransformFlagBits(supportedTransforms)&requiredTransforms != 0 {
		preTransform = requiredTransforms
	} else {
		preTransform = surfaceCapabilities.CurrentTransform
	}

	compositeAlpha := vk.CompositeAlphaOpaqueBit
	compositeAlphaFlags := []vk.CompositeAlphaFlagBits{
		vk.CompositeAlphaOpaqueBit,
		vk.CompositeAlphaPreMultipliedBit,
		vk.CompositeAlphaPostMultipliedBit,
		vk.CompositeAlphaInheritBit,
	}
	for i := 0; i < len(compositeAlphaFlags); i++ {
		alphaFlags := vk.CompositeAlphaFlags(compositeAlphaFlags[i])
		if surfaceCapabilities.SupportedCompositeAlpha&alphaFlags != 0 {
			compositeAlpha = compositeAlphaFlags[i]
			break
		}
	}

	var swapchain vk.Swapchain
	oldSwapchain := c.Swapchain
	ret = vk.CreateSwapchain(c.Device, &vk.SwapchainCreateInfo{
		SType:           vk.StructureTypeSwapchainCreateInfo,
		Surface:         c.Surface,
		MinImageCount:   desiredSwapchainImages,
		ImageFormat:     format.Format,
		ImageColorSpace: format.ColorSpace,
		ImageExtent: vk.Extent2D{
			Width:  swapchainSize.Width,
			Height: swapchainSize.Height,
		},
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:     preTransform,
		CompositeAlpha:   compositeAlpha,
		ImageArrayLayers: 1,
		ImageSharingMode: vk.SharingModeExclusive,
		PresentMode:      swapchainPresentMode,
		OldSwapchain:     oldSwapchain,
		Clipped:          vk.True,
	}, nil, &swapchain)
	orPanic(NewError(ret))

	if oldSwapchain != vk.NullSwapchain {
		vk.DestroySwapchain(c.Device, oldSwapchain, nil)
	}
	c.Swapchain = swapchain
	c.SwapchainDimensions = &backend.SwapchainDimensions{
		Width:  swapchainSize.Width,
		Height: swapchainSize.Height,
		Format: format.Format,
	}

	var imageCount uint32
	ret = vk.GetSwapchainImages(c.Device, c.Swapchain, &imageCount, nil)
	orPanic(NewError(ret))
	swapchainImages := make([]vk.Image, imageCount)
	ret = vk.GetSwapchainImages(c.Device, c.Swapchain, &imageCount, swapchainImages)
	orPanic(NewError(ret))

	// Re-allocate semaphores if image count changed.
	if len(c.ImageAcquiredSemaphores) != int(imageCount) {
		for i := range c.ImageAcquiredSemaphores {
			vk.DestroySemaphore(c.Device, c.ImageAcquiredSemaphores[i], nil)
			vk.DestroySemaphore(c.Device, c.DrawCompleteSemaphores[i], nil)
			if c.PresentQueueIndex != c.GraphicsQueueIndex {
				vk.DestroySemaphore(c.Device, c.ImageOwnershipSemaphores[i], nil)
			}
		}

		semaphoreCreateInfo := &vk.SemaphoreCreateInfo{
			SType: vk.StructureTypeSemaphoreCreateInfo,
		}
		c.ImageAcquiredSemaphores = make([]vk.Semaphore, imageCount)
		c.DrawCompleteSemaphores = make([]vk.Semaphore, imageCount)
		c.ImageOwnershipSemaphores = make([]vk.Semaphore, imageCount)
		for i := uint32(0); i < imageCount; i++ {
			ret := vk.CreateSemaphore(c.Device, semaphoreCreateInfo, nil, &c.ImageAcquiredSemaphores[i])
			orPanic(NewError(ret))
			ret = vk.CreateSemaphore(c.Device, semaphoreCreateInfo, nil, &c.DrawCompleteSemaphores[i])
			orPanic(NewError(ret))
			if c.PresentQueueIndex != c.GraphicsQueueIndex {
				ret = vk.CreateSemaphore(c.Device, semaphoreCreateInfo, nil, &c.ImageOwnershipSemaphores[i])
				orPanic(NewError(ret))
			}
		}
	}

	for i := 0; i < len(c.SwapchainImageResources); i++ {
		c.SwapchainImageResources[i].Destroy(c.Device, c.CmdPool)
	}
	c.SwapchainImageResources = make([]*backend.SwapchainImageResources, 0, imageCount)
	for i := 0; i < len(swapchainImages); i++ {
		c.SwapchainImageResources = append(c.SwapchainImageResources, &backend.SwapchainImageResources{
			Image: swapchainImages[i],
		})
	}
}

func (c *VulkanContext) Prepare() {
	vk.DeviceWaitIdle(c.Device)

	var cmdPool vk.CommandPool
	ret := vk.CreateCommandPool(c.Device, &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: c.GraphicsQueueIndex,
	}, nil, &cmdPool)
	orPanic(NewError(ret))
	c.CmdPool = cmdPool

	for i := 0; i < len(c.SwapchainImageResources); i++ {
		var cmd = make([]vk.CommandBuffer, 1)
		vk.AllocateCommandBuffers(c.Device, &vk.CommandBufferAllocateInfo{
			SType:              vk.StructureTypeCommandBufferAllocateInfo,
			CommandPool:        c.CmdPool,
			Level:              vk.CommandBufferLevelPrimary,
			CommandBufferCount: 1,
		}, cmd)
		c.SwapchainImageResources[i].Cmd = cmd[0]
	}

	if c.PresentQueueIndex != c.GraphicsQueueIndex {
		var presentCmdPool vk.CommandPool
		ret = vk.CreateCommandPool(c.Device, &vk.CommandPoolCreateInfo{
			SType:            vk.StructureTypeCommandPoolCreateInfo,
			QueueFamilyIndex: c.PresentQueueIndex,
		}, nil, &presentCmdPool)
		orPanic(NewError(ret))
		c.PresentCmdPool = presentCmdPool

		for i := 0; i < len(c.SwapchainImageResources); i++ {
			var cmd = make([]vk.CommandBuffer, 1)
			ret = vk.AllocateCommandBuffers(c.Device, &vk.CommandBufferAllocateInfo{
				SType:              vk.StructureTypeCommandBufferAllocateInfo,
				CommandPool:        c.PresentCmdPool,
				Level:              vk.CommandBufferLevelPrimary,
				CommandBufferCount: 1,
			}, cmd)
			orPanic(NewError(ret))
			c.SwapchainImageResources[i].GraphicsToPresentCmd = cmd[0]

			c.SwapchainImageResources[i].SetImageOwnership(c.GraphicsQueueIndex, c.PresentQueueIndex)
		}
	}

	for i := 0; i < len(c.SwapchainImageResources); i++ {
		var view vk.ImageView
		ret = vk.CreateImageView(c.Device, &vk.ImageViewCreateInfo{
			SType:  vk.StructureTypeImageViewCreateInfo,
			Format: c.SwapchainDimensions.Format,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleR,
				G: vk.ComponentSwizzleG,
				B: vk.ComponentSwizzleB,
				A: vk.ComponentSwizzleA,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1,
				LayerCount: 1,
			},
			ViewType: vk.ImageViewType2d,
			Image:    c.SwapchainImageResources[i].Image,
		}, nil, &view)
		orPanic(NewError(ret))
		c.SwapchainImageResources[i].View = view
	}
}

func (c *VulkanContext) CurrentFrameFence() vk.Fence {
	return c.FrameFences[c.FrameIndex]
}

func (c *VulkanContext) CurrentImageAcquiredSemaphore(imageIdx int) vk.Semaphore {
	return c.ImageAcquiredSemaphores[imageIdx]
}

func (c *VulkanContext) CurrentDrawCompleteSemaphore(imageIdx int) vk.Semaphore {
	return c.DrawCompleteSemaphores[imageIdx]
}

func (c *VulkanContext) AcquireNextImage() (imageIndex int, outdated bool, err error) {
	defer checkErr(&err)

	ret := vk.WaitForFences(c.Device, 1, []vk.Fence{c.FrameFences[c.FrameIndex]}, vk.True, vk.MaxUint64)
	orPanic(NewError(ret))

	var idx uint32
	// We don't know the image index yet, so we must use a separate set of acquisition semaphores
	// or just a single one if we only have one frame in flight?
	// Actually, the common pattern is to have one acquisition semaphore per FRAME in flight.
	// But the validation error said: "Use a separate semaphore per swapchain image. Index these semaphores using the index of the acquired image."
	// That's for the SIGNALING of the semaphore during AcquireNextImage.

	// Wait, I should have a separate pool of semaphores for acquisition.
	// Let's stick to the recommendation: per-image semaphores.
	// But how do we know which semaphore to pass to AcquireNextImage if we don't know the index yet?
	// Usually you have a "next" available semaphore from a pool.

	// Actually, the validation error is about signaled semaphores in QueueSubmit.
	// "Swapchain image 3 was presented but was not re-acquired, so VkSemaphore 0x40000000004 may still be in use and cannot be safely reused with image index 0."
	// This confirms we should use image index for the DrawComplete/Ownership semaphores.
	// For acquisition, we can use FrameIndex because we wait for the FENCE of that frame.

	ret = vk.AcquireNextImage(c.Device, c.Swapchain, vk.MaxUint64,
		c.ImageAcquiredSemaphores[c.FrameIndex], vk.NullFence, &idx)
	imageIndex = int(idx)

	switch ret {
	case vk.ErrorOutOfDate:
		c.FrameIndex = (c.FrameIndex + 1) % c.FrameLag
		c.PrepareSwapchain(c.SwapchainDimensions)
		c.Prepare()
		outdated = true
		return
	case vk.Suboptimal, vk.Success:
	default:
		orPanic(NewError(ret))
	}

	vk.ResetFences(c.Device, 1, []vk.Fence{c.FrameFences[c.FrameIndex]})

	return
}

func (c *VulkanContext) PresentImage(imageIdx int) (outdated bool, err error) {
	var semaphore vk.Semaphore
	if c.PresentQueueIndex != c.GraphicsQueueIndex {
		semaphore = c.ImageOwnershipSemaphores[imageIdx]
	} else {
		semaphore = c.DrawCompleteSemaphores[imageIdx]
	}

	ret := vk.QueuePresent(c.PresentQueue, &vk.PresentInfo{
		SType:              vk.StructureTypePresentInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    []vk.Semaphore{semaphore},
		SwapchainCount:     1,
		PSwapchains:        []vk.Swapchain{c.Swapchain},
		PImageIndices:      []uint32{uint32(imageIdx)},
	})
	c.FrameIndex = (c.FrameIndex + 1) % c.FrameLag

	switch ret {
	case vk.ErrorOutOfDate:
		outdated = true
		return
	case vk.Suboptimal, vk.Success:
		return
	default:
		err = NewError(ret)
		return
	}
}

func (c *VulkanContext) Submit(imageIdx int) error {
	submitInfo := vk.SubmitInfo{
		SType: vk.StructureTypeSubmitInfo,
		PWaitDstStageMask: []vk.PipelineStageFlags{
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		},
		WaitSemaphoreCount: 1,
		PWaitSemaphores: []vk.Semaphore{
			c.ImageAcquiredSemaphores[c.FrameIndex],
		},
		SignalSemaphoreCount: 1,
		PSignalSemaphores: []vk.Semaphore{
			c.DrawCompleteSemaphores[imageIdx],
		},
	}

	ret := vk.QueueSubmit(c.GraphicsQueue, 1, []vk.SubmitInfo{submitInfo}, c.FrameFences[c.FrameIndex])
	return NewError(ret)
}

func dbgCallbackFunc(flags vk.DebugReportFlags, objectType vk.DebugReportObjectType,
	object uint64, location uint64, messageCode int32, pLayerPrefix string,
	pMessage string, pUserData unsafe.Pointer) vk.Bool32 {

	switch {
	case flags&vk.DebugReportFlags(vk.DebugReportInformationBit) != 0:
		log.Printf("INFORMATION: [%s] Code %d : %s", pLayerPrefix, messageCode, pMessage)
	case flags&vk.DebugReportFlags(vk.DebugReportWarningBit) != 0:
		log.Printf("WARNING: [%s] Code %d : %s", pLayerPrefix, messageCode, pMessage)
	case flags&vk.DebugReportFlags(vk.DebugReportPerformanceWarningBit) != 0:
		log.Printf("PERFORMANCE WARNING: [%s] Code %d : %s", pLayerPrefix, messageCode, pMessage)
	case flags&vk.DebugReportFlags(vk.DebugReportErrorBit) != 0:
		log.Printf("ERROR: [%s] Code %d : %s", pLayerPrefix, messageCode, pMessage)
	case flags&vk.DebugReportFlags(vk.DebugReportDebugBit) != 0:
		log.Printf("DEBUG: [%s] Code %d : %s", pLayerPrefix, messageCode, pMessage)
	default:
		log.Printf("INFORMATION: [%s] Code %d : %s", pLayerPrefix, messageCode, pMessage)
	}
	return vk.Bool32(vk.False)
}
