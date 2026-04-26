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
		},
		PoolSizeCount: 1,
		MaxSets:       1024,
		Flags:         vk.DescriptorPoolCreateFlags(vk.DescriptorPoolCreateFreeDescriptorSetBit),
	}, nil, &pool)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("GpuTranslator: descriptor pool: %w", err)
	}
	t.descriptorPool = pool

	// Allocate descriptor sets.
	t.texelDescriptorSets = make([]vk.DescriptorSet, 1024)
	layouts := make([]vk.DescriptorSetLayout, 1024)
	for i := range 1024 {
		layouts[i] = t.texelDescriptorSetLayout
	}
	result = vk.AllocateDescriptorSets(t.handles.Device, &vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     t.descriptorPool,
		DescriptorSetCount: 1024,
		PSetLayouts:        layouts,
	}, unsafe.SliceData(t.texelDescriptorSets))
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("GpuTranslator: allocate descriptor sets: %w", err)
	}

	return nil
}

func (t *GpuTranslator) imageBarrier(commandBuffer vk.CommandBuffer, image vk.Image,
	oldLayout, newLayout vk.ImageLayout,
	srcAccess, dstAccess vk.AccessFlags,
	srcStage, dstStage vk.PipelineStageFlags,
) {
	vk.CmdPipelineBarrier(commandBuffer,
		srcStage, dstStage,
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           oldLayout,
			NewLayout:           newLayout,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               image,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1,
				LayerCount: 1,
			},
			SrcAccessMask: srcAccess,
			DstAccessMask: dstAccess,
		}})
}
