package vulkan

import (
	"unsafe"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
)

func CreateStubPipelineLayout(handles *VulkanHandles) (vk.PipelineLayout, vk.DescriptorSetLayout, error) {
	// Create descriptor set layout for statically-bound resources (Set 2).
	var staticLayout vk.DescriptorSetLayout
	staticBindingFlags := vk.DescriptorSetLayoutBindingFlagsCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutBindingFlagsCreateInfo,
		PBindingFlags: []vk.DescriptorBindingFlags{
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit),
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit),
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit),
		},
		BindingCount: 3,
	}
	result := vk.CreateDescriptorSetLayout(handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PNext: unsafe.Pointer(&staticBindingFlags),
		PBindings: []vk.DescriptorSetLayoutBinding{
			{
				Binding:            spirvStructs.StaticBindingSampledImages,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingStorageImages,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingSampledBuffers,
				DescriptorType:     vk.DescriptorTypeUniformTexelBuffer,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
		},
		BindingCount: 3,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreateUpdateAfterBindPoolBit),
	}, nil, &staticLayout)
	if err := NewError(result); err != nil {
		return vk.NullPipelineLayout, vk.NullDescriptorSetLayout, err
	}

	var layout vk.PipelineLayout
	result = vk.CreatePipelineLayout(handles.Device, &vk.PipelineLayoutCreateInfo{
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
		PSetLayouts:            []vk.DescriptorSetLayout{staticLayout},
		SetLayoutCount:         1,
	}, nil, &layout)
	if err := NewError(result); err != nil {
		return vk.NullPipelineLayout, vk.NullDescriptorSetLayout, err
	}

	return layout, staticLayout, nil
}
