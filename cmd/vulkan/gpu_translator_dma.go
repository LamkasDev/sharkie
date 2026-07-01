package vulkan

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
)

func dmaRangesOverlap(aStart, aSize, bStart, bSize uintptr) bool {
	if aSize == 0 || bSize == 0 {
		return false
	}
	aEnd := aStart + aSize
	bEnd := bStart + bSize

	return aStart < bEnd && bStart < aEnd
}

func (t *GpuTranslator) DmaCopy(frame uint64, commandBuffer vk.CommandBuffer, dmaCopy *gpu.LiverpoolDmaCopy) {
	t.ResearchLogDmaCopy(frame, dmaCopy)
	if t.activePass != vk.NullRenderPass {
		t.EndRenderPass(commandBuffer)
	}

	// Get buffers.
	srcBuffer, srcOffset, err1 := t.GetBufferFromAddress(dmaCopy.SrcAddress)
	dstBuffer, dstOffset, err2 := t.GetBufferFromAddress(dmaCopy.DstAddress)
	if err1 != nil {
		logger.Print(err1)
		return
	}
	if err2 != nil {
		logger.Print(err2)
		return
	}

	// Perform raw linear copy between the buffers using Vulkan.
	vk.CmdCopyBuffer(commandBuffer, srcBuffer, dstBuffer, 1, []vk.BufferCopy{{
		SrcOffset: vk.DeviceSize(srcOffset),
		DstOffset: vk.DeviceSize(dstOffset),
		Size:      vk.DeviceSize(dmaCopy.Count * 4),
	}})

	// Add a barrier before the image copies.
	vk.CmdPipelineBarrier(commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
		}}, 0, nil, 0, nil,
	)

	// If the DMA destination overlaps with any known texture, we must update the texture directly.
	t.imagesMutex.Lock()
	for _, descriptor := range t.imageDescriptors {
		_, bpp := TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
		pitch := descriptor.Pitch
		if pitch == 0 {
			pitch = descriptor.Width
		}
		size := uintptr(uint32(pitch) * uint32(descriptor.Height) * uint32(bpp))
		if !dmaRangesOverlap(dmaCopy.DstAddress, uintptr(dmaCopy.Count*4), descriptor.BaseAddress, size) {
			continue
		}

		if image, ok := t.images[descriptor.BaseAddress]; ok {
			texBuffer, texOffset, err := t.GetBufferFromAddress(descriptor.BaseAddress)
			if err != nil {
				logger.Print(err)
				continue
			}
			vk.CmdCopyBufferToImage(commandBuffer, texBuffer, image, vk.ImageLayoutGeneral, 1, []vk.BufferImageCopy{{
				BufferOffset:      vk.DeviceSize(texOffset),
				BufferRowLength:   uint32(descriptor.Pitch),
				BufferImageHeight: 0,
				ImageSubresource: vk.ImageSubresourceLayers{
					AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
					MipLevel:   uint32(descriptor.BaseLevel),
					LayerCount: 1,
				},
				ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
				ImageExtent: vk.Extent3D{
					Width:  uint32(descriptor.Width),
					Height: uint32(descriptor.Height),
					Depth:  uint32(descriptor.Depth),
				},
			}})
		}
	}

	// If the DMA destination overlaps with any known surface, we must update the surface directly.
	t.surfacesMutex.Lock()
	for key, surface := range t.surfaces {
		size := uintptr(surface.Value.width * surface.Value.height * 4) // worst case 4 bytes
		if !dmaRangesOverlap(dmaCopy.DstAddress, uintptr(dmaCopy.Count*4), key.GpuAddress, size) || surface.Value.image == vk.NullImage {
			continue
		}

		// For surfaces, assume they cover the whole thing.
		texBuffer, texOffset, err := t.GetBufferFromAddress(key.GpuAddress)
		if err != nil {
			logger.Print(err)
			continue
		}
		vk.CmdCopyBufferToImage(commandBuffer, texBuffer, surface.Value.image, vk.ImageLayoutGeneral, 1, []vk.BufferImageCopy{{
			BufferOffset:      vk.DeviceSize(texOffset),
			BufferRowLength:   uint32(surface.Value.width),
			BufferImageHeight: 0,
			ImageSubresource: vk.ImageSubresourceLayers{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:   0,
				LayerCount: 1,
			},
			ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
			ImageExtent: vk.Extent3D{
				Width:  uint32(surface.Value.width),
				Height: uint32(surface.Value.height),
				Depth:  1,
			},
		}})
	}
	t.surfacesMutex.Unlock()
	t.imagesMutex.Unlock()

	// Add a barrier after the buffer and image copies.
	vk.CmdPipelineBarrier(commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit|vk.PipelineStageComputeShaderBit|vk.PipelineStageAllGraphicsBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit | vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessIndexReadBit | vk.AccessVertexAttributeReadBit),
		}}, 0, nil, 0, nil,
	)
}
