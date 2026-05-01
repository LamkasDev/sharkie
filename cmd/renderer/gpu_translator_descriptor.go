package renderer

import (
	"fmt"
	"unsafe"

	as "github.com/LamkasDev/asche"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) createDescriptorPool() error {
	var pool vk.DescriptorPool
	result := vk.CreateDescriptorPool(t.handles.Device, &vk.DescriptorPoolCreateInfo{
		SType: vk.StructureTypeDescriptorPoolCreateInfo,
		PPoolSizes: []vk.DescriptorPoolSize{
			{
				Type:            vk.DescriptorTypeUniformTexelBuffer,
				DescriptorCount: 2048,
			},
			{
				Type:            vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount: 2048,
			},
			{
				Type:            vk.DescriptorTypeStorageBuffer,
				DescriptorCount: 16,
			},
		},
		PoolSizeCount: 3,
		MaxSets:       2048,
		Flags:         vk.DescriptorPoolCreateFlags(vk.DescriptorPoolCreateFreeDescriptorSetBit),
	}, nil, &pool)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("create descriptor pool: %w", err)
	}
	t.descriptorPool = pool

	// Allocate texel descriptor sets.
	t.texelDescriptorSets = make([]vk.DescriptorSet, 1024)
	texelLayouts := make([]vk.DescriptorSetLayout, 1024)
	for i := range 1024 {
		texelLayouts[i] = t.texelDescriptorSetLayout
	}
	result = vk.AllocateDescriptorSets(t.handles.Device, &vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     t.descriptorPool,
		DescriptorSetCount: 1024,
		PSetLayouts:        texelLayouts,
	}, unsafe.SliceData(t.texelDescriptorSets))
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("allocate texel descriptor sets: %w", err)
	}

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
		PSetLayouts:        []vk.DescriptorSetLayout{t.stubDescriptorSetLayout},
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
			DstBinding:      0,
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
			DstBinding:      1,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeStorageBuffer,
			PBufferInfo: []vk.DescriptorBufferInfo{{
				Buffer: t.discoveryReportBuffer,
				Offset: 0,
				Range:  vk.DeviceSize(vk.WholeSize),
			}},
		},
	}, 0, nil)
}
