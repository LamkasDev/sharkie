package vulkan

import (
	"fmt"
	"hash/fnv"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetImageView(d spirv.ImageDescriptor) vk.ImageView {
	// Check if it's a render target surface.
	t.surfacesMutex.Lock()
	if s, ok := t.surfaces[d.BaseAddress]; ok {
		t.surfacesMutex.Unlock()
		return s.imageView
	}
	t.surfacesMutex.Unlock()

	// Get already created image view.
	t.imagesMutex.Lock()
	defer t.imagesMutex.Unlock()
	if view, ok := t.imageViews[d.BaseAddress]; ok {
		return view
	}

	// Create the image view.
	format, _ := TranslateGcnFormat(d.DataFormat, d.NumFormat)
	if format == vk.FormatUndefined {
		return vk.NullImageView
	}

	var image vk.Image
	result := vk.CreateImage(t.handles.Device, &vk.ImageCreateInfo{
		SType:       vk.StructureTypeImageCreateInfo,
		ImageType:   vk.ImageType2d,
		Format:      format,
		Extent:      vk.Extent3D{Width: d.Width, Height: d.Height, Depth: 1},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage:       vk.ImageUsageFlags(vk.ImageUsageSampledBit | vk.ImageUsageTransferDstBit),
	}, nil, &image)
	if err := as.NewError(result); err != nil {
		return vk.NullImageView
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
		vk.DestroyImage(t.handles.Device, image, nil)
		return vk.NullImageView
	}
	vk.BindImageMemory(t.handles.Device, image, imageMem, 0)

	var view vk.ImageView
	result = vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    image,
		ViewType: vk.ImageViewType2d,
		Format:   format,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			LevelCount: 1,
			LayerCount: 1,
		},
	}, nil, &view)
	if err := as.NewError(result); err != nil {
		vk.FreeMemory(t.handles.Device, imageMem, nil)
		vk.DestroyImage(t.handles.Device, image, nil)
		return vk.NullImageView
	}

	t.images[d.BaseAddress] = image
	t.imageViews[d.BaseAddress] = view
	t.imageMems[d.BaseAddress] = imageMem

	return view
}

func (t *GpuTranslator) GetSampler(d spirv.SamplerDescriptor) vk.Sampler {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%v", d)))
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
		MagFilter:    translateFilter(d.MagFilter),
		MinFilter:    translateFilter(d.MinFilter),
		MipmapMode:   translateMipmapMode(d.MipFilter),
		AddressModeU: translateClampMode(d.ClampX),
		AddressModeV: translateClampMode(d.ClampY),
		AddressModeW: translateClampMode(d.ClampZ),
		MinLod:       0,
		MaxLod:       15,
	}, nil, &sampler)
	t.samplers[hash] = sampler

	return sampler
}

func (t *GpuTranslator) uploadBufferDataToImage(address uintptr, image vk.Image, width, height, bpp uint32) bool {
	srcBuffer, srcOffset, err := t.GetBufferFromAddress(address)
	if err != nil {
		return true
	}

	cb := t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(cb, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})

	// Transition Undefined -> TransferDst (or ShaderReadOnly if no copy)
	targetLayout := vk.ImageLayoutTransferDstOptimal
	if err != nil {
		targetLayout = vk.ImageLayoutShaderReadOnlyOptimal
	}

	t.imageBarrier(cb, image,
		vk.ImageLayoutUndefined, targetLayout,
		0, vk.AccessFlags(vk.AccessTransferWriteBit|vk.AccessShaderReadBit),
		vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit|vk.PipelineStageAllGraphicsBit))

	// Copy if we have a source.
	if err == nil {
		vk.CmdCopyBufferToImage(cb, srcBuffer, image, vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{{
			BufferOffset:      vk.DeviceSize(srcOffset),
			BufferRowLength:   0,
			BufferImageHeight: 0,
			ImageSubresource: vk.ImageSubresourceLayers{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LayerCount: 1,
			},
			ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
			ImageExtent: vk.Extent3D{Width: width, Height: height, Depth: 1},
		}})

		// Transition TransferDst -> ShaderReadOnly
		t.imageBarrier(cb, image,
			vk.ImageLayoutTransferDstOptimal, vk.ImageLayoutShaderReadOnlyOptimal,
			vk.AccessFlags(vk.AccessTransferWriteBit), vk.AccessFlags(vk.AccessShaderReadBit),
			vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit))
	}

	vk.EndCommandBuffer(cb)

	// Submit and wait for completion (simple but slow for now).
	vk.QueueSubmit(t.handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cb},
	}}, vk.NullFence)
	vk.QueueWaitIdle(t.handles.GraphicsQueue)

	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{cb})
	return true
}
