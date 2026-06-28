package vulkan

import (
	"runtime"
	"unsafe"

	vk "github.com/goki/vulkan"
)

// VulkanHandles holds vulkan handles for lifetime of the process.
type VulkanHandles struct {
	Context *VulkanContext
	Device  vk.Device

	Instance                 vk.Instance
	PhysicalDevice           vk.PhysicalDevice
	GraphicsQueue            vk.Queue
	GraphicsQueueFamilyIndex uint32
	MemoryProperties         vk.PhysicalDeviceMemoryProperties
	DeviceProperties         vk.PhysicalDeviceProperties
	SubgroupSizeProperties   VkPhysicalDeviceSubgroupSizeControlPropertiesEXT

	// UploadPool is dedicated to our pre-render upload command buffers.
	UploadPool vk.CommandPool
}

// NewVulkanHandles extracts handles from the vulkan context and creates our upload command pool.
func NewVulkanHandles(context *VulkanContext) VulkanHandles {
	vkh := VulkanHandles{
		Context: context,
		Device:  context.Device,

		Instance:                 context.Instance,
		PhysicalDevice:           context.PhysicalDevice,
		GraphicsQueue:            context.GraphicsQueue,
		GraphicsQueueFamilyIndex: context.GraphicsQueueIndex,
		MemoryProperties:         context.MemoryProperties,
	}
	vkh.MemoryProperties.Deref()

	vk.GetPhysicalDeviceProperties(vkh.PhysicalDevice, &vkh.DeviceProperties)
	vkh.DeviceProperties.Deref()

	vkh.SubgroupSizeProperties = VkPhysicalDeviceSubgroupSizeControlPropertiesEXT{
		SType: StructureTypePhysicalDeviceSubgroupSizeControlPropertiesExt,
	}
	props2 := vk.PhysicalDeviceProperties2{
		SType: vk.StructureTypePhysicalDeviceProperties2,
		PNext: unsafe.Pointer(&vkh.SubgroupSizeProperties),
	}

	pinner := &runtime.Pinner{}
	pinner.Pin(&props2)
	pinner.Pin(&vkh.SubgroupSizeProperties)
	defer pinner.Unpin()

	GetPhysicalDeviceProperties2(vkh.Instance, vkh.PhysicalDevice, unsafe.Pointer(&props2))

	var pool vk.CommandPool
	result := vk.CreateCommandPool(vkh.Device, &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: vkh.GraphicsQueueFamilyIndex,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}, nil, &pool)
	if err := NewError(result); err != nil {
		panic(err)
	}
	vkh.UploadPool = pool

	return vkh
}

func (vkh *VulkanHandles) Destroy() {
	vkh.Context.Destroy()
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
