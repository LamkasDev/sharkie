package renderer

import (
	"fmt"
	"hash/fnv"

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
	vk.CreateImage(t.handles.Device, &vk.ImageCreateInfo{
		SType:       vk.StructureTypeImageCreateInfo,
		ImageType:   vk.ImageType2d,
		Format:      format,
		Extent:      vk.Extent3D{Width: d.Width, Height: d.Height, Depth: 1},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingLinear, // Use linear for direct guest memory mapping?
		Usage:       vk.ImageUsageFlags(vk.ImageUsageSampledBit),
	}, nil, &image)

	// Since we can't easily map a VkImage to a VkBuffer address in standard Vulkan,
	// for now we'll just return Null until we implement a proper texture uploader.
	// But let's at least keep the image handle if we were to create one.
	// t.images[d.BaseAddress] = image

	return vk.NullImageView
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

func (t *GpuTranslator) imageBarrier(commandBuffer vk.CommandBuffer, image vk.Image,
	oldLayout, newLayout vk.ImageLayout,
	srcAccess, dstAccess vk.AccessFlags,
	srcStage, dstStage vk.PipelineStageFlags,
) {
	vk.CmdPipelineBarrier(commandBuffer,
		srcStage, dstStage,
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           oldLayout,
			NewLayout:           newLayout,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               image,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     vk.RemainingMipLevels,
				BaseArrayLayer: 0,
				LayerCount:     vk.RemainingArrayLayers,
			},
			SrcAccessMask: srcAccess,
			DstAccessMask: dstAccess,
		}},
	)
}
