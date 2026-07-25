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
	FormatProperties         map[vk.Format]vk.FormatProperties
	SubgroupSizeProperties   VkPhysicalDeviceSubgroupSizeControlPropertiesEXT

	UploadPool vk.CommandPool
	FencePool  *VulkanFencePool2

	QueueMutex      *sync.Mutex
	UploadPoolMutex sync.Mutex

	DeferredDestroyMutex sync.Mutex
	DeferredDestroy      deferredDestroyQueue
}

// NewVulkanHandles extracts handles from the vulkan context and creates our upload command pool.
func NewVulkanHandles(context *VulkanContext, queueMutex *sync.Mutex) *VulkanHandles {
	vkh := &VulkanHandles{
		Context: context,
		Device:  context.Device,

		Instance:                 context.Instance,
		PhysicalDevice:           context.PhysicalDevice,
		GraphicsQueue:            context.GraphicsQueue,
		GraphicsQueueFamilyIndex: context.GraphicsQueueIndex,
		MemoryProperties:         context.MemoryProperties,
		FormatProperties:         map[vk.Format]vk.FormatProperties{},

		QueueMutex:      queueMutex,
		UploadPoolMutex: sync.Mutex{},

		DeferredDestroyMutex: sync.Mutex{},
	}
	vkh.MemoryProperties.Deref()

	vk.GetPhysicalDeviceProperties(vkh.PhysicalDevice, &vkh.DeviceProperties)
	vkh.DeviceProperties.Deref()
	for _, format := range []vk.Format{vk.FormatD16UnormS8Uint, vk.FormatD24UnormS8Uint, vk.FormatD32SfloatS8Uint} {
		var formatProperties vk.FormatProperties
		vk.GetPhysicalDeviceFormatProperties(vkh.PhysicalDevice, format, &formatProperties)
		formatProperties.Deref()
		vkh.FormatProperties[format] = formatProperties
	}

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

	pool, err := CreateCommandPool(vkh)
	if err != nil {
		panic(err)
	}
	vkh.UploadPool = pool

	fencePool, err := CreateFencePool2(vkh, 32)
	if err != nil {
		panic(err)
	}
	vkh.FencePool = fencePool

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

// MapMemory maps device memory and returns a Go byte slice over it.
func (vkh *VulkanHandles) MapMemory(mem vk.DeviceMemory, size vk.DeviceSize) []byte {
	var memPtr unsafe.Pointer
	vk.MapMemory(vkh.Device, mem, 0, size, 0, &memPtr)
	return (*[1 << 30]byte)(memPtr)[:size:size]
}
