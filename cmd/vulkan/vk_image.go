package vulkan

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	vk "github.com/goki/vulkan"
)

type ImageSyncFlags uint8

const (
	ImageSyncGpuModified      ImageSyncFlags = 1 << 0
	ImageSyncCpuDirty         ImageSyncFlags = 1 << 1
	ImageSyncGuestUploaded    ImageSyncFlags = 1 << 2
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

type VulkanImage struct {
	Address  uintptr
	Image    vk.Image
	imageMem vk.DeviceMemory

	FirstDescriptor spirvStructs.ImageDescriptor
	ImageFormat     vk.Format
	imageAspect     vk.ImageAspectFlags

	imageLayout vk.ImageLayout
	imageAccess vk.AccessFlags
	imageStage  vk.PipelineStageFlags

	IsSurface    bool
	SyncFlags    ImageSyncFlags
	MirrorSynced bool
	Generation   uint32

	GuestSize       uintptr
	SyncMarkedFrame uint64
	FrameTouched    uint64
}

type VulkanImageRequest struct {
	Descriptor spirvStructs.ImageDescriptor
	Format     vk.Format
	IsSurface  bool
}

func CreateImage(handles *VulkanHandles, request VulkanImageRequest, commandBuffer vk.CommandBuffer) (*VulkanImage, error) {
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
		imageLayout:     vk.ImageLayoutUndefined,
		imageAspect:     aspectMask,
		imageAccess:     vk.AccessFlags(vk.AccessNone),
		imageStage:      vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit),
		IsSurface:       request.IsSurface,
		Generation:      1,
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
	SetDebugUtilsObjectName(handles.Instance, handles.Device, vk.ObjectTypeImage, uint64(uintptr(unsafe.Pointer(image))), fmt.Sprintf("2D Image 0x%X", request.Descriptor.BaseAddress))

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
	}, nil, &image.imageMem)
	if err := NewError(result); err != nil {
		return nil, err
	}
	SetDeviceMemoryPriority(handles.Instance, handles.Device, image.imageMem, 1.0)
	vk.BindImageMemory(handles.Device, image.Image, image.imageMem, 0)

	// Transition image.
	err := RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer vk.CommandBuffer) {
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
	vk.FreeMemory(device, image.imageMem, nil)
	image.imageMem = vk.NullDeviceMemory
}

