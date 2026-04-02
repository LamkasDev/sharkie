package renderer

import (
	"sync/atomic"
	"unsafe"

	"github.com/elokore/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/elokore/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/elokore/cimgui-go-vulkan/imgui"
	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
)

// FramebufferTexture is a persistent graphics surface for displaying the guest framebuffer.
type FramebufferTexture struct {
	// TextureId is the descriptor set registered with ImGui's Vulkan backend.
	TextureId imgui.TextureRef

	width, height uint32
	byteSize      vk.DeviceSize

	// Graphics card texture resources.
	image     vk.Image
	imageMem  vk.DeviceMemory
	imageView vk.ImageView
	sampler   vk.Sampler

	// Double-buffers for writing and uploading frame data.
	buffers       [2]vk.Buffer
	buffersMemory [2]vk.DeviceMemory
	buffersPtr    [2][]byte

	nextWriteIdx     int
	pendingUploadIdx atomic.Int32
	firstUpload      bool
}

// NewFramebufferTexture creates and allocates all resources for a new FramebufferTexture.
func NewFramebufferTexture(vkh *VulkanHandles, backend backend.Backend[glfwvulkanbackend.GLFWWindowFlags], width, height uint32) (*FramebufferTexture, error) {
	texture := &FramebufferTexture{
		width:       width,
		height:      height,
		byteSize:    vk.DeviceSize(width) * vk.DeviceSize(height) * 4,
		firstUpload: true,
	}
	texture.pendingUploadIdx.Store(-1)

	if err := texture.allocImage(vkh); err != nil {
		return nil, err
	}
	if err := texture.allocSampler(vkh); err != nil {
		return nil, err
	}
	if err := texture.allocBuffers(vkh); err != nil {
		return nil, err
	}
	texture.TextureId = backend.CreateVulkanTexture(texture.sampler, texture.imageView, vk.ImageLayoutShaderReadOnlyOptimal)

	return texture, nil
}

// Destroy frees all Vulkan resources owned by this FramebufferTexture.
func (t *FramebufferTexture) Destroy(vkh *VulkanHandles) {
	vk.DeviceWaitIdle(vkh.Device)
	if t.imageView != vk.NullImageView {
		vk.DestroyImageView(vkh.Device, t.imageView, nil)
	}
	if t.sampler != vk.NullSampler {
		vk.DestroySampler(vkh.Device, t.sampler, nil)
	}
	if t.image != vk.NullImage {
		vk.DestroyImage(vkh.Device, t.image, nil)
	}
	if t.imageMem != vk.NullDeviceMemory {
		vk.FreeMemory(vkh.Device, t.imageMem, nil)
	}
	for bufferIndex := range t.buffers {
		if t.buffers[bufferIndex] != vk.NullBuffer {
			vk.DestroyBuffer(vkh.Device, t.buffers[bufferIndex], nil)
		}
		if t.buffersMemory[bufferIndex] != vk.NullDeviceMemory {
			vk.UnmapMemory(vkh.Device, t.buffersMemory[bufferIndex])
			vk.FreeMemory(vkh.Device, t.buffersMemory[bufferIndex], nil)
		}
	}
}

// WritePixels copies an RGBA frame into the current staging buffer.
func (t *FramebufferTexture) WritePixels(pixels []byte) {
	copy(t.buffersPtr[t.nextWriteIdx], pixels)
	t.pendingUploadIdx.Store(int32(t.nextWriteIdx))
	t.nextWriteIdx ^= 1
}

// UploadPending submits a command buffer that copies a pending frame to graphics memory.
// Must be called from the main thread, BEFORE currentBackend.NewFrame().
func (t *FramebufferTexture) UploadPending(vkh *VulkanHandles) {
	srcIdx := int(t.pendingUploadIdx.Swap(-1))
	if srcIdx < 0 {
		return
	}

	// Allocate a new command buffer for upload.
	cmdBuffer := vkh.AllocateCommandBuffer(vkh.UploadPool)
	vk.BeginCommandBuffer(cmdBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})

	// Prepare destination memory for write.
	oldLayout := vk.ImageLayoutShaderReadOnlyOptimal
	oldAccess := vk.AccessFlags(vk.AccessShaderReadBit)
	if t.firstUpload {
		oldLayout = vk.ImageLayoutUndefined
		oldAccess = 0
		t.firstUpload = false
	}
	vk.CmdPipelineBarrier(cmdBuffer,
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           oldLayout,
			NewLayout:           vk.ImageLayoutTransferDstOptimal,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.image,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1, LayerCount: 1,
			},
			SrcAccessMask: oldAccess,
			DstAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
		}})

	// Copy staging buffer to graphics memory.
	vk.CmdCopyBufferToImage(cmdBuffer,
		t.buffers[srcIdx], t.image,
		vk.ImageLayoutTransferDstOptimal,
		1, []vk.BufferImageCopy{{
			ImageSubresource: vk.ImageSubresourceLayers{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LayerCount: 1,
			},
			ImageExtent: vk.Extent3D{Width: t.width, Height: t.height, Depth: 1},
		}})

	// End write and set destination memory as read-only.
	vk.CmdPipelineBarrier(cmdBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferDstOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.image,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1, LayerCount: 1,
			},
			SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit),
		}})
	vk.EndCommandBuffer(cmdBuffer)

	// Wait for the upload to finish before imgui uses the texture.
	vk.QueueSubmit(vkh.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cmdBuffer},
	}}, vk.NullFence)
	vk.QueueWaitIdle(vkh.GraphicsQueue)
	vk.FreeCommandBuffers(vkh.Device, vkh.UploadPool, 1, []vk.CommandBuffer{cmdBuffer})
}

