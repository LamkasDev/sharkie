package vulkan

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

type ImageSyncFlags uint8

const (
	ImageSyncGpuModified      ImageSyncFlags = 1 << 0
	ImageSyncCpuModified      ImageSyncFlags = 1 << 1
	ImageSyncNeedsReadBarrier ImageSyncFlags = 1 << 3
)

func (image *VulkanImage) HasSync(flag ImageSyncFlags) bool {
	return image.SyncFlags&flag != 0
}

func (image *VulkanImage) SetSync(flag ImageSyncFlags) {
	image.SyncFlags |= flag
}

func (image *VulkanImage) ClearSync(flag ImageSyncFlags) {
	image.SyncFlags &^= flag
}

func (image *VulkanImage) MarkCpuModified(frame uint64) {
	image.SetSync(ImageSyncCpuModified)
	image.ClearSync(ImageSyncGpuModified)
}

func (image *VulkanImage) MarkGpuModified(frame uint64) {
	if image.HasSync(ImageSyncGpuModified) {
		image.SetSync(ImageSyncNeedsReadBarrier)
		return
	}
	image.SetSync(ImageSyncGpuModified | ImageSyncNeedsReadBarrier)
	image.ClearSync(ImageSyncCpuModified)
	structs.GlobalMemoryManager.UpdateTraps(image.Address, image.GuestSize, 0) // PROT_NONE
}

func (image *VulkanImage) MarkSynced(frame uint64) {
	image.ClearSync(ImageSyncCpuModified | ImageSyncGpuModified)
	structs.GlobalMemoryManager.UpdateTraps(image.Address, image.GuestSize, 1) // PROT_READ
}

type VulkanImage struct {
	Address  uintptr
	Image    vk.Image
	ImageMem vk.DeviceMemory

	FirstDescriptor spirvStructs.ImageDescriptor
	ImageFormat     vk.Format
	ImageAspect     vk.ImageAspectFlags

	ImageLayout vk.ImageLayout
	ImageAccess vk.AccessFlags
	ImageStage  vk.PipelineStageFlags

	IsSurface  bool
	Generation uint32
	Layouts    []MipLayout
	GuestSize  uintptr

	SyncFlags ImageSyncFlags
	SyncLock  sync.Mutex
}

type VulkanImageRequest struct {
	Descriptor spirvStructs.ImageDescriptor
	Format     vk.Format
	IsSurface  bool
}

