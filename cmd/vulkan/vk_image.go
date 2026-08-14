package vulkan

import (
	"runtime"
	"sync"
	"unsafe"

	gcn2 "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
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
	if image.HasSync(ImageSyncCpuModified) {
		return
	}
	image.SetSync(ImageSyncCpuModified)
	image.ClearSync(ImageSyncGpuModified)
}

func (image *VulkanImage) MarkGpuModified(frame uint64) {
	image.SetSync(ImageSyncNeedsReadBarrier)
	if image.HasSync(ImageSyncGpuModified) {
		return
	}
	image.SetSync(ImageSyncGpuModified)
	image.ClearSync(ImageSyncCpuModified)
	if !image.IsSurface {
		structs.GlobalMemoryManager.UpdateTraps(image.Address, image.GuestSize, 0) // PROT_NONE
	}
}

func (image *VulkanImage) MarkSynced(frame uint64) {
	if !image.HasSync(ImageSyncCpuModified) && !image.HasSync(ImageSyncGpuModified) {
		return
	}
	image.ClearSync(ImageSyncCpuModified | ImageSyncGpuModified)
	if !image.IsSurface {
		structs.GlobalMemoryManager.UpdateTraps(image.Address, image.GuestSize, 1) // PROT_READ
	}
}

type VulkanImage struct {
	Address  uintptr
	Image    vk.Image
	ImageMem vk.DeviceMemory

	FirstDescriptor spirvStructs.ImageDescriptor
	ImageFormat     vk.Format
	ImageAspect     vk.ImageAspectFlags
	ImageUsage      vk.ImageUsageFlags

	ImageLayout vk.ImageLayout
	ImageAccess vk.AccessFlags
	ImageStage  vk.PipelineStageFlags

	IsSurface bool
	Layouts   []MipLayout
	GuestSize uintptr

	SyncFlags ImageSyncFlags
	SyncLock  sync.Mutex
}

type VulkanImageRequest struct {
	Descriptor spirvStructs.ImageDescriptor
	CompSwap   uint32
	IsSurface  bool
}

