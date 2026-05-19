package vulkan

import (
	"unsafe"

	as "github.com/LamkasDev/asche"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
)

const BindlessTextureCapacity = 8192

func (t *GpuTranslator) createStubPipelineLayout() error {
	// Create descriptor set layout for bindless textures.
	var stubLayout vk.DescriptorSetLayout
	stubBindingFlags := vk.DescriptorSetLayoutBindingFlagsCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutBindingFlagsCreateInfo,
		PBindingFlags: []vk.DescriptorBindingFlags{
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit),
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit),
		},
		BindingCount: 2,
	}
	result := vk.CreateDescriptorSetLayout(t.handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PNext: unsafe.Pointer(&stubBindingFlags),
		PBindings: []vk.DescriptorSetLayoutBinding{
			{
				Binding:            spirvStructs.BindlessBindingSampledImages,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    BindlessTextureCapacity,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.BindlessBindingStorageImages,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    BindlessTextureCapacity,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
		},
		BindingCount: 2,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreateUpdateAfterBindPoolBit),
	}, nil, &stubLayout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.bindlessDescriptorSetLayout = stubLayout

	// Create descriptor set layout for discovery.
	var discoveryLayout vk.DescriptorSetLayout
	discoveryBindings := []vk.DescriptorSetLayoutBinding{
		{
			Binding:            spirvStructs.DiscoveryBindingMap,
			DescriptorType:     vk.DescriptorTypeStorageBuffer,
			DescriptorCount:    1,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
			PImmutableSamplers: nil,
		},
		{
			Binding:            spirvStructs.DiscoveryBindingMissingResource,
			DescriptorType:     vk.DescriptorTypeStorageBuffer,
			DescriptorCount:    1,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
			PImmutableSamplers: nil,
		},
	}
	result = vk.CreateDescriptorSetLayout(t.handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PBindings:    discoveryBindings,
		BindingCount: uint32(len(discoveryBindings)),
	}, nil, &discoveryLayout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.discoveryDescriptorSetLayout = discoveryLayout

	var layout vk.PipelineLayout
	result = vk.CreatePipelineLayout(t.handles.Device, &vk.PipelineLayoutCreateInfo{
		SType: vk.StructureTypePipelineLayoutCreateInfo,
		PPushConstantRanges: []vk.PushConstantRange{
			{
				StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageComputeBit),
				Offset:     0,
				Size:       spirvStructs.PushConstantsSize,
			},
			{
				StageFlags: vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
				Offset:     spirvStructs.PushConstantsSize,
				Size:       spirvStructs.PushConstantsSize,
			},
		},
		PushConstantRangeCount: 2,
		PSetLayouts: []vk.DescriptorSetLayout{
			spirvStructs.DescriptorSetSlotBindless:  t.bindlessDescriptorSetLayout,
			spirvStructs.DescriptorSetSlotDiscovery: t.discoveryDescriptorSetLayout,
		},
		SetLayoutCount: 2,
	}, nil, &layout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.pipelineLayout = layout

	return nil
}
