package vulkan

import (
	"fmt"

	vk "github.com/goki/vulkan"
)

func CreateDescriptorPool(handles *VulkanHandles, staticLayout vk.DescriptorSetLayout) (vk.DescriptorPool, vk.DescriptorSet, error) {
	var pool vk.DescriptorPool
	result := vk.CreateDescriptorPool(handles.Device, &vk.DescriptorPoolCreateInfo{
		SType: vk.StructureTypeDescriptorPoolCreateInfo,
		PPoolSizes: []vk.DescriptorPoolSize{
			{
				Type:            vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount: 8192,
			},
			{
				Type:            vk.DescriptorTypeStorageImage,
				DescriptorCount: 8192,
			},
			{
				Type:            vk.DescriptorTypeStorageBuffer,
				DescriptorCount: 256,
			},
		},
		PoolSizeCount: 3,
		MaxSets:       8192,
		Flags:         vk.DescriptorPoolCreateFlags(vk.DescriptorPoolCreateFreeDescriptorSetBit | vk.DescriptorPoolCreateUpdateAfterBindBit),
	}, nil, &pool)
	if err := NewError(result); err != nil {
		return vk.NullDescriptorPool, vk.NullDescriptorSet, fmt.Errorf("create descriptor pool: %w", err)
	}

	var staticSet vk.DescriptorSet
	result = vk.AllocateDescriptorSets(handles.Device, &vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     pool,
		DescriptorSetCount: 1,
		PSetLayouts:        []vk.DescriptorSetLayout{staticLayout},
	}, &staticSet)
	if err := NewError(result); err != nil {
		return vk.NullDescriptorPool, vk.NullDescriptorSet, fmt.Errorf("allocate static descriptor set: %w", err)
	}

	return pool, staticSet, nil
}