func CreateImage(handles *VulkanHandles, request VulkanImageRequest, commandBuffer *VulkanCommandBuffer, frame uint64) (*VulkanImage, error) {
	// Figure out format.
	format, _ := gcn.TranslateGcnFormat(request.Descriptor.DataFormat, request.Descriptor.NumFormat, request.CompSwap)

	// Figure out image flags.
	imageUsage := vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit)
	isBlock := request.Descriptor.DataFormat >= gcn2.GcnDataFormatBC1 && request.Descriptor.DataFormat <= gcn2.GcnDataFormatBC7
	if !isBlock {
		imageUsage |= vk.ImageUsageFlags(vk.ImageUsageStorageBit)
	}
	aspectMask := vk.ImageAspectFlags(vk.ImageAspectColorBit)
	dstLayout := vk.ImageLayoutGeneral
	dstAccess := vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit)
	if IsDepthFormat(format) {
		imageUsage = vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit)
		aspectMask = vk.ImageAspectFlags(vk.ImageAspectDepthBit | vk.ImageAspectStencilBit)
		dstLayout = vk.ImageLayoutDepthStencilAttachmentOptimal
		dstAccess = vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit | vk.AccessDepthStencilAttachmentWriteBit)
	}

	// Fix up format.
	if imageUsage&vk.ImageUsageFlags(vk.ImageUsageStorageBit) != 0 && request.Descriptor.NumFormat == gcn2.GcnNumFormatSrgb {
		format, _ = gcn.TranslateGcnFormat(request.Descriptor.DataFormat, gcn2.GcnNumFormatUnorm, request.CompSwap)
	}

	// Filter unsupported usage bits based on format properties.
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(handles.PhysicalDevice, format, &formatProps)
	formatProps.Deref()
	if (formatProps.OptimalTilingFeatures & vk.FormatFeatureFlags(vk.FormatFeatureStorageImageBit)) == 0 {
		if imageUsage&vk.ImageUsageFlags(vk.ImageUsageStorageBit) != 0 {
			logger.Printf("Failed assigning storage bit to format %d.\n", format)
		}
		imageUsage &^= vk.ImageUsageFlags(vk.ImageUsageStorageBit)
	}
	if (formatProps.OptimalTilingFeatures & vk.FormatFeatureFlags(vk.FormatFeatureColorAttachmentBit)) == 0 {
		if imageUsage&vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit) != 0 {
			logger.Printf("Failed assigning color attachment bit to format %d.\n", format)
		}
		imageUsage &^= vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit)
	}
	if (formatProps.OptimalTilingFeatures & vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit)) == 0 {
		if imageUsage&vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit) != 0 {
			logger.Printf("Failed assigning depth stencil attachment bit to format %d.\n", format)
		}
		imageUsage &^= vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit)
	}

	// Create our image.
	image := &VulkanImage{
		Address:         request.Descriptor.BaseAddress,
		FirstDescriptor: request.Descriptor,
		ImageFormat:     format,
		ImageLayout:     vk.ImageLayoutUndefined,
		ImageAspect:     aspectMask,
		ImageUsage:      imageUsage,
		ImageAccess:     vk.AccessFlags(vk.AccessNone),
		ImageStage:      vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit),
		IsSurface:       request.IsSurface,
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

	// Determine image type.
	imageType := vk.ImageType2d
	extentDepth := uint32(1)
	arrayLayers := uint32(1)
	switch request.Descriptor.InferredType() {
	case gcn2.GcnImageTypeBuffer, gcn2.GcnImageTypeColor1D:
		imageType = vk.ImageType1d
	case gcn2.GcnImageTypeColor3D:
		imageType = vk.ImageType3d
		extentDepth = uint32(request.Descriptor.Depth) + 1
	case gcn2.GcnImageTypeCubeOrArray:
		imageType = vk.ImageType2d
		arrayLayers = (uint32(request.Descriptor.Depth) + 1) * 6
	case gcn2.GcnImageTypeColor1DArray:
		imageType = vk.ImageType1d
		arrayLayers = uint32(request.Descriptor.Depth) + 1
	case gcn2.GcnImageTypeColor2DArray, gcn2.GcnImageTypeColor2DMsaaArray:
		imageType = vk.ImageType2d
		arrayLayers = uint32(request.Descriptor.Depth) + 1
	}

	createFlags := vk.ImageCreateFlags(vk.ImageCreateMutableFormatBit)
	if request.Descriptor.InferredType() == gcn2.GcnImageTypeCubeOrArray {
		createFlags |= vk.ImageCreateFlags(vk.ImageCreateCubeCompatibleBit)
	}

	// Create Vulkan image.
	result := vk.CreateImage(handles.Device, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		Flags:     createFlags,
		ImageType: imageType,
		Format:    format,
		Extent: vk.Extent3D{
			Width:  uint32(request.Descriptor.Width),
			Height: uint32(request.Descriptor.Height),
			Depth:  extentDepth,
		},
		MipLevels:     uint32(len(image.Layouts)),
		ArrayLayers:   arrayLayers,
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

		srcBlockExtent := uint32(1)
		if image.GetLinearDimensions().IsBlock {
			srcBlockExtent = 4
		}
		dstBlockExtent := uint32(1)
		if dst.GetLinearDimensions().IsBlock {
			dstBlockExtent = 4
		}

		srcBlocksWidth := max(srcWidth/srcBlockExtent, 1)
		srcBlocksHeight := max(srcHeight/srcBlockExtent, 1)
		dstBlocksWidth := max(dstWidth/dstBlockExtent, 1)
		dstBlocksHeight := max(dstHeight/dstBlockExtent, 1)

		blocksToCopyWidth := min(srcBlocksWidth, dstBlocksWidth)
		blocksToCopyHeight := min(srcBlocksHeight, dstBlocksHeight)

		extentWidth := blocksToCopyWidth * srcBlockExtent
		extentHeight := blocksToCopyHeight * srcBlockExtent
		extentDepth := min(max(uint32(image.FirstDescriptor.Depth), 1), max(uint32(dst.FirstDescriptor.Depth), 1))

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
				Width:  extentWidth,
				Height: extentHeight,
				Depth:  extentDepth,
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
	if logger.LogRenderer {
		logger.Printf("[%s] copied image 0x%X (%dx%d) to 0x%X (%dx%d).\n",
			color.Blue.Sprintf("Frame %d", frame),
			image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
			dst.Address, dst.FirstDescriptor.Width, dst.FirstDescriptor.Height,
		)
		logger.Printf("%+v\n", image.FirstDescriptor)
		logger.Printf("%+v\n", dst.FirstDescriptor)
	}

	return nil
}

