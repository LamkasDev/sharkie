package translation

import (
	"fmt"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vkGcn "github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetImageView(descriptor spirvStructs.ImageDescriptor) (*vulkan.VulkanImageView, error, bool) {
	hash := descriptor.Hash()
	format, _ := vkGcn.TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
	if format == vk.FormatUndefined {
		return nil, fmt.Errorf("invalid format"), false
	}

	image, err, created := t.GetImage(descriptor, format, false)
	if err != nil {
		return nil, err, false
	}

	t.imagesMutex.Lock()
	view, ok := t.imageViews[hash]
	if ok && view.Image == image && view.Image.Generation == image.Generation {
		t.imagesMutex.Unlock()
		if image.ShouldUploadToVkImage(t.currentGuestFrame) {
			t.EndRenderPass()
			if err = image.UploadToVkImage(t.handles, t.commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
				return nil, err, false
			}
		}

		return view, nil, false
	}
	recreated := created || ok
	if ok && !t.IsOwnedSurfaceImageView(view) {
		view.Destroy(t.handles.Device)
	}
	if ok {
		delete(t.imageViews, hash)
	}
	t.imagesMutex.Unlock()

	view, err = vulkan.CreateImageView(t.handles, vulkan.VulkanImageViewRequest{
		Image:      image,
		Descriptor: descriptor,
	})
	if err != nil {
		return nil, err, false
	}
	t.imagesMutex.Lock()
	t.imageViews[hash] = view
	t.imagesMutex.Unlock()

	return view, nil, recreated
}
