package vulkan

import (
	"unsafe"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
)

func CreateStubPipelineLayout(handles *VulkanHandles) (vk.PipelineLayout, vk.DescriptorSetLayout, vk.DescriptorSetLayout, error) {
	var globalLayout vk.DescriptorSetLayout
	globalBindingFlags := vk.DescriptorSetLayoutBindingFlagsCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutBindingFlagsCreateInfo,
		PBindingFlags: []vk.DescriptorBindingFlags{
			vk.DescriptorBindingFlags(0), // AddressTranslation
		},
		BindingCount: 1,
	}
	result := vk.CreateDescriptorSetLayout(handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PNext: unsafe.Pointer(&globalBindingFlags),
		PBindings: []vk.DescriptorSetLayoutBinding{
			{
				Binding:            spirvStructs.GlobalBindingAddressTranslation,
				DescriptorType:     vk.DescriptorTypeStorageBuffer,
				DescriptorCount:    1,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
		},
		BindingCount: 1,
		Flags:        0,
	}, nil, &globalLayout)
	if err := NewError(result); err != nil {
		return vk.NullPipelineLayout, vk.NullDescriptorSetLayout, vk.NullDescriptorSetLayout, err
	}

	var imageLayout vk.DescriptorSetLayout
	imageBindingFlags := vk.DescriptorSetLayoutBindingFlagsCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutBindingFlagsCreateInfo,
		PBindingFlags: []vk.DescriptorBindingFlags{
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages1D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages1D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages2D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages2D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages3D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages3D
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // SampledImages2DArray
			vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit), // StorageImages2DArray
		},
		BindingCount: 8,
	}
	result = vk.CreateDescriptorSetLayout(handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PNext: unsafe.Pointer(&imageBindingFlags),
		PBindings: []vk.DescriptorSetLayoutBinding{
			{
				Binding:            spirvStructs.ImageBindingSampledImages1D,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.ImageBindingStorageImages1D,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.ImageBindingSampledImages2D,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.ImageBindingStorageImages2D,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.ImageBindingSampledImages3D,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.ImageBindingStorageImages3D,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.ImageBindingSampledImages2DArray,
				DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
			{
				Binding:            spirvStructs.ImageBindingStorageImages2DArray,
				DescriptorType:     vk.DescriptorTypeStorageImage,
				DescriptorCount:    spirvStructs.MaxStaticBindings,
				StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
				PImmutableSamplers: nil,
			},
		},
		BindingCount: 8,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreateUpdateAfterBindPoolBit),
	}, nil, &imageLayout)
	if err := NewError(result); err != nil {
		return vk.NullPipelineLayout, vk.NullDescriptorSetLayout, vk.NullDescriptorSetLayout, err
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
		PSetLayouts:            []vk.DescriptorSetLayout{globalLayout, imageLayout},
		SetLayoutCount:         2,
	}, nil, &layout)
	if err := NewError(result); err != nil {
		return vk.NullPipelineLayout, vk.NullDescriptorSetLayout, vk.NullDescriptorSetLayout, err
	}

	return layout, globalLayout, imageLayout, nil
}