func (t *FramebufferTexture) allocImage(vkh *VulkanHandles) error {
	result := vk.CreateImage(vkh.Device, &vk.ImageCreateInfo{
		SType:         vk.StructureTypeImageCreateInfo,
		ImageType:     vk.ImageType2d,
		Format:        vk.FormatR8g8b8a8Unorm,
		Extent:        vk.Extent3D{Width: t.width, Height: t.height, Depth: 1},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        vk.ImageTilingOptimal,
		Usage:         vk.ImageUsageFlags(vk.ImageUsageTransferDstBit | vk.ImageUsageSampledBit),
		SharingMode:   vk.SharingModeExclusive,
		InitialLayout: vk.ImageLayoutUndefined,
	}, nil, &t.image)
	if err := as.NewError(result); err != nil {
		return err
	}

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(vkh.Device, t.image, &memReqs)
	memReqs.Deref()

	result = vk.AllocateMemory(vkh.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: vkh.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyDeviceLocalBit),
	}, nil, &t.imageMem)
	if err := as.NewError(result); err != nil {
		return err
	}
	vk.BindImageMemory(vkh.Device, t.image, t.imageMem, 0)

	result = vk.CreateImageView(vkh.Device, &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    t.image,
		ViewType: vk.ImageViewType2d,
		Format:   vk.FormatR8g8b8a8Unorm,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			LevelCount: 1, LayerCount: 1,
		},
	}, nil, &t.imageView)

	return as.NewError(result)
}

func (t *FramebufferTexture) allocSampler(vkh *VulkanHandles) error {
	result := vk.CreateSampler(vkh.Device, &vk.SamplerCreateInfo{
		SType:        vk.StructureTypeSamplerCreateInfo,
		MagFilter:    vk.FilterLinear,
		MinFilter:    vk.FilterLinear,
		AddressModeU: vk.SamplerAddressModeClampToEdge,
		AddressModeV: vk.SamplerAddressModeClampToEdge,
		AddressModeW: vk.SamplerAddressModeClampToEdge,
		BorderColor:  vk.BorderColorFloatOpaqueWhite,
	}, nil, &t.sampler)

	return as.NewError(result)
}

func (t *FramebufferTexture) allocBuffers(vkh *VulkanHandles) error {
	for bufferIndex := range t.buffers {
		result := vk.CreateBuffer(vkh.Device, &vk.BufferCreateInfo{
			SType:       vk.StructureTypeBufferCreateInfo,
			Size:        t.byteSize,
			Usage:       vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
			SharingMode: vk.SharingModeExclusive,
		}, nil, &t.buffers[bufferIndex])
		if err := as.NewError(result); err != nil {
			return err
		}

		var memReqs vk.MemoryRequirements
		vk.GetBufferMemoryRequirements(vkh.Device, t.buffers[bufferIndex], &memReqs)
		memReqs.Deref()

		result = vk.AllocateMemory(vkh.Device, &vk.MemoryAllocateInfo{
			SType:          vk.StructureTypeMemoryAllocateInfo,
			AllocationSize: memReqs.Size,
			MemoryTypeIndex: vkh.FindMemoryType(memReqs.MemoryTypeBits,
				vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		}, nil, &t.buffersMemory[bufferIndex])
		if err := as.NewError(result); err != nil {
			return err
		}
		vk.BindBufferMemory(vkh.Device, t.buffers[bufferIndex], t.buffersMemory[bufferIndex], 0)

		var memPtr unsafe.Pointer
		vk.MapMemory(vkh.Device, t.buffersMemory[bufferIndex], 0, t.byteSize, 0, &memPtr)
		t.buffersPtr[bufferIndex] = (*[1 << 30]byte)(memPtr)[:t.byteSize:t.byteSize]
	}

	return nil
}
