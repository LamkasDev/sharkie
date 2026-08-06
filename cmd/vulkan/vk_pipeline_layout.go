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
			vk.DescriptorBindingFlags(0), // AddressTranslation
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledBuffers
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages1D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages2D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages2D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages3D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages2DArray
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages1D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages3D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages2DArray
		},
		BindingCount: 10,
	}
	result := vk.CreateDescriptorSetLayout(handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PNext: unsafe.Pointer(&staticBindingFlags),
		PBindings: []vk.DescriptorSetLayoutBinding{
			{
				Binding:            spirvStructs.StaticBindingAddressTranslation,
				DescriptorType:     vk.DescriptorTypeStorageBuffer,
				DescriptorCount:    1,
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
			{
				Binding:            spirvStructs.StaticBindingSampledImages1D,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingStorageImages1D,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingSampledImages2D,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingStorageImages2D,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingSampledImages3D,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingStorageImages3D,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingSampledImages2DArray,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.StaticBindingStorageImages2DArray,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
		},
		BindingCount: 10,
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
