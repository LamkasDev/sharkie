package vulkan

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"runtime"
	"unsafe"

	as "github.com/LamkasDev/asche"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetImageView(descriptor spirvStructs.ImageDescriptor) (vk.ImageView, error, bool) {
	hash := descriptor.Hash()

	// Get already created image view.
	t.imagesMutex.Lock()
	defer t.imagesMutex.Unlock()
	if view, ok := t.imageViews[hash]; ok {
		return view, nil, false
	}

	// Create a new image if needed.
	format, _ := TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
	if format == vk.FormatUndefined {
		return vk.NullImageView, fmt.Errorf("invalid format"), false
	}
	image, imageExists := t.images[descriptor.BaseAddress]
	if !imageExists {
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

		// Create the image.
		result := vk.CreateImage(t.handles.Device, &vk.ImageCreateInfo{
			SType:     vk.StructureTypeImageCreateInfo,
			ImageType: vk.ImageType2d,
			Format:    format,
			Extent: vk.Extent3D{
				Width:  uint32(descriptor.Width),
				Height: uint32(descriptor.Height),
				Depth:  uint32(descriptor.Depth),
			},
			MipLevels:   mipLevels,
			ArrayLayers: 1,
			Samples:     vk.SampleCount1Bit,
			Tiling:      vk.ImageTilingOptimal,
			Usage:       vk.ImageUsageFlags(vk.ImageUsageSampledBit | vk.ImageUsageStorageBit | vk.ImageUsageTransferDstBit | vk.ImageUsageTransferSrcBit),
		}, nil, &image)
		if err := as.NewError(result); err != nil {
			return vk.NullImageView, err, false
		}

		// Allocate image memory.
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
			return vk.NullImageView, err, false
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
			return vk.NullImageView, err, false
		}
		t.WaitOnWorkerFence()
	}

	// Create the image view.
	var view vk.ImageView
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
			BaseMipLevel: min(uint32(descriptor.BaseLevel), 15),
			LevelCount:   1,
			LayerCount:   1,
		},
	}, nil, &view)
	if err := as.NewError(result); err != nil {
		return vk.NullImageView, err, false
	}
	t.imageViews[hash] = view
	t.imageDescriptors[hash] = descriptor

	return view, nil, !imageExists
}

func (t *GpuTranslator) GetSampler(descriptor spirvStructs.SamplerDescriptor) (vk.Sampler, error) {
	hash := descriptor.Hash()

	// Get already created sampler.
	t.samplersMutex.Lock()
	defer t.samplersMutex.Unlock()
	if sampler, ok := t.samplers[hash]; ok {
		return sampler, nil
	}

	// Create the sampler.
	var sampler vk.Sampler
	result := vk.CreateSampler(t.handles.Device, &vk.SamplerCreateInfo{
		SType:        vk.StructureTypeSamplerCreateInfo,
		MagFilter:    translateFilter(descriptor.XyMagFilter),
		MinFilter:    translateFilter(descriptor.XyMinFilter),
		MipmapMode:   translateMipmapMode(descriptor.MipFilter),
		AddressModeU: translateClampMode(descriptor.ClampX),
		AddressModeV: translateClampMode(descriptor.ClampY),
		AddressModeW: translateClampMode(descriptor.ClampZ),
		MinLod:       descriptor.MinLod,
		MaxLod:       descriptor.MaxLod,
		BorderColor:  translateBorderColorType(descriptor.BorderColorType),
	}, nil, &sampler)
	if err := as.NewError(result); err != nil {
		return vk.NullSampler, err
	}
	t.samplers[hash] = sampler

	return sampler, nil
}

func (t *GpuTranslator) uploadBufferDataToImage(descriptor spirvStructs.ImageDescriptor, image vk.Image) bool {
	srcBuffer, srcOffset, err := t.GetBufferFromAddress(descriptor.BaseAddress)
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
		BufferRowLength:   uint32(descriptor.Pitch),
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:   uint32(descriptor.BaseLevel),
			LayerCount: 1,
		},
		ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
		ImageExtent: vk.Extent3D{
			Width:  uint32(descriptor.Width),
			Height: uint32(descriptor.Height),
			Depth:  uint32(descriptor.Depth),
		},
	}})
	t.imageBarrier(cb, image,
		vk.ImageLayoutGeneral, vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessTransferWriteBit), vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit))
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

	t.dumpImage(descriptor)

	return true
}

func (t *GpuTranslator) dumpImage(descriptor spirvStructs.ImageDescriptor) {
	_, bpp := TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
	if bpp != 4 {
		return
	}

	width := int(descriptor.Width)
	height := int(descriptor.Height)
	pitch := int(descriptor.Pitch)
	if pitch == 0 {
		pitch = width
	}

	// Use pitch to ensure we read enough data.
	dataSize := pitch * height * int(bpp)
	inOnion := descriptor.BaseAddress >= GlobalAllocator.Base && descriptor.BaseAddress < GlobalAllocator.Base+uintptr(GlobalAllocator.Size)
	inGarlic := descriptor.BaseAddress >= GlobalGpuAllocator.Base && descriptor.BaseAddress < GlobalGpuAllocator.Base+uintptr(GlobalGpuAllocator.Size)
	if !inOnion && !inGarlic {
		return
	}
	srcData := unsafe.Slice((*byte)(unsafe.Pointer(descriptor.BaseAddress)), dataSize)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*pitch + x) * int(bpp)
			if offset+3 >= len(srcData) {
				continue
			}

			// We assume A8R8G8B8 little-endian which is [B, G, R, A] in memory.
			img.SetRGBA(x, y, color.RGBA{
				R: srcData[offset+2],
				G: srcData[offset+1],
				B: srcData[offset+0],
				A: srcData[offset+3],
			})
		}
	}

	f, err := os.Create(fmt.Sprintf("temp/images/image_0x%X_%dx%d_t%d.png", descriptor.BaseAddress, width, height, descriptor.TilingIndex))
	if err != nil {
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Printf("failed to encode image: %v\n", err)
	}
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
