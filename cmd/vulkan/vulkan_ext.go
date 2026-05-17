package vulkan

import (
	"unsafe"

	vk "github.com/goki/vulkan"
)

const (
	StructureTypePhysicalDeviceSubgroupSizeControlPropertiesExt       vk.StructureType = 1000225000
	StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt vk.StructureType = 1000225001
	StructureTypePhysicalDeviceSubgroupSizeControlFeaturesExt         vk.StructureType = 1000225002
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
