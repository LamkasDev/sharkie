package vulkan

import (
	"runtime"
	"sync"
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
	UploadPool  vk.CommandPool
	MainFence   vk.Fence
	WorkerFence vk.Fence
	QueueMutex  *sync.Mutex
}

// NewVulkanHandles extracts handles from the vulkan context and creates our upload command pool.
func NewVulkanHandles(context *VulkanContext, queueMutex *sync.Mutex) VulkanHandles {
	vkh := VulkanHandles{
		Context: context,
		Device:  context.Device,

		Instance:                 context.Instance,
		PhysicalDevice:           context.PhysicalDevice,
		GraphicsQueue:            context.GraphicsQueue,
		GraphicsQueueFamilyIndex: context.GraphicsQueueIndex,
		MemoryProperties:         context.MemoryProperties,
		QueueMutex:               queueMutex,
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

	pool, err := CreateCommandPool(&vkh)
	if err != nil {
		panic(err)
	}
	vkh.UploadPool = pool

	mainFence, err := CreateFence(&vkh)
	if err != nil {
		panic(err)
	}
	vkh.MainFence = mainFence
	workerFence, err := CreateFence(&vkh)
	if err != nil {
		panic(err)
	}
	vkh.WorkerFence = workerFence

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

// AllocateCommandBuffer allocates a single command buffer from pool.
func (vkh *VulkanHandles) AllocateCommandBuffer() vk.CommandBuffer {
	vkh.QueueMutex.Lock()
	defer vkh.QueueMutex.Unlock()
	buffers := make([]vk.CommandBuffer, 1)
	result := vk.AllocateCommandBuffers(vkh.Device, &vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        vkh.UploadPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}, buffers)
	if err := NewError(result); err != nil {
		panic(err)
	}

	return buffers[0]
}

// FreeCommandBuffer frees a command buffer from pool.
func (vkh *VulkanHandles) FreeCommandBuffer(commandBuffer vk.CommandBuffer) {
	vkh.QueueMutex.Lock()
	defer vkh.QueueMutex.Unlock()
	vk.FreeCommandBuffers(vkh.Device, vkh.UploadPool, 1, []vk.CommandBuffer{commandBuffer})
}

// MapMemory maps device memory and returns a Go byte slice over it.
func (vkh *VulkanHandles) MapMemory(mem vk.DeviceMemory, size vk.DeviceSize) []byte {
	var memPtr unsafe.Pointer
	vk.MapMemory(vkh.Device, mem, 0, size, 0, &memPtr)
	return (*[1 << 30]byte)(memPtr)[:size:size]
}
