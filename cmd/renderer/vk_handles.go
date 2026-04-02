package renderer

import (
	"unsafe"

	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
)

// VulkanHandles holds vulkan handles for lifetime of the process.
type VulkanHandles struct {
	Context as.Context
	Device  vk.Device

	Platform                 as.Platform
	Instance                 vk.Instance
	PhysicalDevice           vk.PhysicalDevice
	GraphicsQueue            vk.Queue
	GraphicsQueueFamilyIndex uint32
	MemoryProperties         vk.PhysicalDeviceMemoryProperties

	// UploadPool is dedicated to our pre-render upload command buffers.
	UploadPool vk.CommandPool
}

// NewVulkanHandles extracts handles from the asche context and creates our upload command pool.
func NewVulkanHandles(context as.Context) VulkanHandles {
	vkh := VulkanHandles{
		Context: context,
		Device:  context.Device(),

		Platform:                 context.Platform(),
		Instance:                 context.Platform().Instance(),
		PhysicalDevice:           context.Platform().PhysicalDevice(),
		GraphicsQueue:            context.Platform().GraphicsQueue(),
		GraphicsQueueFamilyIndex: context.Platform().GraphicsQueueFamilyIndex(),
		MemoryProperties:         context.Platform().MemoryProperties(),
	}
	vkh.MemoryProperties.Deref()

	var pool vk.CommandPool
	result := vk.CreateCommandPool(vkh.Device, &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: vkh.GraphicsQueueFamilyIndex,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}, nil, &pool)
	if err := as.NewError(result); err != nil {
		panic(err)
	}
	vkh.UploadPool = pool

	return vkh
}

func (vkh *VulkanHandles) Destroy() {
	vkh.Platform.Destroy()
}

// FindMemoryType returns index of a memory type that satisfies typeFilter and has all required property flags set.
func (vkh *VulkanHandles) FindMemoryType(typeFilter uint32, props vk.MemoryPropertyFlagBits) uint32 {
	for i := range vkh.MemoryProperties.MemoryTypeCount {
		memoryType := vkh.MemoryProperties.MemoryTypes[i]
		memoryType.Deref()
		if (typeFilter&(1<<i)) != 0 && vk.MemoryPropertyFlagBits(memoryType.PropertyFlags)&props == props {
			return i
		}
	}

	return 0
}

// AllocateCommandBuffer allocates a single primary command buffer from pool.
func (vkh *VulkanHandles) AllocateCommandBuffer(pool vk.CommandPool) vk.CommandBuffer {
	buffers := make([]vk.CommandBuffer, 1)
	vk.AllocateCommandBuffers(vkh.Device, &vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        pool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}, buffers)
	return buffers[0]
}

// MapMemory maps device memory and returns a Go byte slice over it.
func (vkh *VulkanHandles) MapMemory(mem vk.DeviceMemory, size vk.DeviceSize) []byte {
	var memPtr unsafe.Pointer
	vk.MapMemory(vkh.Device, mem, 0, size, 0, &memPtr)
	return (*[1 << 30]byte)(memPtr)[:size:size]
}
