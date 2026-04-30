package renderer

import (
	"fmt"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	vk "github.com/goki/vulkan"
)

// GpuSurface is a Vulkan-side render target that corresponds to a single
// GPU-address-identified framebuffer surface registered by the game.
type GpuSurface struct {
	GPUAddress uintptr
	Width      uint32
	Height     uint32
	Format     vk.Format
	TextureId  imgui.TextureRef

	// Vulkan objects.
	image       vk.Image
	imageMem    vk.DeviceMemory
	imageView   vk.ImageView
	sampler     vk.Sampler
	framebuffer vk.Framebuffer
	renderPass  vk.RenderPass

	// firstUse tracks whether the image has been transitioned from UNDEFINED.
	firstUse bool
}

// Destroy frees all Vulkan resources owned by this surface.
func (s *GpuSurface) Destroy(dev vk.Device) {
	if s.framebuffer != vk.NullFramebuffer {
		vk.DestroyFramebuffer(dev, s.framebuffer, nil)
	}
	if s.renderPass != vk.NullRenderPass {
		vk.DestroyRenderPass(dev, s.renderPass, nil)
	}
	if s.imageView != vk.NullImageView {
		vk.DestroyImageView(dev, s.imageView, nil)
	}
	if s.image != vk.NullImage {
		vk.DestroyImage(dev, s.image, nil)
	}
	if s.imageMem != vk.NullDeviceMemory {
		vk.FreeMemory(dev, s.imageMem, nil)
	}
}

func (t *GpuTranslator) GetSurface(address uintptr, width, height uint32) (imgui.TextureRef, error) {
	// Check if it already exists.
	t.surfacesMutex.Lock()
	surface, ok := t.surfaces[address]
	t.surfacesMutex.Unlock()
	if ok {
		return surface.TextureId, nil
	}

	// Create the surface.
	surface = &GpuSurface{
		GPUAddress: address,
		Width:      width,
		Height:     height,
		Format:     vk.FormatR8g8b8a8Unorm,
		firstUse:   true,
	}
	if err := t.allocSurface(surface); err != nil {
		return imgui.TextureRef{}, fmt.Errorf("RegisterSurface 0x%X: %w", address, err)
	}
	surface.TextureId = t.backend.CreateVulkanTexture(surface.sampler, surface.imageView, vk.ImageLayoutShaderReadOnlyOptimal)
	t.surfacesMutex.Lock()
	t.surfaces[address] = surface
	t.surfacesMutex.Unlock()

	return surface.TextureId, nil
}

// GetSurfaceImageView returns the VkImageView for a registered surface so the renderer can display it as a texture.
// Returns vk.NullImageView if unknown.
func (t *GpuTranslator) GetSurfaceImageView(gpuAddress uintptr) vk.ImageView {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	if s, ok := t.surfaces[gpuAddress]; ok {
		return s.imageView
	}
	return vk.NullImageView
}

func (t *GpuTranslator) allocSurface(s *GpuSurface) error {
	// Create the render-target image.
	var image vk.Image
	result := vk.CreateImage(t.handles.Device, &vk.ImageCreateInfo{
		SType:         vk.StructureTypeImageCreateInfo,
		ImageType:     vk.ImageType2d,
		Format:        s.Format,
		Extent:        vk.Extent3D{Width: s.Width, Height: s.Height, Depth: 1},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        vk.ImageTilingOptimal,
		Usage:         vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageTransferSrcBit),
		SharingMode:   vk.SharingModeExclusive,
		InitialLayout: vk.ImageLayoutUndefined,
	}, nil, &image)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateImage: %w", err)
	}
	s.image = image

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(t.handles.Device, s.image, &memReqs)
	memReqs.Deref()

	var imageMem vk.DeviceMemory
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyDeviceLocalBit),
	}, nil, &imageMem)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkAllocateMemory: %w", err)
	}
	s.imageMem = imageMem
	vk.BindImageMemory(t.handles.Device, s.image, s.imageMem, 0)

	var imageView vk.ImageView
	result = vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    s.image,
		ViewType: vk.ImageViewType2d,
		Format:   s.Format,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			LevelCount: 1,
			LayerCount: 1,
		},
	}, nil, &imageView)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateImageView: %w", err)
	}
	s.imageView = imageView

	var sampler vk.Sampler
	result = vk.CreateSampler(t.handles.Device, &vk.SamplerCreateInfo{
		SType:        vk.StructureTypeSamplerCreateInfo,
		MagFilter:    vk.FilterNearest,
		MinFilter:    vk.FilterNearest,
		AddressModeU: vk.SamplerAddressModeClampToEdge,
		AddressModeV: vk.SamplerAddressModeClampToEdge,
		AddressModeW: vk.SamplerAddressModeClampToEdge,
	}, nil, &sampler)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateSampler: %w", err)
	}
	s.sampler = sampler

	var renderPass vk.RenderPass
	result = vk.CreateRenderPass(t.handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments: []vk.AttachmentDescription{{
			Format:         s.Format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutColorAttachmentOptimal,
			FinalLayout:    vk.ImageLayoutShaderReadOnlyOptimal,
		}},
		SubpassCount: 1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
		}},
	}, nil, &renderPass)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateRenderPass: %w", err)
	}
	s.renderPass = renderPass

	var framebuffer vk.Framebuffer
	result = vk.CreateFramebuffer(t.handles.Device, &vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		RenderPass:      s.renderPass,
		AttachmentCount: 1,
		PAttachments:    []vk.ImageView{s.imageView},
		Width:           s.Width,
		Height:          s.Height,
		Layers:          1,
	}, nil, &framebuffer)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateFramebuffer: %w", err)
	}
	s.framebuffer = framebuffer

	return nil
}
