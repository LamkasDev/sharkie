package vulkan

import (
	"unsafe"

	as "github.com/LamkasDev/asche"
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
		},
		BindingCount: 1,
	}
	result := vk.CreateDescriptorSetLayout(t.handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PNext: unsafe.Pointer(&stubBindingFlags),
		PBindings: []vk.DescriptorSetLayoutBinding{{
			Binding:            0,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount:    BindlessTextureCapacity,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics),
			PImmutableSamplers: nil,
		}},
		BindingCount: 1,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreateUpdateAfterBindPoolBit),
	}, nil, &stubLayout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.stubDescriptorSetLayout = stubLayout

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

	// Create descriptor set layout for discovery (Set 2).
	var discoveryLayout vk.DescriptorSetLayout
	discoveryBindings := []vk.DescriptorSetLayoutBinding{
		{
			Binding:            0,
			DescriptorType:     vk.DescriptorTypeStorageBuffer,
			DescriptorCount:    1,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics | vk.ShaderStageComputeBit),
			PImmutableSamplers: nil,
		},
		{
			Binding:            1,
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
		PPushConstantRanges: []vk.PushConstantRange{{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       uint32(unsafe.Sizeof(StubPushConstants{})),
		}},
		PushConstantRangeCount: 1,
		PSetLayouts: []vk.DescriptorSetLayout{
			t.stubDescriptorSetLayout,
			t.texelDescriptorSetLayout,
			t.discoveryDescriptorSetLayout,
		},
		SetLayoutCount: 3,
	}, nil, &layout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.stubPipelineLayout = layout

	return nil
}
