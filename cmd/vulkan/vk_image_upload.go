package vulkan

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (image *VulkanImage) UploadToVkImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) error {
	image.SyncLock.Lock()
	defer image.SyncLock.Unlock()

	var srcBuffer vk.Buffer
	var srcOffset vk.DeviceSize
	var rowPitch uint32
	var width, height uint32
	if isLinearTileMode(image.FirstDescriptor.TilingIndex) {
		// Guest backing is already row-major in mapped device buffer - copy directly.
		texBuffer, texOffset, err := getLinearBuffer(image.Address)
		if err != nil {
			return err
		}

		srcBuffer = texBuffer
		srcOffset = vk.DeviceSize(texOffset)
		rowPitch = uint32(image.FirstDescriptor.Pitch)
		if rowPitch == 0 {
			rowPitch = uint32(image.FirstDescriptor.Width)
		}
		width = uint32(image.FirstDescriptor.Width)
		height = uint32(image.FirstDescriptor.Height)

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
		layout := image.Layouts[mipLevel]
		width = uint32(layout.Pitch)
		height = uint32(layout.Height)
		bpp := gcn.GetBytesPerPixel(image.FirstDescriptor.DataFormat)
		isBlock := image.FirstDescriptor.DataFormat >= 35 && image.FirstDescriptor.DataFormat <= 41
		if isBlock {
			rowPitch = uint32(layout.Pitch)
			width = uint32(layout.Pitch / 4)
			height = uint32(layout.Height / 4)
		} else {
			rowPitch = uint32(layout.Pitch)
		}

		// Allocate staging buffer.
		size := vk.DeviceSize(width * height * uint32(bpp))
		buffer, bufferMem, err := AllocateBuffer(handles, size,
			vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageStorageBufferBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit))
		if err != nil {
			return err
		}
		handles.DeferDestroyBuffer(buffer, bufferMem)

		// Prepare source image buffer.
		srcBuffer = buffer
		srcOffset = 0
		texBuffer, texOffsetBase, err := getLinearBuffer(image.Address)
		if err != nil {
			return err
		}
		texOffset := vk.DeviceSize(texOffsetBase) + vk.DeviceSize(layout.Offset)

		// Prepare detile pipeline.
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

		// Wait for CPU writes to finish before reading the buffer.
		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageHostBit|vk.PipelineStageComputeShaderBit),
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessHostWriteBit | vk.AccessShaderWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit),
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
					{
						Buffer: texBuffer,
						Offset: texOffset,
						Range:  vk.DeviceSize(^uint64(0)),
					},
				},
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          setToBind,
				DstBinding:      1,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeStorageBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{
					{
						Buffer: srcBuffer,
						Offset: 0,
						Range:  size,
					},
				},
			},
		}, 0, nil)

		// Set descriptor sets (input for detile pipeline).
		vk.CmdBindDescriptorSets(
			commandBuffer.CommandBuffer, vk.PipelineBindPointCompute,
			pipeline.PipelineLayout, spirvStructs.DescriptorSetSlotStatic,
			1, []vk.DescriptorSet{setToBind},
			0, nil,
		)

		// Push detile options.
		pitch := uint32(layout.Pitch)
		heightPush := uint32(layout.Height)
		if isBlock {
			pitch /= 4
			heightPush /= 4
		}
		c0 := pitch / 8
		c1 := c0 * ((heightPush + 7) / 8)
		pushConstants := make([]uint32, 6)
		pushConstants[0] = 0 // num_levels
		pushConstants[1] = pitch
		pushConstants[2] = heightPush
		pushConstants[3] = c0
		pushConstants[4] = c1
		pushConstants[5] = 0 // is_retile = false
		vk.CmdPushConstants(
			commandBuffer.CommandBuffer, pipeline.PipelineLayout,
			vk.ShaderStageFlags(vk.ShaderStageComputeBit),
			0, uint32(len(pushConstants)*4),
			unsafe.Pointer(&pushConstants[0]),
		)

		// Dispatch detile shader.
		texels := width * height
		if isBlock {
			texels = width * height
		}
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
		mipLevel := int(image.FirstDescriptor.BaseLevel)
		layout := image.Layouts[mipLevel]
		extentWidth = uint32(layout.Width)
		extentHeight = uint32(layout.Height)
	}
	vk.CmdCopyBufferToImage(commandBuffer.CommandBuffer, srcBuffer, image.Image, vk.ImageLayoutGeneral, 1, []vk.BufferImageCopy{{
		BufferOffset:      srcOffset,
		BufferRowLength:   rowPitch,
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
	logger.Printf("[%s] uploaded image 0x%X (%dx%d) to VRAM.\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)

	return nil
}

func (image *VulkanImage) ShouldUploadToVkImage(frame uint64) bool {
	// Uploading surfaces is too expensive.
	if IsDepthFormat(image.ImageFormat) {
		return false
	}
	if image.HasSync(ImageSyncCpuModified) {
		return true
	}

	return false
}