func (image *VulkanImage) BarrierShaderRead(commandBuffer vk.CommandBuffer) {
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

func (image *VulkanImage) BarrierColorAttachment(commandBuffer vk.CommandBuffer) {
	if IsDepthFormat(image.ImageFormat) {
		return
	}
	if image.imageLayout != vk.ImageLayoutGeneral {
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

func (image *VulkanImage) BarrierSampledRead(commandBuffer vk.CommandBuffer) {
	if image.imageLayout == vk.ImageLayoutGeneral {
		return
	}
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessShaderReadBit),
		vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit|vk.PipelineStageComputeShaderBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierComputeStorageWrite(commandBuffer vk.CommandBuffer) {
	if image.imageLayout == vk.ImageLayoutGeneral {
		return
	}
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierTransferSrc(commandBuffer vk.CommandBuffer) {
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutTransferSrcOptimal,
		vk.AccessFlags(vk.AccessTransferReadBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierTransferWrite(commandBuffer vk.CommandBuffer) {
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessTransferWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) BarrierGeneralShaderAccess(commandBuffer vk.CommandBuffer) {
	ImageBarrier(commandBuffer, image,
		vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit),
		GetFormatAspectFlags(image.ImageFormat),
	)
}

func (image *VulkanImage) DownloadFromVkImage(handles *VulkanHandles) error {
	// Downloading modified surfaces is too expensive.
	if image.IsSurface {
		image.MirrorSynced = true
		return nil
	}
	if IsDepthFormat(image.ImageFormat) {
		image.MirrorSynced = true
		return nil
	}
	if image.MirrorSynced {
		return nil
	}

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
	err = RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer vk.CommandBuffer) {
		image.BarrierTransferSrc(commandBuffer)
		bufferCopies := []vk.BufferImageCopy{{
			ImageSubresource: vk.ImageSubresourceLayers{AspectMask: GetFormatAspectFlags(image.ImageFormat), LayerCount: 1},
			ImageExtent:      vk.Extent3D{Width: width, Height: height, Depth: 1},
		}}
		vk.CmdCopyImageToBuffer(commandBuffer, image.Image, vk.ImageLayoutTransferSrcOptimal, buffer, uint32(len(bufferCopies)), bufferCopies)
		image.BarrierGeneralShaderAccess(commandBuffer)
	})
	if err != nil {
		return err
	}

	memPtr := handles.MapMemory(bufferMem, copyBytes)
	defer vk.UnmapMemory(handles.Device, bufferMem)

	// Restore fault handlers.
	pageMask := uintptr(lib_structs.SystemPageSize - 1)
	alignedAddress := image.Address &^ pageMask
	alignedSize := (guestBytes + (image.Address - alignedAddress) + pageMask) &^ pageMask

	mprotectSlice := unsafe.Slice((*byte)(unsafe.Pointer(alignedAddress)), alignedSize)
	_ = syscall.Mprotect(mprotectSlice, syscall.PROT_READ|syscall.PROT_WRITE)

	// Copy staging buffer to RAM.
	cpuSlice := unsafe.Slice((*byte)(unsafe.Pointer(image.Address)), guestBytes)
	swizzled := structs.SwizzleTexture(memPtr, image.FirstDescriptor)
	if len(swizzled) <= len(cpuSlice) {
		copy(cpuSlice, swizzled)
	} else {
		copy(cpuSlice, swizzled[:len(cpuSlice)])
	}
	image.MirrorSynced = true

	return nil
}

func (image *VulkanImage) UploadToVkImage(handles *VulkanHandles, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error)) error {
	// Uploading surfaces is too expensive.
	if image.IsSurface {
		return nil
	}
	if IsDepthFormat(image.ImageFormat) {
		return nil
	}

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
	err := RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer vk.CommandBuffer) {
		image.BarrierTransferWrite(commandBuffer)
		vk.CmdCopyBufferToImage(commandBuffer, srcBuffer, image.Image, vk.ImageLayoutGeneral, 1, []vk.BufferImageCopy{{
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

	return nil
}

func (image *VulkanImage) CopyToImage(handles *VulkanHandles, dst *VulkanImage) error {
	err := RunWithCommandBuffer(handles, handles.WorkerFence, func(commandBuffer vk.CommandBuffer) {
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

		vk.CmdCopyImage(commandBuffer,
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
				Extent: vk.Extent3D{Width: uint32(image.FirstDescriptor.Width), Height: uint32(image.FirstDescriptor.Height), Depth: 1},
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

	return err
}

func (image *VulkanImage) NeedsRecreate(descriptor spirvStructs.ImageDescriptor, format vk.Format, requestIsSurface bool) bool {
	if image.ImageFormat != format {
		return true
	}
	stored := image.FirstDescriptor
	if descriptor.Width > stored.Width || descriptor.Height > stored.Height {
		return true
	}
	if image.IsSurface && !requestIsSurface {
		return false
	}
	if region := DescriptorRegionSize(descriptor); region > image.GuestSize {
		return true
	}

	return false
}

func (image *VulkanImage) ShouldUploadToVkImage() bool {
	// Uploading surfaces is too expensive.
	if image.IsSurface {
		return false
	}
	regionSize := DescriptorRegionSize(image.FirstDescriptor)
	if regionSize == 0 {
		bpp := structs.GetBytesPerPixel(image.FirstDescriptor.DataFormat)
		regionSize = uintptr(image.FirstDescriptor.Pitch*image.FirstDescriptor.Height) * uintptr(bpp)
	}

	// Check dirty flags.
	if image.HasSync(ImageSyncCpuDirty) {
		return true
	}
	if structs.GlobalMemoryManager.IsRegionCpuModified(image.Address, regionSize) {
		return true
	}
	if image.HasSync(ImageSyncGpuModified) && !image.HasSync(ImageSyncGuestUploaded) {
		return false
	}
	if !image.HasSync(ImageSyncGuestUploaded) {
		return true
	}
	if image.HasSync(ImageSyncGpuModified) {
		return false
	}

	return false
}
