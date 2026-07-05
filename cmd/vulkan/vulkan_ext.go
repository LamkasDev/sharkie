package vulkan

import (
	"unsafe"

	vk "github.com/goki/vulkan"
)

const (
	StructureTypePhysicalDeviceSubgroupSizeControlPropertiesExt       vk.StructureType = 1000225000
	StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt vk.StructureType = 1000225001
	StructureTypePhysicalDeviceSubgroupSizeControlFeaturesExt         vk.StructureType = 1000225002
	StructureTypeMemoryPriorityAllocateInfoExt                        vk.StructureType = 1000238001
	StructureTypePhysicalDevicePageableDeviceLocalMemoryFeaturesExt   vk.StructureType = 1000412000
	StructureTypePhysicalDeviceMemoryPriorityFeaturesExt              vk.StructureType = 1000238000
)

type VkPhysicalDeviceSubgroupSizeControlPropertiesEXT struct {
	SType                        vk.StructureType
	PNext                        unsafe.Pointer
	MinSubgroupSize              uint32
	MaxSubgroupSize              uint32
	MaxComputeWorkgroupSubgroups uint32
	RequiredSubgroupSizeStages   vk.ShaderStageFlags
}

type VkPhysicalDeviceSubgroupSizeControlFeaturesEXT struct {
	SType                vk.StructureType
	PNext                unsafe.Pointer
	SubgroupSizeControl  vk.Bool32
	ComputeFullSubgroups vk.Bool32
}

type VkPipelineShaderStageRequiredSubgroupSizeCreateInfoEXT struct {
	SType                vk.StructureType
	PNext                unsafe.Pointer
	RequiredSubgroupSize uint32
}

type VkMemoryPriorityAllocateInfoEXT struct {
	SType    vk.StructureType
	PNext    unsafe.Pointer
	Priority float32
}

type VkPhysicalDevicePageableDeviceLocalMemoryFeaturesEXT struct {
	SType                     vk.StructureType
	PNext                     unsafe.Pointer
	PageableDeviceLocalMemory uint32
}

type VkPhysicalDeviceMemoryPriorityFeaturesEXT struct {
	SType          vk.StructureType
	PNext          unsafe.Pointer
	MemoryPriority vk.Bool32
}
