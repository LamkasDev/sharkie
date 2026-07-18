package vulkan

//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/macro_8bpp.comp -o ../../data/shaders/detile_macro_8bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/macro_32bpp.comp -o ../../data/shaders/detile_macro_32bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/macro_64bpp.comp -o ../../data/shaders/detile_macro_64bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_8bpp.comp -o ../../data/shaders/detile_micro_8bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_16bpp.comp -o ../../data/shaders/detile_micro_16bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_32bpp.comp -o ../../data/shaders/detile_micro_32bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_64bpp.comp -o ../../data/shaders/detile_micro_64bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_128bpp.comp -o ../../data/shaders/detile_micro_128bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/display_micro_64bpp.comp -o ../../data/shaders/detile_display_micro_64bpp.spv

import (
	"fmt"
	"io/ioutil"
	"unsafe"

	vk "github.com/goki/vulkan"
)

type DetilePipeline struct {
	Pipeline         vk.Pipeline
	PipelineLayout   vk.PipelineLayout
	DescriptorLayout vk.DescriptorSetLayout
	Module           vk.ShaderModule
	DescriptorPool   vk.DescriptorPool
	DescriptorSet    vk.DescriptorSet
}

var detilePipelines map[string]*DetilePipeline

func init() {
	detilePipelines = make(map[string]*DetilePipeline)
}

func createDetileShaderModule(device vk.Device, path string) (vk.ShaderModule, error) {
	code, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var module vk.ShaderModule
	var codeUint32 []uint32
	if len(code) > 0 {
		codeUint32 = unsafe.Slice((*uint32)(unsafe.Pointer(&code[0])), len(code)/4)
	}
	result := vk.CreateShaderModule(device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(code)),
		PCode:    codeUint32,
	}, nil, &module)
	if err := NewError(result); err != nil {
		return nil, err
	}
	return module, nil
}

func GetDetilePipeline(handles *VulkanHandles, bpp int, isMicro bool, isDisplayMicro bool) (*DetilePipeline, error) {
	modeStr := "macro"
	if isDisplayMicro {
		modeStr = "display_micro"
	} else if isMicro {
		modeStr = "micro"
	}
	key := fmt.Sprintf("%s_%dbpp", modeStr, bpp*8)
	if pipeline, ok := detilePipelines[key]; ok {
		return pipeline, nil
	}

	spvPath := fmt.Sprintf("data/shaders/detile_%s.spv", key)
	module, err := createDetileShaderModule(handles.Device, spvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load shader %s: %v", spvPath, err)
	}

	bindings := []vk.DescriptorSetLayoutBinding{
		{
			Binding:         0,
			DescriptorType:  vk.DescriptorTypeStorageBuffer,
			DescriptorCount: 1,
			StageFlags:      vk.ShaderStageFlags(vk.ShaderStageComputeBit),
		},
		{
			Binding:         1,
			DescriptorType:  vk.DescriptorTypeStorageBuffer,
			DescriptorCount: 1,
			StageFlags:      vk.ShaderStageFlags(vk.ShaderStageComputeBit),
		},
	}

	var descLayout vk.DescriptorSetLayout
	result := vk.CreateDescriptorSetLayout(handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: uint32(len(bindings)),
		PBindings:    bindings,
	}, nil, &descLayout)
	if err := NewError(result); err != nil {
		return nil, err
	}

	var pool vk.DescriptorPool
	result = vk.CreateDescriptorPool(handles.Device, &vk.DescriptorPoolCreateInfo{
		SType: vk.StructureTypeDescriptorPoolCreateInfo,
		PPoolSizes: []vk.DescriptorPoolSize{
			{
				Type:            vk.DescriptorTypeStorageBuffer,
				DescriptorCount: 2,
			},
		},
		PoolSizeCount: 1,
		MaxSets:       1,
	}, nil, &pool)
	if err := NewError(result); err != nil {
		return nil, err
	}

	var descSet vk.DescriptorSet
	descSets := make([]vk.DescriptorSet, 1)
	result = vk.AllocateDescriptorSets(handles.Device, &vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     pool,
		DescriptorSetCount: 1,
		PSetLayouts:        []vk.DescriptorSetLayout{descLayout},
	}, &descSets[0])
	if err := NewError(result); err != nil {
		return nil, err
	}
	descSet = descSets[0]

	pushConstantRanges := []vk.PushConstantRange{
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageComputeBit),
			Offset:     0,
			Size:       128, // Max size needed by any detile shader (e.g. micro_32bpp needs 76 bytes)
		},
	}

	var pipelineLayout vk.PipelineLayout
	result = vk.CreatePipelineLayout(handles.Device, &vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descLayout},
		PushConstantRangeCount: 1,
		PPushConstantRanges:    pushConstantRanges,
	}, nil, &pipelineLayout)
	if err := NewError(result); err != nil {
		return nil, err
	}

	pipelineName := []byte("main\x00")
	pipelines := make([]vk.Pipeline, 1)
	result = vk.CreateComputePipelines(handles.Device, nil, 1, []vk.ComputePipelineCreateInfo{
		{
			SType: vk.StructureTypeComputePipelineCreateInfo,
			Stage: vk.PipelineShaderStageCreateInfo{
				SType:  vk.StructureTypePipelineShaderStageCreateInfo,
				Stage:  vk.ShaderStageComputeBit,
				Module: module,
				PName:  string(pipelineName),
			},
			Layout: pipelineLayout,
		},
	}, nil, pipelines)
	if err := NewError(result); err != nil {
		return nil, err
	}

	var pipeline = pipelines[0]
	detilePipelines[key] = &DetilePipeline{
		Pipeline:         pipeline,
		PipelineLayout:   pipelineLayout,
		DescriptorLayout: descLayout,
		Module:           module,
		DescriptorPool:   pool,
		DescriptorSet:    descSet,
	}

	return detilePipelines[key], nil
}
