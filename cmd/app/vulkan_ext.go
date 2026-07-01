package app

import (
	"unsafe"

	vk "github.com/goki/vulkan"
)

const (
	StructureTypePhysicalDeviceSubgroupSizeControlFeaturesExt         vk.StructureType = 1000225000
	StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt vk.StructureType = 1000225001
)

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
