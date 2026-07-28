package vulkan

import (
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
	err := RunWithCommandBuffer(handles.GraphicsQueue, handles, func(commandBuffer *VulkanCommandBuffer) {
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

func (image *VulkanImage) CopyToImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, dst *VulkanImage, frame uint64) error {
	image.SyncLock.Lock()
	dst.SyncLock.Lock()
	defer image.SyncLock.Unlock()
	defer dst.SyncLock.Unlock()

	image.BarrierTransferSrc(commandBuffer)
	dst.BarrierTransferWrite(commandBuffer)

	var copies []vk.ImageCopy
	mipCount := min(uint32(len(image.Layouts)), uint32(len(dst.Layouts)))
	for mip := uint32(0); mip < mipCount; mip++ {
		srcWidth := max(uint32(image.FirstDescriptor.Width)>>mip, 1)
		srcHeight := max(uint32(image.FirstDescriptor.Height)>>mip, 1)
		dstWidth := max(uint32(dst.FirstDescriptor.Width)>>mip, 1)
		dstHeight := max(uint32(dst.FirstDescriptor.Height)>>mip, 1)
		copies = append(copies, vk.ImageCopy{
			SrcSubresource: vk.ImageSubresourceLayers{
				AspectMask: GetFormatAspectFlags(image.ImageFormat),
				MipLevel:   mip,
				LayerCount: 1,
			},
			DstSubresource: vk.ImageSubresourceLayers{
				AspectMask: GetFormatAspectFlags(dst.ImageFormat),
				MipLevel:   mip,
				LayerCount: 1,
			},
			Extent: vk.Extent3D{
				Width:  min(srcWidth, dstWidth),
				Height: min(srcHeight, dstHeight),
				Depth:  1,
			},
		})
	}

	vk.CmdCopyImage(commandBuffer.CommandBuffer,
		image.Image, vk.ImageLayoutTransferSrcOptimal,
		dst.Image, vk.ImageLayoutGeneral,
		uint32(len(copies)), copies)

	image.BarrierGeneralShaderAccess(commandBuffer)
	if IsDepthFormat(dst.ImageFormat) {
		ImageBarrier(commandBuffer, dst,
			vk.ImageLayoutDepthStencilAttachmentOptimal,
			vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit|vk.AccessDepthStencilAttachmentWriteBit),
			vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit),
			GetFormatAspectFlags(dst.ImageFormat),
		)
	} else {
		dst.BarrierGeneralShaderAccess(commandBuffer)
	}

	dst.MarkSynced(frame)
	logger.Printf("[%s] copied image 0x%X (%dx%d) to 0x%X (%dx%d).\n",
		color.Blue.Sprintf("Frame %d", frame),
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
		dst.Address, dst.FirstDescriptor.Width, dst.FirstDescriptor.Height,
	)

	return nil
}

func (image *VulkanImage) NeedsRecreate(descriptor spirvStructs.ImageDescriptor, format vk.Format, requestIsSurface bool) bool {
	if image.ImageFormat != format {
		return true
	}
	stored := image.FirstDescriptor
	requestedBpp := structs.GetBytesPerPixel(descriptor.DataFormat)
	storedBpp := structs.GetBytesPerPixel(stored.DataFormat)
	requestedIsBlock := descriptor.DataFormat >= 35 && descriptor.DataFormat <= 41
	storedIsBlock := stored.DataFormat >= 35 && stored.DataFormat <= 41
	if descriptor.TilingIndex != stored.TilingIndex || requestedBpp != storedBpp || requestedIsBlock != storedIsBlock {
		return true
	}
	/* if descriptor.Width != stored.Width || descriptor.Height != stored.Height || descriptor.Pitch != stored.Pitch {
		return true, false
	} */
	requestedSize := DescriptorGuestSize(descriptor)
	if requestedSize != image.GuestSize {
		return true
	}

	return false
}
