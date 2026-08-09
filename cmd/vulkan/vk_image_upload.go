package vulkan

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (image *VulkanImage) UploadToVkImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) error {
	image.SyncLock.Lock()
	defer image.SyncLock.Unlock()

	var srcBuffer vk.Buffer
	var srcOffset vk.DeviceSize
	var width, height, pitch uint32
	linear := isLinearTileMode(image.FirstDescriptor.TilingIndex)
	if linear {
		// Guest backing is already row-major in mapped device buffer - copy directly.
		linearDimensions := image.GetLinearDimensions()
		width = linearDimensions.Width
		height = linearDimensions.Height
		pitch = linearDimensions.Pitch

		// Assign buffer.
		texBuffer, texOffset, err := getLinearBuffer(image.Address)
		if err != nil {
			return err
		}
		srcBuffer = texBuffer
		srcOffset = vk.DeviceSize(texOffset)

		// Wait for CPU writes to finish before reading the buffer.
		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageHostBit|vk.PipelineStageComputeShaderBit),
			vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessHostWriteBit | vk.AccessShaderWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
			}}, 0, nil, 0, nil)
	} else {
		// Tiled guest layouts must be detiled using compute shader.
		mipLevel := int(image.FirstDescriptor.BaseLevel)
		if mipLevel >= len(image.Layouts) {
			return fmt.Errorf("mip level %d out of range", mipLevel)
		}
		tiledDimensions := image.GetTiledDimensions(mipLevel)
		width = tiledDimensions.Width
		height = tiledDimensions.Height
		pitch = tiledDimensions.Pitch

		// Allocate staging buffer.
		size := image.GetStagingBufferSize()
		stagingBuffer, err := handles.StagingBufferPool.Get(handles, size)
		if err != nil {
			return err
		}
		defer handles.StagingBufferPool.Put(stagingBuffer)

		// Prepare source image buffer.
		srcBuffer = stagingBuffer.Buffer
		srcOffset = 0
		texBuffer, texOffsetBase, err := getLinearBuffer(image.Address)
		if err != nil {
			return err
		}
		texOffset := vk.DeviceSize(texOffsetBase) + vk.DeviceSize(image.Layouts[mipLevel].Offset)

		// Prepare detile pipeline.
		isMicro := !isMacroTiledMode(image.FirstDescriptor.TilingIndex)
		isDisplayMicro := usesDisplayMicroTiling(image.FirstDescriptor.TilingIndex)
		pipeline, err := GetDetilePipeline(handles, int(tiledDimensions.Bpp), isMicro, isDisplayMicro)
		if err != nil {
			return err
		}
		setToBind, err := pipeline.DescriptorPool.Get(handles, frame)
		if err != nil {
			return err
		}

		// Wait for CPU writes and previous staging buffer operations to finish before reading/writing the buffer.
		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageHostBit|vk.PipelineStageComputeShaderBit|vk.PipelineStageTransferBit),
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessHostWriteBit | vk.AccessShaderWriteBit | vk.AccessTransferReadBit | vk.AccessTransferWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit),
			}}, 0, nil, 0, nil)

		// Bind detile pipeline.
		vk.CmdBindPipeline(commandBuffer.CommandBuffer, vk.PipelineBindPointCompute, pipeline.Pipeline)

		vk.UpdateDescriptorSets(handles.Device, 2, []vk.WriteDescriptorSet{
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          setToBind,
				DstBinding:      0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeStorageBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{
					{Buffer: texBuffer, Offset: texOffset, Range: vk.DeviceSize(^uint64(0))},
				},
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          setToBind,
				DstBinding:      1,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeStorageBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{
					{Buffer: srcBuffer, Offset: 0, Range: size},
				},
			},
		}, 0, nil)

		// Set descriptor sets (input for detile pipeline).
		vk.CmdBindDescriptorSets(
			commandBuffer.CommandBuffer, vk.PipelineBindPointCompute,
			pipeline.PipelineLayout, 0,
			1, []vk.DescriptorSet{setToBind},
			0, nil,
		)

		// Push detile options.
		c0 := width / 8
		c1 := c0 * ((height + 7) / 8)
		pushConstants := DetilePushConstants{
			NumLevels: 0,
			Pitch:     width,
			Height:    height,
			C0:        c0,
			C1:        c1,
			IsRetile:  0,
		}
		vk.CmdPushConstants(
			commandBuffer.CommandBuffer, pipeline.PipelineLayout,
			vk.ShaderStageFlags(vk.ShaderStageComputeBit),
			0, uint32(DetilePushConstantsSize),
			unsafe.Pointer(&pushConstants),
		)

		// Dispatch detile shader.
		texels := width * height
		groups := (texels + 63) / 64
		vk.CmdDispatch(commandBuffer.CommandBuffer, groups, 1, 1)

		// Wait for compute shader to finish writing to the staging buffer.
		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
			vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessShaderWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
			}},
			0, nil, 0, nil)
	}

	// Copy staging buffer to GPU.
	image.BarrierTransferWrite(commandBuffer)
	extentWidth := width
	extentHeight := height
	if !isLinearTileMode(image.FirstDescriptor.TilingIndex) {
		mipLevel := uint(image.FirstDescriptor.BaseLevel)
		extentWidth = max(uint32(image.FirstDescriptor.Width)>>mipLevel, 1)
		extentHeight = max(uint32(image.FirstDescriptor.Height)>>mipLevel, 1)
	}
	vk.CmdCopyBufferToImage(commandBuffer.CommandBuffer, srcBuffer, image.Image, vk.ImageLayoutGeneral, 1, []vk.BufferImageCopy{{
		BufferOffset:      srcOffset,
		BufferRowLength:   pitch,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask: GetFormatAspectFlags(image.ImageFormat),
			MipLevel:   uint32(image.FirstDescriptor.BaseLevel),
			LayerCount: 1,
		},
		ImageOffset: vk.Offset3D{},
		ImageExtent: vk.Extent3D{Width: extentWidth, Height: extentHeight, Depth: 1},
	}})
	image.BarrierGeneralShaderAccess(commandBuffer)

	image.MarkSynced(frame)
	logger.Printf("[%s] uploaded image 0x%X (%dx%d/%v) to VRAM.\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
		linear,
	)

	return nil
}

func (image *VulkanImage) ShouldUploadToVkImage(frame uint64) bool {
	// Uploading surfaces is too expensive.
	if image.IsSurface || IsDepthFormat(image.ImageFormat) {
		return false
	}
	if image.HasSync(ImageSyncCpuModified) {
		return true
	}

	return false
}
