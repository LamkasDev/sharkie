package vulkan

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (image *VulkanImage) DownloadFromVkImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) error {
	image.SyncLock.Lock()
	defer image.SyncLock.Unlock()

	linear := isLinearTileMode(image.FirstDescriptor.TilingIndex)
	if linear {
		// Guest backing is already row-major in mapped device buffer - copy directly.
		linearDimensions := image.GetLinearDimensions()
		width := linearDimensions.Width
		height := linearDimensions.Height
		pitch := linearDimensions.Pitch

		// Assign buffer.
		texBuffer, texOffset, err := getLinearBuffer(image.Address)
		if err != nil {
			return err
		}

		// Copy GPU buffer to RAM.
		image.BarrierTransferSrc(commandBuffer)
		bufferCopies := []vk.BufferImageCopy{{
			BufferOffset:    vk.DeviceSize(texOffset),
			BufferRowLength: pitch,
			ImageSubresource: vk.ImageSubresourceLayers{
				AspectMask: GetFormatAspectFlags(image.ImageFormat),
				MipLevel:   uint32(image.FirstDescriptor.BaseLevel),
				LayerCount: 1,
			},
			ImageExtent: vk.Extent3D{Width: width, Height: height, Depth: 1},
		}}
		vk.CmdCopyImageToBuffer(commandBuffer.CommandBuffer, image.Image, vk.ImageLayoutTransferSrcOptimal, texBuffer, 1, bufferCopies)

		// Wait for shader writes to finish before reading the buffer.
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
		// Tiled guest layouts must be retiled using compute shader.
		mipLevel := int(image.FirstDescriptor.BaseLevel)
		if mipLevel >= len(image.Layouts) {
			return fmt.Errorf("mip level %d out of range", mipLevel)
		}
		tiledDimensions := image.GetTiledDimensions(mipLevel)
		width := tiledDimensions.Width
		height := tiledDimensions.Height
		bpp := tiledDimensions.Bpp

		// Allocate staging buffer.
		size := image.GetStagingBufferSize()
		stagingBuffer, err := handles.StagingBufferPool.Get(handles, size)
		if err != nil {
			return err
		}
		defer handles.StagingBufferPool.Put(stagingBuffer)

		// Wait for previous staging buffer operations to finish before transfer writes to it.
		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageTransferBit|vk.PipelineStageComputeShaderBit),
			vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessTransferReadBit | vk.AccessTransferWriteBit | vk.AccessShaderReadBit | vk.AccessShaderWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			}}, 0, nil, 0, nil)

		// Copy image to staging buffer.
		image.BarrierTransferSrc(commandBuffer)
		extentWidth := max(uint32(image.FirstDescriptor.Width)>>uint(mipLevel), 1)
		extentHeight := max(uint32(image.FirstDescriptor.Height)>>uint(mipLevel), 1)
		bufferCopies := []vk.BufferImageCopy{{
			BufferRowLength: width,
			ImageSubresource: vk.ImageSubresourceLayers{
				AspectMask: GetFormatAspectFlags(image.ImageFormat),
				MipLevel:   uint32(mipLevel),
				LayerCount: 1,
			},
			ImageExtent: vk.Extent3D{Width: extentWidth, Height: extentHeight, Depth: 1},
		}}
		vk.CmdCopyImageToBuffer(commandBuffer.CommandBuffer, image.Image, vk.ImageLayoutTransferSrcOptimal, stagingBuffer.Buffer, 1, bufferCopies)
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
		texOffset := vk.DeviceSize(texOffsetBase) + vk.DeviceSize(image.Layouts[mipLevel].Offset)

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
					{Buffer: stagingBuffer.Buffer, Offset: 0, Range: size},
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
			pipeline.PipelineLayout, 0,
			1, []vk.DescriptorSet{setToBind},
			0, nil,
		)

		// Push retile options.
		c0 := width / 8
		c1 := c0 * ((height + 7) / 8)
		pushConstants := DetilePushConstants{
			NumLevels: 0,
			Pitch:     width,
			Height:    height,
			C0:        c0,
			C1:        c1,
			IsRetile:  1,
		}
		vk.CmdPushConstants(
			commandBuffer.CommandBuffer, pipeline.PipelineLayout,
			vk.ShaderStageFlags(vk.ShaderStageComputeBit),
			0, uint32(DetilePushConstantsSize),
			unsafe.Pointer(&pushConstants),
		)

		// Dispatch retile shader.
		texels := width * height
		groups := (texels + 63) / 64
		vk.CmdDispatch(commandBuffer.CommandBuffer, groups, 1, 1)

		// Wait for compute shader to finish writing to RAM.
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
	if logger.LogRenderer {
		logger.Printf("[%s] downloaded image 0x%X (%dx%d/%v) to RAM.\n",
			color.Blue.Sprintf("Frame %d", frame),
			image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
			linear,
		)
	}

	return nil
}

func (image *VulkanImage) ShouldDownloadFromVkImage() bool {
	// Uploading surfaces is too expensive.
	if image.IsSurface || IsDepthFormat(image.ImageFormat) {
		return false
	}
	if image.HasSync(ImageSyncGpuModified) {
		return true
	}

	return false
}
