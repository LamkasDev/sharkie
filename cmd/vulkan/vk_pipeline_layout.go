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

	// Create descriptor set layout for texel buffers.
	var texelLayout vk.DescriptorSetLayout
	texelBindings := make([]vk.DescriptorSetLayoutBinding, 4)
	texelBindingFlagsArray := make([]vk.DescriptorBindingFlags, 4)
	for i := range 4 {
		texelBindings[i] = vk.DescriptorSetLayoutBinding{
			Binding:            uint32(i),
			DescriptorType:     vk.DescriptorTypeUniformTexelBuffer,
			DescriptorCount:    1,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
			PImmutableSamplers: nil,
		}
		texelBindingFlagsArray[i] = vk.DescriptorBindingFlags(vk.DescriptorBindingUpdateAfterBindBit | vk.DescriptorBindingPartiallyBoundBit)
	}
	texelBindingFlags := vk.DescriptorSetLayoutBindingFlagsCreateInfo{
		SType:         vk.StructureTypeDescriptorSetLayoutBindingFlagsCreateInfo,
		PBindingFlags: texelBindingFlagsArray,
		BindingCount:  4,
	}
	result = vk.CreateDescriptorSetLayout(t.handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PNext:        unsafe.Pointer(&texelBindingFlags),
		PBindings:    texelBindings,
		BindingCount: 4,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreateUpdateAfterBindPoolBit),
	}, nil, &texelLayout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.texelDescriptorSetLayout = texelLayout

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
			spirvStructs.DescriptorSetSlotBindless:       t.bindlessDescriptorSetLayout,
			spirvStructs.DescriptorSetSlotDiscovery:      t.discoveryDescriptorSetLayout,
			spirvStructs.DescriptorSetSlotTexel:          t.texelDescriptorSetLayout,
			spirvStructs.DescriptorSetSlotTexelSecondary: t.texelDescriptorSetLayout,
		},
		SetLayoutCount: 4,
	}, nil, &layout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.pipelineLayout = layout

	return nil
}
