package vulkan

//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/macro_8bpp.comp -o ../../data/shaders/detile_macro_8bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/macro_16bpp.comp -o ../../data/shaders/detile_macro_16bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/macro_32bpp.comp -o ../../data/shaders/detile_macro_32bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/macro_64bpp.comp -o ../../data/shaders/detile_macro_64bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_8bpp.comp -o ../../data/shaders/detile_micro_8bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_16bpp.comp -o ../../data/shaders/detile_micro_16bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_32bpp.comp -o ../../data/shaders/detile_micro_32bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_64bpp.comp -o ../../data/shaders/detile_micro_64bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/micro_128bpp.comp -o ../../data/shaders/detile_micro_128bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/display_micro_32bpp.comp -o ../../data/shaders/detile_display_micro_32bpp.spv
//go:generate glslc --target-env=vulkan1.2 ../../data/shaders/display_micro_64bpp.comp -o ../../data/shaders/detile_display_micro_64bpp.spv

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie"
	vk "github.com/goki/vulkan"
)

type DetilePipeline struct {
	Pipeline       vk.Pipeline
	PipelineLayout vk.PipelineLayout
	Module         vk.ShaderModule
	DescriptorPool *VulkanDescriptorPool2
}

var detilePipelines map[string]*DetilePipeline

func init() {
	detilePipelines = make(map[string]*DetilePipeline)
}

func createDetileShaderModule(device vk.Device, path string) (vk.ShaderModule, error) {
	code, err := sharkie.Assets.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Copy code into aligned slice.
	codeSize := len(code)
	uint32Count := (codeSize + 3) / 4
	codeUint32 := make([]uint32, uint32Count)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&codeUint32[0])), codeSize), code)

	var module vk.ShaderModule
	result := vk.CreateShaderModule(device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(codeSize),
		PCode:    codeUint32,
	}, nil, &module)
	if err = NewError(result); err != nil {
		return nil, err
	}

	return module, nil
}

func ResetDetilePipelines(frame uint64) {
	for _, pipeline := range detilePipelines {
		pipeline.DescriptorPool.Reset(frame)
	}
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

	pool2, err := CreateDescriptorPool2(handles, descLayout, []vk.DescriptorPoolSize{
		{
			Type:            vk.DescriptorTypeStorageBuffer,
			DescriptorCount: 1024,
		},
	}, 128)
	if err != nil {
		return nil, err
	}

	var pipelineLayout vk.PipelineLayout
	result = vk.CreatePipelineLayout(handles.Device, &vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descLayout},
		PushConstantRangeCount: 1,
		PPushConstantRanges: []vk.PushConstantRange{
			{
				StageFlags: vk.ShaderStageFlags(vk.ShaderStageComputeBit),
				Offset:     0,
				Size:       128,
			},
		},
	}, nil, &pipelineLayout)
	if err = NewError(result); err != nil {
		return nil, err
	}

	pipelines := make([]vk.Pipeline, 1)
	result = vk.CreateComputePipelines(handles.Device, nil, 1, []vk.ComputePipelineCreateInfo{
		{
			SType: vk.StructureTypeComputePipelineCreateInfo,
			Stage: vk.PipelineShaderStageCreateInfo{
				SType:  vk.StructureTypePipelineShaderStageCreateInfo,
				Stage:  vk.ShaderStageComputeBit,
				Module: module,
				PName:  "main\x00",
			},
			Layout: pipelineLayout,
		},
	}, nil, pipelines)
	if err = NewError(result); err != nil {
		return nil, err
	}

	detilePipelines[key] = &DetilePipeline{
		Pipeline:       pipelines[0],
		PipelineLayout: pipelineLayout,
		Module:         module,
		DescriptorPool: pool2,
	}

	return detilePipelines[key], nil
}
