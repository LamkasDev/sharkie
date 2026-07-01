package vulkan

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	vk "github.com/goki/vulkan"
)

type VulkanSurface struct {
	image            vk.Image
	imageMem         vk.DeviceMemory
	imageView        vk.ImageView
	storageImageView vk.ImageView
	sampler          vk.Sampler
	format           vk.Format
	width            uint32
	height           uint32
}

func (s *VulkanSurface) DestroyViews(device vk.Device) {
	if s.imageView != vk.NullImageView {
		vk.DestroyImageView(device, s.imageView, nil)
	}
	if s.storageImageView != vk.NullImageView {
		vk.DestroyImageView(device, s.storageImageView, nil)
	}
	if s.sampler != vk.NullSampler {
		vk.DestroySampler(device, s.sampler, nil)
	}
}

type SurfaceRequest struct {
	SurfaceKey
	Format vk.Format
	Width  uint32
	Height uint32
}

type SurfaceKey struct {
	GpuAddress uintptr
}

func (t *GpuTranslator) createSurface(request SurfaceRequest) (VulkanSurface, error) {
	// Reuse existing texture image if one exists at this address, so
	// render-target writes are immediately visible when the same memory
	// is sampled as a texture (unified surface/texture cache pattern).
	if oldImage, ok := t.images[request.GpuAddress]; ok && !IsDepthFormat(request.Format) {
		if logger.LogToFile {
			logger.Printf("CREATESURFACE: reusing texture image at 0x%X\n", request.GpuAddress)
		}
		surface := VulkanSurface{
			image:    oldImage,
			imageMem: t.imageMems[request.GpuAddress],
			format:   request.Format,
			width:    request.Width,
			height:   request.Height,
		}

		// Create image views from the shared image.
		aspect := vk.ImageAspectFlags(vk.ImageAspectColorBit)
		vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
			SType: vk.StructureTypeImageViewCreateInfo, Image: surface.image,
			ViewType: vk.ImageViewType2d, Format: request.Format,
			SubresourceRange: vk.ImageSubresourceRange{AspectMask: aspect, LevelCount: 1, LayerCount: 1},
		}, nil, &surface.imageView)
		vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
			SType: vk.StructureTypeImageViewCreateInfo, Image: surface.image,
			ViewType: vk.ImageViewType2d, Format: request.Format,
			SubresourceRange: vk.ImageSubresourceRange{AspectMask: aspect, LevelCount: 1, LayerCount: 1},
		}, nil, &surface.storageImageView)
		vk.CreateSampler(t.handles.Device, &vk.SamplerCreateInfo{SType: vk.StructureTypeSamplerCreateInfo}, nil, &surface.sampler)

		return surface, nil
	}

	surface := VulkanSurface{
		format: request.Format,
		width:  request.Width,
		height: request.Height,
	}

	// Create the render-target image.
	imageUsage := vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageStorageBit | vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit)
	aspectMask := vk.ImageAspectFlags(vk.ImageAspectColorBit)
	dstLayout := vk.ImageLayoutGeneral
	dstAccess := vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit)
	if IsDepthFormat(request.Format) {
		imageUsage = vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit)
		aspectMask = vk.ImageAspectFlags(vk.ImageAspectDepthBit | vk.ImageAspectStencilBit)
		dstLayout = vk.ImageLayoutDepthStencilAttachmentOptimal
		dstAccess = vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit | vk.AccessDepthStencilAttachmentWriteBit)
	}

	// Calculate max mip levels.
	maxDim := request.Width
	if request.Height > maxDim {
		maxDim = request.Height
	}
	mipLevels := uint32(1)
	for maxDim > 1 {
		maxDim /= 2
		mipLevels++
	}
	if mipLevels > 16 {
		mipLevels = 16
	}

	var image vk.Image
	result := vk.CreateImage(t.handles.Device, &vk.ImageCreateInfo{
		SType:         vk.StructureTypeImageCreateInfo,
		Flags:         vk.ImageCreateFlags(vk.ImageCreateMutableFormatBit),
		ImageType:     vk.ImageType2d,
		Format:        request.Format,
		Extent:        vk.Extent3D{Width: request.Width, Height: request.Height, Depth: 1},
		MipLevels:     mipLevels,
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        vk.ImageTilingOptimal,
		Usage:         imageUsage,
		SharingMode:   vk.SharingModeExclusive,
		InitialLayout: vk.ImageLayoutUndefined,
	}, nil, &image)
	if err := NewError(result); err != nil {
		return surface, fmt.Errorf("vkCreateImage: %w", err)
	}
	SetDebugUtilsObjectName(t.handles.Instance, t.handles.Device, vk.ObjectTypeImage, uint64(uintptr(unsafe.Pointer(image))), fmt.Sprintf("2D Surface Image 0x%X", request.GpuAddress))
	surface.image = image

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(t.handles.Device, surface.image, &memReqs)
	memReqs.Deref()

	var imageMem vk.DeviceMemory
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		PNext:           unsafe.Pointer(new(NewPriorityInfo(0, 1))),
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyDeviceLocalBit),
	}, nil, &imageMem)
	if err := NewError(result); err != nil {
		return surface, fmt.Errorf("vkAllocateMemory: %w", err)
	}
	surface.imageMem = imageMem
	vk.BindImageMemory(t.handles.Device, surface.image, surface.imageMem, 0)
	structs.GlobalMemoryManager.TrackRegion(request.GpuAddress, uintptr(memReqs.Size))

	if !IsDepthFormat(request.Format) {
		t.imagesMutex.Lock()
		t.images[request.GpuAddress] = image
		t.imageMems[request.GpuAddress] = imageMem
		t.imagesMutex.Unlock()
	}

	var imageView vk.ImageView
	result = vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    surface.image,
		ViewType: vk.ImageViewType2d,
		Format:   request.Format,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: aspectMask,
			LevelCount: 1,
			LayerCount: 1,
		},
	}, nil, &imageView)
	if err := NewError(result); err != nil {
		return surface, fmt.Errorf("vkCreateImageView: %w", err)
	}
	surface.imageView = imageView

	if !IsDepthFormat(request.Format) {
		var storageImageView vk.ImageView
		result = vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    surface.image,
			ViewType: vk.ImageViewType2d,
			Format:   request.Format,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: aspectMask,
				LevelCount: 1,
				LayerCount: 1,
			},
		}, nil, &storageImageView)
		if err := NewError(result); err != nil {
			return surface, fmt.Errorf("vkCreateImageView (storage): %w", err)
		}
		surface.storageImageView = storageImageView
	}

	var sampler vk.Sampler
	result = vk.CreateSampler(t.handles.Device, &vk.SamplerCreateInfo{
		SType:        vk.StructureTypeSamplerCreateInfo,
		MagFilter:    vk.FilterNearest,
		MinFilter:    vk.FilterNearest,
		AddressModeU: vk.SamplerAddressModeClampToEdge,
		AddressModeV: vk.SamplerAddressModeClampToEdge,
		AddressModeW: vk.SamplerAddressModeClampToEdge,
	}, nil, &sampler)
	if err := NewError(result); err != nil {
		return surface, fmt.Errorf("vkCreateSampler: %w", err)
	}
	surface.sampler = sampler

	// Transition to a valid initial layout for first use.
	cb := t.AllocateCommandBuffer()
	vk.BeginCommandBuffer(cb, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	t.imageBarrier(cb, surface.image,
		vk.ImageLayoutUndefined, dstLayout,
		0, dstAccess,
		vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit),
		aspectMask)
	vk.EndCommandBuffer(cb)
	defer t.FreeCommandBuffer(cb)

	// Submit and wait for completion.
	commandBuffers := []vk.CommandBuffer{cb}
	submitInfos := []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    commandBuffers,
	}}

	pinner := &runtime.Pinner{}
	pinner.Pin(&commandBuffers)
	pinner.Pin(&submitInfos)
	defer pinner.Unpin()

	t.ResetWorkerFence()
	t.QueueMutex.Lock()
	result = vk.QueueSubmit(t.handles.GraphicsQueue, 1, submitInfos, t.workerFence)
	t.QueueMutex.Unlock()

	if err := NewError(result); err != nil {
		return surface, fmt.Errorf("QueueSubmit: %w", err)
	}
	t.WaitOnWorkerFence()

	return surface, nil
}

func (s *VulkanSurface) Destroy(device vk.Device) {
	if s.imageView != vk.NullImageView {
		vk.DestroyImageView(device, s.imageView, nil)
	}
	if s.storageImageView != vk.NullImageView {
		vk.DestroyImageView(device, s.storageImageView, nil)
	}
	if s.image != vk.NullImage {
		vk.DestroyImage(device, s.image, nil)
	}
	if s.imageMem != vk.NullDeviceMemory {
		vk.FreeMemory(device, s.imageMem, nil)
	}
}