func CreateImage(handles *VulkanHandles, request VulkanImageRequest, commandBuffer *VulkanCommandBuffer, frame uint64) (*VulkanImage, error) {
	// Figure out image flags.
	imageUsage := vk.ImageUsageFlags(vk.ImageUsageSampledBit | vk.ImageUsageStorageBit | vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit)
	aspectMask := vk.ImageAspectFlags(vk.ImageAspectColorBit)
	dstLayout := vk.ImageLayoutGeneral
	dstAccess := vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit)
	if request.IsSurface {
		imageUsage = vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageStorageBit | vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit)
		dstAccess = vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit)
		if IsDepthFormat(request.Format) {
			imageUsage = vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit)
			aspectMask = vk.ImageAspectFlags(vk.ImageAspectDepthBit | vk.ImageAspectStencilBit)
			dstLayout = vk.ImageLayoutDepthStencilAttachmentOptimal
			dstAccess = vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit | vk.AccessDepthStencilAttachmentWriteBit)
		}
	}
	image := &VulkanImage{
		Address:         request.Descriptor.BaseAddress,
		FirstDescriptor: request.Descriptor,
		ImageFormat:     request.Format,
		ImageLayout:     vk.ImageLayoutUndefined,
		ImageAspect:     aspectMask,
		ImageAccess:     vk.AccessFlags(vk.AccessNone),
		ImageStage:      vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit),
		IsSurface:       request.IsSurface,
		Generation:      1,
		Layouts:         computeMipLayouts(request.Descriptor, mipLevelCount(request.Descriptor)),
		SyncLock:        sync.Mutex{},
		GuestSize:       DescriptorGuestSize(request.Descriptor),
	}
	pinner := runtime.Pinner{}
	pinner.Pin(image)
	if len(image.Layouts) > 0 {
		pinner.Pin(&image.Layouts[0])
	}
	defer pinner.Unpin()

	// Create VkImage.
	if !IsDepthFormat(request.Format) {
		imageUsage |= vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit)
	}
	result := vk.CreateImage(handles.Device, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		Flags:     vk.ImageCreateFlags(vk.ImageCreateMutableFormatBit),
		ImageType: vk.ImageType2d,
		Format:    request.Format,
		Extent: vk.Extent3D{
			Width:  uint32(request.Descriptor.Width),
			Height: uint32(request.Descriptor.Height),
			Depth:  uint32(request.Descriptor.Depth),
		},
		MipLevels:     uint32(len(image.Layouts)),
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        vk.ImageTilingOptimal,
		Usage:         imageUsage,
		InitialLayout: vk.ImageLayoutUndefined,
	}, nil, &image.Image)
	if err := NewError(result); err != nil {
		return nil, err
	}
	// SetDebugUtilsObjectName(handles.Instance, handles.Device, vk.ObjectTypeImage, uint64(uintptr(unsafe.Pointer(image))), fmt.Sprintf("2D Image 0x%X", request.Descriptor.BaseAddress))

	// Allocate and bind memory.
	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(handles.Device, image.Image, &memReqs)
	memReqs.Deref()

	priorityInfo := unsafe.Pointer(nil)
	var priorityInfoExt VkMemoryPriorityAllocateInfoEXT
	if request.IsSurface {
		priorityInfoExt = NewPriorityInfo(0, 1.0)
		priorityInfo = unsafe.Pointer(&priorityInfoExt)
	}

	result = vk.AllocateMemory(handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		PNext:           priorityInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyDeviceLocalBit),
	}, nil, &image.ImageMem)
	if err := NewError(result); err != nil {
		return nil, err
	}
	SetDeviceMemoryPriority(handles.Instance, handles.Device, image.ImageMem, 1.0)
	vk.BindImageMemory(handles.Device, image.Image, image.ImageMem, 0)

	// Transition image.
	err := RunWithCommandBuffer(handles, func(commandBuffer *VulkanCommandBuffer) {
		ImageBarrier(commandBuffer, image, dstLayout, dstAccess, vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit), aspectMask)
	}, frame)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (image *VulkanImage) Destroy(device vk.Device) {
	if image.Image == vk.NullImage {
		return
	}
	vk.DestroyImage(device, image.Image, nil)
	image.Image = vk.NullImage
	vk.FreeMemory(device, image.ImageMem, nil)
	image.ImageMem = vk.NullDeviceMemory
}

func (image *VulkanImage) BarrierShaderRead(commandBuffer *VulkanCommandBuffer) {
	image.ClearSync(ImageSyncNeedsReadBarrier)
	if IsDepthFormat(image.ImageFormat) {
		ImageBarrier(commandBuffer, image,
			vk.ImageLayoutDepthStencilAttachmentOptimal,
			vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessDepthStencilAttachmentReadBit),
			vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
			GetFormatAspectFlags(image.ImageFormat),
		)
		return
	}

	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessShaderReadBit),
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierColorAttachment(commandBuffer *VulkanCommandBuffer) {
	if IsDepthFormat(image.ImageFormat) {
		return
	}
	if image.ImageLayout != vk.ImageLayoutGeneral {
		ImageBarrier(commandBuffer, image,
			vk.ImageLayoutGeneral,
			vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit),
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
			vk.ImageAspectFlags(vk.ImageAspectColorBit),
		)
	}
	if !image.HasSync(ImageSyncNeedsReadBarrier) {
		return
	}

	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessColorAttachmentReadBit|vk.AccessColorAttachmentWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		vk.ImageAspectFlags(vk.ImageAspectColorBit),
	)
}

