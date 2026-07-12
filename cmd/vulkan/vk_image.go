package vulkan

import (
	"fmt"
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
	GuestSize  uintptr

	SyncFlags ImageSyncFlags
	SyncLock  sync.Mutex
}

type VulkanImageRequest struct {
	Descriptor spirvStructs.ImageDescriptor
	Format     vk.Format
	IsSurface  bool
}

func CreateImage(handles *VulkanHandles, request VulkanImageRequest, commandBuffer *VulkanCommandBuffer) (*VulkanImage, error) {
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
		SyncLock:        sync.Mutex{},
		GuestSize:       DescriptorRegionSize(request.Descriptor),
	}

	// Create VkImage.
	result := vk.CreateImage(handles.Device, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    request.Format,
		Extent: vk.Extent3D{
			Width:  uint32(request.Descriptor.Width),
			Height: uint32(request.Descriptor.Height),
			Depth:  uint32(request.Descriptor.Depth),
		},
		MipLevels:     uint32(mipLevelCount(request.Descriptor)),
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
	err := RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer *VulkanCommandBuffer) {
		ImageBarrier(commandBuffer, image, dstLayout, dstAccess, vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit), aspectMask)
	})
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

func (image *VulkanImage) DownloadFromVkImage(handles *VulkanHandles, frame uint64) error {
	image.SyncLock.Lock()
	defer image.SyncLock.Unlock()

	// Prepare download parameters.
	width := uint32(image.FirstDescriptor.Width)
	height := uint32(image.FirstDescriptor.Height)
	pitch := uint32(image.FirstDescriptor.Pitch)
	bpp := structs.GetBytesPerPixel(image.FirstDescriptor.DataFormat)

	copyBytes := vk.DeviceSize(width * height * bpp)
	guestBytes := uintptr(pitch) * uintptr(height) * uintptr(bpp)

	// Allocate staging buffer.
	buffer, bufferMem, err := AllocateBuffer(handles, copyBytes, vk.BufferUsageFlags(vk.BufferUsageTransferDstBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return fmt.Errorf("staging buffer: %w", err)
	}
	defer vk.DestroyBuffer(handles.Device, buffer, nil)
	defer vk.FreeMemory(handles.Device, bufferMem, nil)

	// Copy image to staging buffer.
	err = RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer *VulkanCommandBuffer) {
		image.BarrierTransferSrc(commandBuffer)
		bufferCopies := []vk.BufferImageCopy{{
			ImageSubresource: vk.ImageSubresourceLayers{AspectMask: GetFormatAspectFlags(image.ImageFormat), LayerCount: 1},
			ImageExtent:      vk.Extent3D{Width: width, Height: height, Depth: 1},
		}}
		vk.CmdCopyImageToBuffer(commandBuffer.CommandBuffer, image.Image, vk.ImageLayoutTransferSrcOptimal, buffer, uint32(len(bufferCopies)), bufferCopies)
		image.BarrierGeneralShaderAccess(commandBuffer)
	})
	if err != nil {
		return err
	}

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

	image.MarkSynced(frame)
	logger.Printf("[%s] downloaded image 0x%X (%dx%d) to RAM.\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)

	return nil
}

func (image *VulkanImage) UploadToVkImage(handles *VulkanHandles, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) error {
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
		// Tiled guest layouts must be detiled, then staged for transfer.
		linear, layout, err := DetileGuestTexture(image.FirstDescriptor)
		if err != nil {
			return err
		}
		size := vk.DeviceSize(len(linear))
		buf, mem, err := AllocateBuffer(handles, size,
			vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
		if err != nil {
			return err
		}
		defer vk.DestroyBuffer(handles.Device, buf, nil)
		defer vk.FreeMemory(handles.Device, mem, nil)

		ptr := handles.MapMemory(mem, size)
		copy(ptr, linear)
		vk.UnmapMemory(handles.Device, mem)

		srcBuffer = buf
		width = uint32(layout.Width)
		height = uint32(layout.Height)
	}

	// Copy staging buffer to GPU.
	err := RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer *VulkanCommandBuffer) {
		image.BarrierTransferWrite(commandBuffer)
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
			ImageExtent: vk.Extent3D{Width: width, Height: height, Depth: 1},
		}})
		image.BarrierGeneralShaderAccess(commandBuffer)
	})
	if err != nil {
		return err
	}

	image.MarkSynced(frame)
	logger.Printf("[%s] uploaded image 0x%X (%dx%d) to VRAM.\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)

	return nil
}

func (image *VulkanImage) CopyToImage(handles *VulkanHandles, dst *VulkanImage, frame uint64) error {
	image.SyncLock.Lock()
	dst.SyncLock.Lock()
	defer image.SyncLock.Unlock()
	defer dst.SyncLock.Unlock()

	err := RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer *VulkanCommandBuffer) {
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

		vk.CmdCopyImage(commandBuffer.CommandBuffer,
			image.Image, vk.ImageLayoutTransferSrcOptimal,
			dst.Image, vk.ImageLayoutTransferDstOptimal,
			1, []vk.ImageCopy{{
				SrcSubresource: vk.ImageSubresourceLayers{
					AspectMask: vk.ImageAspectFlags(GetFormatAspectFlags(image.ImageFormat)),
					MipLevel:   0,
					LayerCount: 1,
				},
				DstSubresource: vk.ImageSubresourceLayers{
					AspectMask: vk.ImageAspectFlags(GetFormatAspectFlags(dst.ImageFormat)),
					MipLevel:   0,
					LayerCount: 1,
				},
				Extent: vk.Extent3D{
					Width:  min(uint32(image.FirstDescriptor.Width), uint32(dst.FirstDescriptor.Width)),
					Height: min(uint32(image.FirstDescriptor.Height), uint32(dst.FirstDescriptor.Height)),
					Depth:  1,
				},
			}})

		dstAccess := vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit)
		if IsDepthFormat(dst.ImageFormat) {
			dstAccess = vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit | vk.AccessDepthStencilAttachmentWriteBit)
		}
		dstLayout := vk.ImageLayoutGeneral
		if IsDepthFormat(dst.ImageFormat) {
			dstLayout = vk.ImageLayoutDepthStencilAttachmentOptimal
		}
		ImageBarrier(commandBuffer, dst, dstLayout, dstAccess, vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit), vk.ImageAspectFlags(GetFormatAspectFlags(dst.ImageFormat)))
	})
	if err != nil {
		return err
	}

	// TODO: flags?
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
	requestedSize := DescriptorRegionSize(descriptor)
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
