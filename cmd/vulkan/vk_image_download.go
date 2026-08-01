package vulkan

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (image *VulkanImage) DownloadFromVkImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) error {
	image.SyncLock.Lock()
	defer image.SyncLock.Unlock()

	// Prepare download parameters.
	width := uint32(image.FirstDescriptor.Width)
	height := uint32(image.FirstDescriptor.Height)
	pitch := uint32(image.FirstDescriptor.Pitch)
	bpp := GetBytesPerPixel(image.FirstDescriptor.DataFormat)

	var copyBytes vk.DeviceSize
	var rowPitch uint32
	isBlock := image.FirstDescriptor.DataFormat >= 35 && image.FirstDescriptor.DataFormat <= 41
	if isBlock {
		copyBytes = vk.DeviceSize(((width + 3) / 4) * ((height + 3) / 4) * bpp)
		rowPitch = uint32((pitch + 3) / 4)
	} else {
		copyBytes = vk.DeviceSize(width * height * bpp)
		rowPitch = uint32(pitch)
		if rowPitch == 0 {
			rowPitch = width
		}
	}

	linear := isLinearTileMode(image.FirstDescriptor.TilingIndex)
	if linear {
		texBuffer, texOffset, err := getLinearBuffer(image.Address)
		if err != nil {
			return err
		}

		image.BarrierTransferSrc(commandBuffer)
		bufferCopies := []vk.BufferImageCopy{{
			BufferOffset:     vk.DeviceSize(texOffset),
			BufferRowLength:  rowPitch,
			ImageSubresource: vk.ImageSubresourceLayers{AspectMask: GetFormatAspectFlags(image.ImageFormat), LayerCount: 1},
			ImageExtent:      vk.Extent3D{Width: width, Height: height, Depth: 1},
		}}
		vk.CmdCopyImageToBuffer(commandBuffer.CommandBuffer, image.Image, vk.ImageLayoutTransferSrcOptimal, texBuffer, 1, bufferCopies)

		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit|vk.PipelineStageHostBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessHostReadBit),
			}}, 0, nil, 0, nil)

		image.BarrierGeneralShaderAccess(commandBuffer)
	} else {
		// Retile using compute shader.
		mipLevel := int(image.FirstDescriptor.BaseLevel)
		layout := image.Layouts[mipLevel]
		layoutWidth := uint32(layout.Pitch)
		layoutHeight := uint32(layout.Height)
		if isBlock {
			layoutWidth /= 4
			layoutHeight /= 4
		}

		// Allocate staging buffer.
		stagingBytes := vk.DeviceSize(layoutWidth * layoutHeight * uint32(bpp))
		buffer, bufferMem, err := AllocateBuffer(handles, stagingBytes,
			vk.BufferUsageFlags(vk.BufferUsageTransferDstBit|vk.BufferUsageStorageBufferBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit))
		if err != nil {
			return err
		}
		handles.DeferDestroyBuffer(buffer, bufferMem)

		// Copy image to staging buffer.
		image.BarrierTransferSrc(commandBuffer)
		bufferCopies := []vk.BufferImageCopy{{
			ImageSubresource: vk.ImageSubresourceLayers{AspectMask: GetFormatAspectFlags(image.ImageFormat), LayerCount: 1},
			ImageExtent:      vk.Extent3D{Width: layoutWidth, Height: layoutHeight, Depth: 1},
		}}
		vk.CmdCopyImageToBuffer(commandBuffer.CommandBuffer, image.Image, vk.ImageLayoutTransferSrcOptimal, buffer, 1, bufferCopies)
		image.BarrierGeneralShaderAccess(commandBuffer)

		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit),
			}}, 0, nil, 0, nil)

		// Prepare destination image buffer.
		texBuffer, texOffsetBase, err := getLinearBuffer(image.Address)
		if err != nil {
			return err
		}
		texOffset := vk.DeviceSize(texOffsetBase) + vk.DeviceSize(layout.Offset)

		// Prepare retile pipeline.
		isMicro := !isMacroTiledMode(image.FirstDescriptor.TilingIndex)
		isDisplayMicro := usesDisplayMicroTiling(image.FirstDescriptor.TilingIndex)
		pipeline, err := GetDetilePipeline(handles, int(bpp), isMicro, isDisplayMicro)
		if err != nil {
			return err
		}
		setToBind, err := pipeline.DescriptorPool.Get(handles, frame)
		if err != nil {
			return err
		}

		vk.CmdBindPipeline(commandBuffer.CommandBuffer, vk.PipelineBindPointCompute, pipeline.Pipeline)

		vk.UpdateDescriptorSets(handles.Device, 2, []vk.WriteDescriptorSet{
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          setToBind,
				DstBinding:      0, // in_data (Linear Staging)
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeStorageBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{
					{Buffer: buffer, Offset: 0, Range: copyBytes},
				},
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          setToBind,
				DstBinding:      1, // out_data (Tiled Guest RAM)
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeStorageBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{
					{Buffer: texBuffer, Offset: texOffset, Range: vk.DeviceSize(^uint64(0))},
				},
			},
		}, 0, nil)

		vk.CmdBindDescriptorSets(
			commandBuffer.CommandBuffer, vk.PipelineBindPointCompute,
			pipeline.PipelineLayout, spirvStructs.DescriptorSetSlotStatic,
			1, []vk.DescriptorSet{setToBind},
			0, nil,
		)

		c0 := layoutWidth / 8
		c1 := c0 * ((layoutHeight + 7) / 8)
		pushConstants := make([]uint32, 6)
		pushConstants[0] = 0 // num_levels
		pushConstants[1] = uint32(layout.Pitch)
		pushConstants[2] = layoutHeight
		pushConstants[3] = c0
		pushConstants[4] = c1
		pushConstants[5] = 1 // is_retile = true
		vk.CmdPushConstants(
			commandBuffer.CommandBuffer, pipeline.PipelineLayout,
			vk.ShaderStageFlags(vk.ShaderStageComputeBit),
			0, uint32(len(pushConstants)*4),
			unsafe.Pointer(&pushConstants[0]),
		)

		texels := layoutWidth * layoutHeight
		groups := (texels + 63) / 64
		vk.CmdDispatch(commandBuffer.CommandBuffer, groups, 1, 1)

		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit|vk.PipelineStageHostBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessShaderWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessHostReadBit),
			}}, 0, nil, 0, nil)
	}

	image.MarkSynced(frame)
	logger.Printf("[%s] downloaded image 0x%X (%dx%d/%v) to RAM.\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
		linear,
	)

	return nil
}

func (image *VulkanImage) ShouldDownloadFromVkImage() bool {
	// Uploading surfaces is too expensive.
	if IsDepthFormat(image.ImageFormat) {
		return false
	}
	if image.HasSync(ImageSyncGpuModified) {
		return true
	}

	return false
}
