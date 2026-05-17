package vulkan

func (t *GpuTranslator) GetFramebuffer(request FramebufferRequest) (*VulkanFramebuffer, error) {
	t.framebuffersMutex.Lock()
	fb, ok := t.framebuffers[request]
	t.framebuffersMutex.Unlock()
	if ok {
		return fb, nil
	}

	fb, err := t.createFramebuffer(request)
	if err != nil {
		return nil, err
	}

	t.framebuffersMutex.Lock()
	t.framebuffers[request] = fb
	t.framebuffersMutex.Unlock()

	return fb, nil
}