func (image *VulkanImage) BarrierSampledRead(commandBuffer *VulkanCommandBuffer) {
	if image.ImageLayout == vk.ImageLayoutGeneral {
		return
	}
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessShaderReadBit),
		vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit|vk.PipelineStageComputeShaderBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierComputeStorageWrite(commandBuffer *VulkanCommandBuffer) {
	if image.ImageLayout == vk.ImageLayoutGeneral {
		return
	}
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierTransferSrc(commandBuffer *VulkanCommandBuffer) {
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutTransferSrcOptimal,
		vk.AccessFlags(vk.AccessTransferReadBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierTransferWrite(commandBuffer *VulkanCommandBuffer) {
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessTransferWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierGeneralShaderAccess(commandBuffer *VulkanCommandBuffer) {
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) DownloadFromVkImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, frame uint64) (func(), error) {
	image.SyncLock.Lock()
	defer image.SyncLock.Unlock()

	// Prepare download parameters.
	width := uint32(image.FirstDescriptor.Width)
	height := uint32(image.FirstDescriptor.Height)
	pitch := uint32(image.FirstDescriptor.Pitch)
	bpp := structs.GetBytesPerPixel(image.FirstDescriptor.DataFormat)

	var copyBytes vk.DeviceSize
	var guestBytes uintptr
	isBlock := image.FirstDescriptor.DataFormat >= 35 && image.FirstDescriptor.DataFormat <= 41
	if isBlock {
		copyBytes = vk.DeviceSize(((width + 3) / 4) * ((height + 3) / 4) * bpp)
		guestBytes = uintptr(((pitch + 3) / 4) * ((height + 3) / 4) * bpp)
	} else {
		copyBytes = vk.DeviceSize(width * height * bpp)
		guestBytes = uintptr(pitch) * uintptr(height) * uintptr(bpp)
	}

	// Allocate staging buffer.
	buffer, bufferMem, err := AllocateBuffer(handles, copyBytes, vk.BufferUsageFlags(vk.BufferUsageTransferDstBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("staging buffer: %w", err)
	}

	// Copy image to staging buffer.
	image.BarrierTransferSrc(commandBuffer)
	bufferCopies := []vk.BufferImageCopy{{
		ImageSubresource: vk.ImageSubresourceLayers{AspectMask: GetFormatAspectFlags(image.ImageFormat), LayerCount: 1},
		ImageExtent:      vk.Extent3D{Width: width, Height: height, Depth: 1},
	}}
	vk.CmdCopyImageToBuffer(commandBuffer.CommandBuffer, image.Image, vk.ImageLayoutTransferSrcOptimal, buffer, uint32(len(bufferCopies)), bufferCopies)
	image.BarrierGeneralShaderAccess(commandBuffer)

	image.MarkSynced(frame)
	logger.Printf("[%s] downloaded image 0x%X (%dx%d) to RAM.\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)

	return func() {
		defer vk.DestroyBuffer(handles.Device, buffer, nil)
		defer vk.FreeMemory(handles.Device, bufferMem, nil)

		// Map staging buffer.
		memPtr := handles.MapMemory(bufferMem, copyBytes)
		defer vk.UnmapMemory(handles.Device, bufferMem)

		// Copy staging buffer to RAM.
		cpuSlice := unsafe.Slice((*byte)(unsafe.Pointer(image.Address)), guestBytes)
		swizzled := structs.SwizzleTexture(memPtr, image.FirstDescriptor)
		structs.GlobalMemoryManager.UpdateTraps(image.Address, image.GuestSize, 2)
		if len(swizzled) <= len(cpuSlice) {
			copy(cpuSlice, swizzled)
		} else {
			copy(cpuSlice, swizzled[:len(cpuSlice)])
		}
		structs.GlobalMemoryManager.UpdateTraps(image.Address, image.GuestSize, 0)
	}, nil
}

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
	} else {
		// Tiled guest layouts must be detiled using compute shader.
		mipLevel := int(image.FirstDescriptor.BaseLevel)
		if mipLevel >= len(image.Layouts) {
			return fmt.Errorf("mip level %d out of range", mipLevel)
		}
		layout := image.Layouts[mipLevel]
		width = uint32(layout.Pitch)
		height = uint32(layout.Height)
		bpp := structs.GetBytesPerPixel(image.FirstDescriptor.DataFormat)
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

		// Barrier and bind pipeline.
		vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
			vk.PipelineStageFlags(vk.PipelineStageHostBit),
			vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessHostWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit),
			}}, 0, nil, 0, nil)
		vk.CmdBindPipeline(commandBuffer.CommandBuffer, vk.PipelineBindPointCompute, pipeline.Pipeline)

		// Set both buffers.
		vk.CmdBindDescriptorSets(
			commandBuffer.CommandBuffer, vk.PipelineBindPointCompute,
			pipeline.PipelineLayout, spirvStructs.DescriptorSetSlotStatic,
			1, []vk.DescriptorSet{setToBind},
			0, nil,
		)
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

		// Push detile options.
		pitch := uint32(layout.Pitch)
		heightPush := uint32(layout.Height)
		if isBlock {
			pitch /= 4
			heightPush /= 4
		}
		c0 := pitch / 8
		c1 := c0 * ((heightPush + 7) / 8)
		pushConstants := make([]uint32, 5)
		pushConstants[0] = 0 // num_levels
		pushConstants[1] = pitch
		pushConstants[2] = heightPush
		pushConstants[3] = c0
		pushConstants[4] = c1
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

		// Pipeline for shader.
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
	vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageHostBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessHostWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
		}}, 0, nil, 0, nil)
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

