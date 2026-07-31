package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) DmaCopy(frame uint64, dmaCopy *gpu.LiverpoolDmaCopy) {
	t.EndRenderPass()

	copySize := uintptr(dmaCopy.Count * 4)
	err := t.DownloadRegionVkImages(dmaCopy.SrcAddress, copySize, t.commandBuffer)
	if err != nil {
		panic(err)
	}

	srcBuffer, srcOffset, err1 := t.GetLinearBuffer(dmaCopy.SrcAddress)
	dstBuffer, dstOffset, err2 := t.GetLinearBuffer(dmaCopy.DstAddress)
	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err2)
	}

	if logger.LogRenderer {
		logger.Printf("[%s] DMA copy of %s bytes from %s to %s.\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Yellow.Sprintf("0x%X", dmaCopy.Count),
			color.Yellow.Sprintf("0x%X", dmaCopy.SrcAddress),
			color.Yellow.Sprintf("0x%X", dmaCopy.DstAddress),
		)
	}
	vk.CmdCopyBuffer(t.commandBuffer.CommandBuffer, srcBuffer, dstBuffer, 1, []vk.BufferCopy{{
		SrcOffset: vk.DeviceSize(srcOffset),
		DstOffset: vk.DeviceSize(dstOffset),
		Size:      vk.DeviceSize(copySize),
	}})

	// Upload DMA destination into any overlapping VkImages (guest buffer is now fresh in-GPU-order).
	for _, image := range t.CollectGpuResourcesInRange(dmaCopy.DstAddress, copySize) {
		image.MarkCpuModified(t.currentGuestFrame)
	}
}
