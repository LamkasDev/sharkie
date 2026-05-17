package vulkan

import (
	"fmt"
	"hash/fnv"
	"runtime"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetImageView(descriptor spirvStructs.ImageDescriptor) (vk.ImageView, bool) {
	// Get already created image view.
	t.imagesMutex.Lock()
	defer t.imagesMutex.Unlock()
	viewKey := spirvStructs.NewImageIdentityKey(descriptor)
	if view, ok := t.imageViews[viewKey]; ok {
		return view, false
	}

	// Create the image view.
	format, _ := TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
	if format == vk.FormatUndefined {
		return vk.NullImageView, false
	}

	image, exists := t.images[descriptor.BaseAddress]
	isNewImage := !exists
	if !exists {
		// Calculate max mip levels.
		maxDim := descriptor.Width
		if descriptor.Height > maxDim {
			maxDim = descriptor.Height
		}
		mipLevels := uint32(1)
		for maxDim > 1 {
			maxDim /= 2
			mipLevels++
		}
		if mipLevels > 16 {
			mipLevels = 16
		}

		result := vk.CreateImage(t.handles.Device, &vk.ImageCreateInfo{
			SType:       vk.StructureTypeImageCreateInfo,
			ImageType:   vk.ImageType2d,
			Format:      format,
			Extent:      vk.Extent3D{Width: descriptor.Width, Height: descriptor.Height, Depth: 1},
			MipLevels:   mipLevels,
			ArrayLayers: 1,
			Samples:     vk.SampleCount1Bit,
			Tiling:      vk.ImageTilingOptimal,
			Usage:       vk.ImageUsageFlags(vk.ImageUsageSampledBit | vk.ImageUsageStorageBit | vk.ImageUsageTransferDstBit | vk.ImageUsageTransferSrcBit),
		}, nil, &image)
		if err := as.NewError(result); err != nil {
			logger.Printf("GetImageView: vkCreateImage failed: %v", err)
			return vk.NullImageView, false
		}

		var memReqs vk.MemoryRequirements
		vk.GetImageMemoryRequirements(t.handles.Device, image, &memReqs)
		memReqs.Deref()

		var imageMem vk.DeviceMemory
		result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
			SType:           vk.StructureTypeMemoryAllocateInfo,
			AllocationSize:  memReqs.Size,
			MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyDeviceLocalBit),
		}, nil, &imageMem)
		if err := as.NewError(result); err != nil {
			logger.Printf("GetImageView: vkAllocateMemory failed: %v", err)
			vk.DestroyImage(t.handles.Device, image, nil)
			return vk.NullImageView, false
		}
		vk.BindImageMemory(t.handles.Device, image, imageMem, 0)
		t.images[descriptor.BaseAddress] = image
		t.imageMems[descriptor.BaseAddress] = imageMem

		// Allocate command buffer and transition to general layout.
		cb := t.AllocateCommandBuffer()
		vk.BeginCommandBuffer(cb, &vk.CommandBufferBeginInfo{
			SType: vk.StructureTypeCommandBufferBeginInfo,
			Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
		})
		t.imageBarrier(cb, image,
			vk.ImageLayoutUndefined, vk.ImageLayoutGeneral,
			0, vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
			vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit))
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
		if err := as.NewError(result); err != nil {
			return vk.NullImageView, false
		}
		t.WaitOnWorkerFence()
	}

	var view vk.ImageView
	baseMipLevel := uint32(descriptor.BaseLevel)
	if baseMipLevel > 15 {
		baseMipLevel = 15
	}
	result := vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    image,
		ViewType: vk.ImageViewType2d,
		Format:   format,
		Components: vk.ComponentMapping{
			R: translateDstSelToVkSwizzle(descriptor.DstSelX),
			G: translateDstSelToVkSwizzle(descriptor.DstSelY),
			B: translateDstSelToVkSwizzle(descriptor.DstSelZ),
			A: translateDstSelToVkSwizzle(descriptor.DstSelW),
		},
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:   vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel: baseMipLevel,
			LevelCount:   1,
			LayerCount:   1,
		},
	}, nil, &view)
	if err := as.NewError(result); err != nil {
		logger.Printf("GetImageView: vkCreateImageView failed: %v", err)
		return vk.NullImageView, false
	}
	t.imageViews[viewKey] = view
	t.imageDescriptors[descriptor.BaseAddress] = descriptor

	return view, isNewImage
}

func (t *GpuTranslator) GetSampler(descriptor spirvStructs.SamplerDescriptor) vk.Sampler {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%v", descriptor)))
	hash := h.Sum64()

	// Get already created sampler.
	t.samplersMutex.Lock()
	defer t.samplersMutex.Unlock()
	if s, ok := t.samplers[hash]; ok {
		return s
	}

	// Create the sampler.
	var sampler vk.Sampler
	vk.CreateSampler(t.handles.Device, &vk.SamplerCreateInfo{
		SType:        vk.StructureTypeSamplerCreateInfo,
		MagFilter:    translateFilter(descriptor.MagFilter),
		MinFilter:    translateFilter(descriptor.MinFilter),
		MipmapMode:   translateMipmapMode(descriptor.MipFilter),
		AddressModeU: translateClampMode(descriptor.ClampX),
		AddressModeV: translateClampMode(descriptor.ClampY),
		AddressModeW: translateClampMode(descriptor.ClampZ),
		MinLod:       0,
		MaxLod:       15,
	}, nil, &sampler)

	t.samplers[hash] = sampler

	return sampler
}

func (t *GpuTranslator) uploadBufferDataToImage(address uintptr, image vk.Image, width, height, pitch, bpp uint32, mipLevel uint32) bool {
	srcBuffer, srcOffset, err := t.GetBufferFromAddress(address)
	if err != nil {
		return true
	}

	// Allocate command buffer and copy data.
	cb := t.AllocateCommandBuffer()
	vk.BeginCommandBuffer(cb, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	vk.CmdCopyBufferToImage(cb, srcBuffer, image, vk.ImageLayoutGeneral, 1, []vk.BufferImageCopy{{
		BufferOffset:      vk.DeviceSize(srcOffset),
		BufferRowLength:   pitch,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:   mipLevel,
			LayerCount: 1,
		},
		ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
		ImageExtent: vk.Extent3D{Width: width, Height: height, Depth: 1},
	}})
	vk.EndCommandBuffer(cb)
	defer t.FreeCommandBuffer(cb)

	// Create a local fence to avoid threading issues during concurrent uploads.
	var fence vk.Fence
	vk.CreateFence(t.handles.Device, &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
	}, nil, &fence)
	defer vk.DestroyFence(t.handles.Device, fence, nil)

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

	t.QueueMutex.Lock()
	result := vk.QueueSubmit(t.handles.GraphicsQueue, 1, submitInfos, fence)
	t.QueueMutex.Unlock()
	if err = as.NewError(result); err != nil {
		return false
	}
	vk.WaitForFences(t.handles.Device, 1, []vk.Fence{fence}, vk.True, ^uint64(0))

	return true
}

func translateDstSelToVkSwizzle(sel uint8) vk.ComponentSwizzle {
	switch sel {
	case 0:
		return vk.ComponentSwizzleZero
	case 1:
		return vk.ComponentSwizzleOne
	case 4:
		return vk.ComponentSwizzleR
	case 5:
		return vk.ComponentSwizzleG
	case 6:
		return vk.ComponentSwizzleB
	case 7:
		return vk.ComponentSwizzleA
	default:
		return vk.ComponentSwizzleIdentity
	}
}
