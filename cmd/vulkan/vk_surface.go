package vulkan

import (
	"fmt"

	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
)

// VulkanSurface is a guest-address-identified render target and its Vulkan resources.
type VulkanSurface struct {
	Address   uintptr
	ImageView *VulkanImageView
	Sampler   vk.Sampler
	TextureId imgui.TextureRef

	// FirstUse tracks whether the image has been transitioned from UNDEFINED.
	FirstUse bool

	// ContentValid tracks whether the surface has valid content.
	ContentValid bool

	// FrameUsed tracks the last frame this surface was used in.
	FrameUsed uint64
}

type VulkanSurfaceRequest struct {
	Descriptor spirvStructs.ImageDescriptor
	CompSwap   uint32
	Image      *VulkanImage
}

func CreateSurface(handles *VulkanHandles, req VulkanSurfaceRequest) (*VulkanSurface, error) {
	surface := &VulkanSurface{
		ImageView: &VulkanImageView{},
	}

	imageView, err := CreateImageView(handles, VulkanImageViewRequest{
		Image:      req.Image,
		Descriptor: req.Descriptor,
		IsSurface:  true,
	})
	if err != nil {
		return nil, err
	}
	surface.ImageView = imageView

	var sampler vk.Sampler
	result := vk.CreateSampler(handles.Device, &vk.SamplerCreateInfo{
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
	surface.Sampler = sampler

	return surface, nil
}

func (s *VulkanSurface) Destroy(device vk.Device) {
	if s.ImageView != nil && s.ImageView.Image != nil {
		s.ImageView.Image.Destroy(device)
	}
	if s.ImageView != nil {
		s.ImageView.Destroy(device)
	}
	if s.Sampler != vk.NullSampler {
		vk.DestroySampler(device, s.Sampler, nil)
		s.Sampler = vk.NullSampler
	}
}