func (image *VulkanImage) GetStagingBufferSize() vk.DeviceSize {
	dims := image.GetTiledDimensions(int(image.FirstDescriptor.BaseLevel))

	// Detile/retile shaders process full 8x8 micro-tiles.
	// If the height is unaligned, the shader will still write to the padded rows.
	paddedHeight := (dims.Height + 7) & ^uint32(7)
	size := vk.DeviceSize(dims.Width * paddedHeight * dims.Bpp)

	// Ensure the buffer is at least as large as the dispatched threads (groups * 64 texels).
	texels := dims.Width * dims.Height
	dispatchTexels := ((texels + 63) / 64) * 64
	dispatchSize := vk.DeviceSize(dispatchTexels * dims.Bpp)
	if dispatchSize > size {
		size = dispatchSize
	}

	return size
}

type ImageDimensions struct {
	Width     uint32
	Height    uint32
	Pitch     uint32
	RowPitch  uint32
	CopyBytes vk.DeviceSize
	Bpp       uint32
	IsBlock   bool
}

func (image *VulkanImage) GetLinearDimensions() ImageDimensions {
	isBlock := image.FirstDescriptor.DataFormat >= gcn2.GcnDataFormatBC1 && image.FirstDescriptor.DataFormat <= gcn2.GcnDataFormatBC7
	_, bpp := gcn.TranslateGcnFormat(image.FirstDescriptor.DataFormat, image.FirstDescriptor.NumFormat, 0)
	width := uint32(image.FirstDescriptor.Width)
	height := uint32(image.FirstDescriptor.Height)
	pitch := uint32(image.FirstDescriptor.Pitch)

	var copyBytes vk.DeviceSize
	var rowPitch uint32
	if isBlock {
		copyBytes = vk.DeviceSize(((width + 3) / 4) * ((height + 3) / 4) * bpp)
		rowPitch = (pitch + 3) / 4
	} else {
		copyBytes = vk.DeviceSize(width * height * bpp)
		rowPitch = pitch
		if rowPitch == 0 {
			rowPitch = width
		}
	}

	return ImageDimensions{
		Width:     width,
		Height:    height,
		Pitch:     pitch,
		RowPitch:  rowPitch,
		CopyBytes: copyBytes,
		Bpp:       bpp,
		IsBlock:   isBlock,
	}
}

func (image *VulkanImage) GetTiledDimensions(mipLevel int) ImageDimensions {
	isBlock := image.FirstDescriptor.DataFormat >= gcn2.GcnDataFormatBC1 && image.FirstDescriptor.DataFormat <= gcn2.GcnDataFormatBC7
	_, bpp := gcn.TranslateGcnFormat(image.FirstDescriptor.DataFormat, image.FirstDescriptor.NumFormat, 0)

	if mipLevel >= len(image.Layouts) {
		return ImageDimensions{}
	}
	layout := image.Layouts[mipLevel]
	pitch := uint32(layout.Pitch)

	width := pitch
	height := uint32(layout.Height)
	if isBlock {
		width /= 4
		height /= 4
	}

	return ImageDimensions{
		Width:   width,
		Height:  height,
		Pitch:   pitch,
		Bpp:     bpp,
		IsBlock: isBlock,
	}
}