func (image *VulkanImage) CopyToImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, dst *VulkanImage, frame uint64) error {
	image.SyncLock.Lock()
	dst.SyncLock.Lock()
	defer image.SyncLock.Unlock()
	defer dst.SyncLock.Unlock()

	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutTransferSrcOptimal,
		vk.AccessFlags(vk.AccessTransferReadBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.ImageAspectFlags(GetFormatAspectFlags(image.ImageFormat)))
	ImageBarrier(commandBuffer, dst,
		vk.ImageLayoutTransferDstOptimal,
		vk.AccessFlags(vk.AccessTransferWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.ImageAspectFlags(GetFormatAspectFlags(dst.ImageFormat)))

	mipCount := min(uint32(len(image.Layouts)), uint32(len(dst.Layouts)))

	var copies []vk.ImageCopy
	for mip := uint32(0); mip < mipCount; mip++ {
		copies = append(copies, vk.ImageCopy{
			SrcSubresource: vk.ImageSubresourceLayers{
				AspectMask: vk.ImageAspectFlags(GetFormatAspectFlags(image.ImageFormat)),
				MipLevel:   mip,
				LayerCount: 1,
			},
			DstSubresource: vk.ImageSubresourceLayers{
				AspectMask: vk.ImageAspectFlags(GetFormatAspectFlags(dst.ImageFormat)),
				MipLevel:   mip,
				LayerCount: 1,
			},
			Extent: vk.Extent3D{
				Width:  max(min(uint32(image.FirstDescriptor.Width)>>mip, uint32(dst.FirstDescriptor.Width)>>mip), 1),
				Height: max(min(uint32(image.FirstDescriptor.Height)>>mip, uint32(dst.FirstDescriptor.Height)>>mip), 1),
				Depth:  1,
			},
		})
	}

	vk.CmdCopyImage(commandBuffer.CommandBuffer,
		image.Image, vk.ImageLayoutTransferSrcOptimal,
		dst.Image, vk.ImageLayoutTransferDstOptimal,
		uint32(len(copies)), copies)

	dstAccess := vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit)
	if IsDepthFormat(dst.ImageFormat) {
		dstAccess = vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit | vk.AccessDepthStencilAttachmentWriteBit)
	}
	dstLayout := vk.ImageLayoutGeneral
	if IsDepthFormat(dst.ImageFormat) {
		dstLayout = vk.ImageLayoutDepthStencilAttachmentOptimal
	}
	ImageBarrier(commandBuffer, dst, dstLayout, dstAccess, vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit), vk.ImageAspectFlags(GetFormatAspectFlags(dst.ImageFormat)))

	dst.MarkSynced(frame)
	logger.Printf("[%s] copied image 0x%X (%dx%d) to 0x%X (%dx%d).\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
		dst.Address, dst.FirstDescriptor.Width, dst.FirstDescriptor.Height,
	)

	return nil
}

func (image *VulkanImage) NeedsRecreate(descriptor spirvStructs.ImageDescriptor, format vk.Format, requestIsSurface bool) (recreate bool, copyOld bool) {
	if image.ImageFormat != format {
		return true, false
	}
	stored := image.FirstDescriptor
	requestedBpp := structs.GetBytesPerPixel(descriptor.DataFormat)
	storedBpp := structs.GetBytesPerPixel(stored.DataFormat)
	requestedIsBlock := descriptor.DataFormat >= 35 && descriptor.DataFormat <= 41
	storedIsBlock := stored.DataFormat >= 35 && stored.DataFormat <= 41
	if descriptor.TilingIndex != stored.TilingIndex || requestedBpp != storedBpp || requestedIsBlock != storedIsBlock {
		return true, false
	}
	requestedSize := DescriptorGuestSize(descriptor)
	if requestedSize != image.GuestSize {
		return true, true
	}

	return false, false
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
