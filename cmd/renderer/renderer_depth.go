package renderer

import (
	as "github.com/LamkasDev/asche"
	vk "github.com/goki/vulkan"
)

type Depth struct {
	format   vk.Format
	image    vk.Image
	memAlloc *vk.MemoryAllocateInfo
	mem      vk.DeviceMemory
	view     vk.ImageView
}

func NewDepth(r *Renderer) *Depth {
	depth := &Depth{
		format: vk.FormatD16Unorm,
	}
	result := vk.CreateImage(r.Handles.Device, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    depth.format,
		Extent: vk.Extent3D{
			Width:  r.SwapchainDimensions.Width,
			Height: r.SwapchainDimensions.Height,
			Depth:  1,
		},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage:       vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
	}, nil, &depth.image)
	if err := as.NewError(result); err != nil {
		panic(err)
	}

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(r.Handles.Device, depth.image, &memReqs)
	memReqs.Deref()

	memTypeIndex, _ := as.FindRequiredMemoryTypeFallback(r.Handles.MemoryProperties,
		vk.MemoryPropertyFlagBits(memReqs.MemoryTypeBits), vk.MemoryPropertyDeviceLocalBit)
	depth.memAlloc = &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var mem vk.DeviceMemory
	result = vk.AllocateMemory(r.Handles.Device, depth.memAlloc, nil, &mem)
	if err := as.NewError(result); err != nil {
		panic(err)
	}
	depth.mem = mem

	result = vk.BindImageMemory(r.Handles.Device, depth.image, depth.mem, 0)
	if err := as.NewError(result); err != nil {
		panic(err)
	}

	var view vk.ImageView
	result = vk.CreateImageView(r.Handles.Device, &vk.ImageViewCreateInfo{
		SType:  vk.StructureTypeImageViewCreateInfo,
		Format: depth.format,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			LevelCount: 1,
			LayerCount: 1,
		},
		ViewType: vk.ImageViewType2d,
		Image:    depth.image,
	}, nil, &view)
	if err := as.NewError(result); err != nil {
		panic(err)
	}
	depth.view = view

	return depth
}

func (depth *Depth) Destroy(r *Renderer) {
	vk.DestroyImageView(r.Handles.Device, depth.view, nil)
	vk.DestroyImage(r.Handles.Device, depth.image, nil)
	vk.FreeMemory(r.Handles.Device, depth.mem, nil)
}
