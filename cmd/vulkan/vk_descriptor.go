package vulkan

import (
	"fmt"

	as "github.com/LamkasDev/asche"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) createDescriptorPool() error {
	var pool vk.DescriptorPool
	result := vk.CreateDescriptorPool(t.handles.Device, &vk.DescriptorPoolCreateInfo{
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
		PoolSizeCount: 4,
		MaxSets:       8192,
		Flags:         vk.DescriptorPoolCreateFlags(vk.DescriptorPoolCreateFreeDescriptorSetBit | vk.DescriptorPoolCreateUpdateAfterBindBit),
	}, nil, &pool)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("create descriptor pool: %w", err)
	}
	t.descriptorPool = pool

	// Allocate discovery descriptor set.
	var discoverySet vk.DescriptorSet
	result = vk.AllocateDescriptorSets(t.handles.Device, &vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     t.descriptorPool,
		DescriptorSetCount: 1,
		PSetLayouts:        []vk.DescriptorSetLayout{t.discoveryDescriptorSetLayout},
	}, &discoverySet)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("allocate discovery descriptor set: %w", err)
	}
	t.discoveryDescriptorSet = discoverySet

	// Allocate bindless descriptor set.
	var bindlessSet vk.DescriptorSet
	result = vk.AllocateDescriptorSets(t.handles.Device, &vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     t.descriptorPool,
		DescriptorSetCount: 1,
		PSetLayouts:        []vk.DescriptorSetLayout{t.bindlessDescriptorSetLayout},
	}, &bindlessSet)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("allocate bindless descriptor set: %w", err)
	}
	t.bindlessDescriptorSet = bindlessSet

	return nil
}

func (t *GpuTranslator) updateDiscoveryDescriptorSet() {
	vk.UpdateDescriptorSets(t.handles.Device, 2, []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.discoveryDescriptorSet,
			DstBinding:      spirvStructs.DiscoveryBindingMap,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeStorageBuffer,
			PBufferInfo: []vk.DescriptorBufferInfo{{
				Buffer: t.discoveryMapBuffer,
				Offset: 0,
				Range:  vk.DeviceSize(vk.WholeSize),
			}},
		},
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.discoveryDescriptorSet,
			DstBinding:      spirvStructs.DiscoveryBindingMissingResource,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeStorageBuffer,
			PBufferInfo: []vk.DescriptorBufferInfo{{
				Buffer: t.missingResourceBuffer,
				Offset: 0,
				Range:  vk.DeviceSize(vk.WholeSize),
			}},
		},
	}, 0, nil)
}
