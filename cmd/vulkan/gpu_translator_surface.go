package vulkan

import (
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	vk "github.com/goki/vulkan"
)

// GpuSurface is a Vulkan-side render target that corresponds to a single
// GPU-address-identified framebuffer surface registered by the game.
type GpuSurface struct {
	Key       SurfaceKey
	Value     VulkanSurface
	TextureId imgui.TextureRef

	// FirstUse tracks whether the image has been transitioned from UNDEFINED.
	FirstUse bool

	// ContentValid tracks whether the surface has valid content.
	ContentValid bool

	// FrameUsed tracks the last frame this surface was used in.
	FrameUsed uint64
}

func (s *GpuSurface) Destroy(device vk.Device) {
	s.Value.Destroy(device)
}

func (t *GpuTranslator) GetSurface(request SurfaceRequest) (*GpuSurface, error) {
	t.surfacesMutex.Lock()
	surface, ok := t.surfaces[request.SurfaceKey]
	if ok {
		// TODO: this should be temporary.
		// Recreate surface if format or size changed.
		if surface.Value.format != request.Format || surface.Value.width != request.Width || surface.Value.height != request.Height {
			t.surfacesMutex.Unlock()
			vulkanSurface, err := t.createSurface(request)
			if err != nil {
				return nil, err
			}
			t.surfacesMutex.Lock()
			surface.Value.DestroyViews(t.handles.Device)
			surface.Value = vulkanSurface
			surface.TextureId = imgui.TextureRef{} // Reset texture so it's recreated.
		}
		t.surfacesMutex.Unlock()
		return surface, nil
	}
	t.surfacesMutex.Unlock()

	vulkanSurface, err := t.createSurface(request)
	if err != nil {
		return nil, err
	}

	t.surfacesMutex.Lock()
	surface = &GpuSurface{
		Key:      request.SurfaceKey,
		Value:    vulkanSurface,
		FirstUse: true,
	}
	t.surfaces[request.SurfaceKey] = surface
	t.surfacesMutex.Unlock()

	return surface, nil
}

func (t *GpuTranslator) GetSurfaceByAddress(address uintptr) *GpuSurface {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()

	return t.surfaces[SurfaceKey{GpuAddress: address}]
}

func (t *GpuTranslator) GetSurfaceTexture(surface *GpuSurface) imgui.TextureRef {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	if surface.TextureId.CData == nil {
		surface.TextureId = t.backend.CreateVulkanTexture(surface.Value.sampler, surface.Value.imageView, vk.ImageLayoutShaderReadOnlyOptimal)
	}

	return surface.TextureId
}
