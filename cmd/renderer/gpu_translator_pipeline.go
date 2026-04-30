package renderer

import (
	"fmt"
	"unsafe"

	as "github.com/LamkasDev/asche"
	vk "github.com/goki/vulkan"
)

type GpuTranslatorPipelineKey struct {
	VertexShaderAddress uintptr
	PixelShaderAddress  uintptr
	SurfaceAddress      uintptr
}

func (t *GpuTranslator) GetPipeline(key GpuTranslatorPipelineKey, vsModule, psModule vk.ShaderModule, renderPass vk.RenderPass, width, height uint32) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.pipelinesMutex.Lock()
	pipeline, ok := t.pipelines[key]
	t.pipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := t.createPipelineFromModules(vsModule, psModule, renderPass, width, height)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createCompiledPipeline 0x%X: %w", key.PixelShaderAddress, err)
	}
	t.pipelinesMutex.Lock()
	t.pipelines[key] = pipeline
	t.pipelinesMutex.Unlock()

	return pipeline, nil
}

func (t *GpuTranslator) createStubPipelineLayout() error {
	// Create descriptor set layout for bindless textures.
	var stubLayout vk.DescriptorSetLayout
	result := vk.CreateDescriptorSetLayout(t.handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType: vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PBindings: []vk.DescriptorSetLayoutBinding{{
			Binding:            0,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount:    1024,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageAllGraphics),
			PImmutableSamplers: nil,
		}},
		BindingCount: 1,
	}, nil, &stubLayout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.stubDescriptorSetLayout = stubLayout

	// Create descriptor set layout for texel buffers.
	var texelLayout vk.DescriptorSetLayout
	texelBindings := make([]vk.DescriptorSetLayoutBinding, 4)
	for i := range 4 {
		texelBindings[i] = vk.DescriptorSetLayoutBinding{
			Binding:            uint32(i),
			DescriptorType:     vk.DescriptorTypeUniformTexelBuffer,
			DescriptorCount:    1,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
			PImmutableSamplers: nil,
		}
	}
	result = vk.CreateDescriptorSetLayout(t.handles.Device, &vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		PBindings:    texelBindings,
		BindingCount: 4,
	}, nil, &texelLayout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.texelDescriptorSetLayout = texelLayout

	var layout vk.PipelineLayout
	result = vk.CreatePipelineLayout(t.handles.Device, &vk.PipelineLayoutCreateInfo{
		SType: vk.StructureTypePipelineLayoutCreateInfo,
		PPushConstantRanges: []vk.PushConstantRange{{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       uint32(unsafe.Sizeof(StubPushConstants{})),
		}},
		PushConstantRangeCount: 1,
		PSetLayouts:            []vk.DescriptorSetLayout{t.stubDescriptorSetLayout, t.texelDescriptorSetLayout},
		SetLayoutCount:         2,
	}, nil, &layout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.stubPipelineLayout = layout

	return nil
}

func (t *GpuTranslator) createPipelineFromModules(vsModule, fsModule vk.ShaderModule, renderPass vk.RenderPass, width, height uint32) (vk.Pipeline, error) {
	stages := []vk.PipelineShaderStageCreateInfo{
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageVertexBit,
			Module: vsModule,
			PName:  "main\x00",
		},
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageFragmentBit,
			Module: fsModule,
			PName:  "main\x00",
		},
	}

	// No vertex input.
	vertexInput := vk.PipelineVertexInputStateCreateInfo{
		SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
	}
	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:    vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology: vk.PrimitiveTopologyTriangleList,
	}

	// Viewport and scissor are dynamic so they match each DrawCall without rebuilding the pipeline.
	dynStates := []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateScissor}
	dynamicState := vk.PipelineDynamicStateCreateInfo{
		SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynStates)),
		PDynamicStates:    dynStates,
	}

	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		ScissorCount:  1,
	}

	depthClipExt := vk.PipelineRasterizationDepthClipStateCreateInfo{
		SType:           vk.StructureTypePipelineRasterizationDepthClipStateCreateInfo,
		DepthClipEnable: vk.False,
	}

	raster := vk.PipelineRasterizationStateCreateInfo{
		SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
		PNext:       unsafe.Pointer(&depthClipExt),
		PolygonMode: vk.PolygonModeFill,
		CullMode:    vk.CullModeFlags(vk.CullModeNone),
		FrontFace:   vk.FrontFaceCounterClockwise,
		LineWidth:   1.0,
	}

	multisample := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  vk.SampleCount1Bit,
		SampleShadingEnable:   vk.False,
		MinSampleShading:      1.0,
		PSampleMask:           nil,
		AlphaToCoverageEnable: vk.False,
		AlphaToOneEnable:      vk.False,
	}

	depthStencil := vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       vk.False,
		DepthWriteEnable:      vk.False,
		DepthCompareOp:        vk.CompareOpLess,
		DepthBoundsTestEnable: vk.False,
		StencilTestEnable:     vk.False,
	}

	// Opaque blend.
	blendAttach := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(
			vk.ColorComponentRBit | vk.ColorComponentGBit |
				vk.ColorComponentBBit | vk.ColorComponentABit),
		BlendEnable:         vk.False,
		SrcColorBlendFactor: vk.BlendFactorOne,
		DstColorBlendFactor: vk.BlendFactorZero,
		ColorBlendOp:        vk.BlendOpAdd,
		SrcAlphaBlendFactor: vk.BlendFactorOne,
		DstAlphaBlendFactor: vk.BlendFactorZero,
		AlphaBlendOp:        vk.BlendOpAdd,
	}
	blend := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		LogicOp:         vk.LogicOpCopy,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{blendAttach},
	}

	pipelines := make([]vk.Pipeline, 1)
	result := vk.CreateGraphicsPipelines(t.handles.Device, vk.NullPipelineCache, 1,
		[]vk.GraphicsPipelineCreateInfo{{
			SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
			StageCount:          uint32(len(stages)),
			PStages:             stages,
			PVertexInputState:   &vertexInput,
			PInputAssemblyState: &inputAssembly,
			PViewportState:      &viewportState,
			PRasterizationState: &raster,
			PMultisampleState:   &multisample,
			PDepthStencilState:  &depthStencil,
			PColorBlendState:    &blend,
			PDynamicState:       &dynamicState,
			Layout:              t.stubPipelineLayout,
			RenderPass:          renderPass,
		}},
		nil, pipelines)
	if err := as.NewError(result); err != nil {
		return vk.NullPipeline, err
	}

	return pipelines[0], nil
}
