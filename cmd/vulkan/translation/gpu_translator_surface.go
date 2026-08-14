package translation

import (
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetSurfaceByAddress(address uintptr) *vulkan.VulkanSurface {
	t.surfacesMutex.Lock()
	surface, _ := t.surfaces[address]
	t.surfacesMutex.Unlock()

	return surface
}

func (t *GpuTranslator) GetSurface(descriptor spirvStructs.ImageDescriptor, compSwap uint32) (*vulkan.VulkanSurface, error) {
	image, err, _ := t.GetImage(descriptor, compSwap, true)
	if err != nil {
		return nil, err
	}

	t.surfacesMutex.Lock()
	surface, ok := t.surfaces[descriptor.BaseAddress]
	if ok && surface.ImageView.Image == image {
		t.surfacesMutex.Unlock()
		return surface, nil
	}

	if ok {
		delete(t.surfaces, descriptor.BaseAddress)
		if t.activeSurface != nil && t.activeSurface.Address == descriptor.BaseAddress {
			t.activeSurface = nil
		}
		if t.activeDepthSurface != nil && t.activeDepthSurface.Address == descriptor.BaseAddress {
			t.activeDepthSurface = nil
		}
		t.handles.DeferDestroySurface(surface)
	}
	t.surfacesMutex.Unlock()

	surface, err = vulkan.CreateSurface(t.handles, vulkan.VulkanSurfaceRequest{
		Descriptor: descriptor,
		CompSwap:   compSwap,
		Image:      image,
	})
	if err != nil {
		return nil, err
	}
	surface.Address = descriptor.BaseAddress
	surface.FirstUse = true
	surface.ContentValid = false
	surface.FrameUsed = 0
	surface.TextureId = imgui.TextureRef{}
	t.surfacesMutex.Lock()
	t.surfaces[descriptor.BaseAddress] = surface
	t.surfacesMutex.Unlock()

	if surface.ImageView.Image.ShouldUploadToVkImage(t.currentGuestFrame) {
		if err = surface.ImageView.Image.UploadToVkImage(t.handles, t.commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
			logger.Printf("failed to upload image: %v\n", err)
		}
	}
	return surface, nil
}

func (t *GpuTranslator) GetSurfaceTexture(surface *vulkan.VulkanSurface) imgui.TextureRef {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	if surface.TextureId.CData == nil {
		surface.TextureId = t.backend.CreateVulkanTexture(surface.Sampler, surface.ImageView.ImageView, vk.ImageLayoutGeneral)
	}

	return surface.TextureId
}
