package translation

import "github.com/LamkasDev/sharkie/cmd/vulkan"

func (t *GpuTranslator) GetFramebuffer(request vulkan.FramebufferRequest) (*vulkan.VulkanFramebuffer, error) {
	t.framebuffersMutex.Lock()
	fb, ok := t.framebuffers[request]
	t.framebuffersMutex.Unlock()
	if ok {
		return fb, nil
	}

	fb, err := vulkan.CreateFramebuffer(t.handles, request)
	if err != nil {
		return nil, err
	}

	t.framebuffersMutex.Lock()
	t.framebuffers[request] = fb
	t.framebuffersMutex.Unlock()

	return fb, nil
}
