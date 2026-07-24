package translation

import (
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) EndRenderPass() {
	if t.activePass == vk.NullRenderPass {
		return
	}
	vk.CmdEndRenderPass(t.commandBuffer.CommandBuffer)
	t.activePass = vk.NullRenderPass
	t.FlushPendingResourceBarriers(t.commandBuffer, 0)
}

// StartRenderPass ends any active pass, flushes barriers, and starts a new one.
func (t *GpuTranslator) StartRenderPass(pass vk.RenderPass, passNoClear vk.RenderPass, fb vk.Framebuffer, pipeline vk.Pipeline, clearValues []vk.ClearValue, width, height uint32) {
	t.EndRenderPass()
	vk.CmdBeginRenderPass(t.commandBuffer.CommandBuffer, &vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      pass,
		Framebuffer:     fb,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: width, Height: height}},
		ClearValueCount: uint32(len(clearValues)),
		PClearValues:    clearValues,
	}, vk.SubpassContentsInline)
	vk.CmdBindPipeline(t.commandBuffer.CommandBuffer, vk.PipelineBindPointGraphics, pipeline)

	t.activePass = pass
	t.activePassNoClear = passNoClear
	t.activeFramebuffer = fb
	t.activePipeline = pipeline
}

// ResumeActiveRenderPass continues a render pass without clearing attachments.
func (t *GpuTranslator) ResumeActiveRenderPass() {
	if t.activePass != vk.NullRenderPass {
		return
	}
	if t.activePassNoClear == vk.NullRenderPass || t.activeSurface == nil {
		return
	}

	t.StartRenderPass(t.activePassNoClear, t.activePassNoClear, t.activeFramebuffer, t.activePipeline, nil,
		uint32(t.activeSurface.ImageView.Image.FirstDescriptor.Width),
		uint32(t.activeSurface.ImageView.Image.FirstDescriptor.Height))
	if t.activeDynamicState != nil {
		t.SetDynamicState(t.activeDynamicState)
	}
}

// FlushPendingResourceBarriers must be called outside an active render pass.
func (t *GpuTranslator) FlushPendingResourceBarriers(commandBuffer *vulkan.VulkanCommandBuffer, excludeAddress uintptr) {
	if t.activePass != vk.NullRenderPass {
		return
	}

	t.imagesMutex.Lock()
	images := make([]*vulkan.VulkanImage, 0, len(t.images))
	for _, image := range t.images {
		if address, hasBarrier := image.Address, image.HasSync(vulkan.ImageSyncNeedsReadBarrier); address != excludeAddress && hasBarrier {
			images = append(images, image)
		}
	}
	t.imagesMutex.Unlock()

	for _, image := range images {
		t.surfacesMutex.Lock()
		surface := t.surfaces[image.Address]
		t.surfacesMutex.Unlock()
		if surface != nil {
			surface.ImageView.Image.BarrierShaderRead(commandBuffer)
			continue
		}
		image.BarrierShaderRead(commandBuffer)
	}
}
