package vulkan

import (
	as "github.com/LamkasDev/asche"
	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

type GraphicsPipelineRequest struct {
	VertexModule   vk.ShaderModule
	GeometryModule vk.ShaderModule
	FragmentModule vk.ShaderModule
	RenderPass     vk.RenderPass

	GraphicsPipelineKey
}

type GraphicsPipelineKey struct {
	VertexModuleAddress   uintptr
	FragmentModuleAddress uintptr
	RenderTargetAddress   uintptr

	// Pipeline options.
	Width    uint32
	Height   uint32
	PrimType uint32

	// Render control flags.
	BlendAttachment   vk.PipelineColorBlendAttachmentState
	DepthStencilState vk.PipelineDepthStencilStateCreateInfo
	LogicOpEnable     vk.Bool32
	LogicOp           vk.LogicOp

	// Shader control flags.
	DbKillEnable           bool
	DbCoverageToMaskEnable bool
	DbAlphaToMaskDisable   bool

	// Viewport/window control.
	VpScissorEnable    bool
	WindowOffsetEnable bool

	// Line stipple.
	LineStippleEnable bool

	// Anti-aliasing flags.
	MsaaEnable          bool
	MsaaSampleLocations uint32

	// Primitive restart options.
	MultiPrimIbResetEnable bool
}

type ComputePipelineRequest struct {
	ComputeModule vk.ShaderModule

	ComputePipelineKey
}

type ComputePipelineKey struct {
	ComputeModuleAddress uintptr
}

func (t *GpuTranslator) createGraphicsPipeline(request GraphicsPipelineRequest) (vk.Pipeline, error) {
	// Setup stages.
	/* subgroupSizeVs := &VkPipelineShaderStageRequiredSubgroupSizeCreateInfoEXT{
		SType:                StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt,
		RequiredSubgroupSize: t.handles.SubgroupSizeProperties.MaxSubgroupSize,
	}
	subgroupSizeFs := &VkPipelineShaderStageRequiredSubgroupSizeCreateInfoEXT{
		SType:                StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt,
		RequiredSubgroupSize: t.handles.SubgroupSizeProperties.MaxSubgroupSize,
	} */
	stages := []vk.PipelineShaderStageCreateInfo{
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageVertexBit,
			Module: request.VertexModule,
			PName:  "main\x00",
			PNext:/* unsafe.Pointer(subgroupSizeVs) */ nil,
		},
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageFragmentBit,
			Module: request.FragmentModule,
			PName:  "main\x00",
			PNext:/* unsafe.Pointer(subgroupSizeFs) */ nil,
		},
	}
	if request.GeometryModule != vk.NullShaderModule {
		stages = append(stages, vk.PipelineShaderStageCreateInfo{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageGeometryBit,
			Module: request.GeometryModule,
			PName:  "main\x00",
			PNext:  nil,
		})
	}

	topology := translateTopology(request.PrimType)

	// No vertex input.
	vertexInput := vk.PipelineVertexInputStateCreateInfo{
		SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
	}
	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:    vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology: topology,
		PrimitiveRestartEnable: vk.Bool32(
			nstd.Btoi(request.MultiPrimIbResetEnable &&
				(topology == vk.PrimitiveTopologyLineStrip ||
					topology == vk.PrimitiveTopologyTriangleStrip ||
					topology == vk.PrimitiveTopologyTriangleFan)),
		),
	}

	// Viewport and scissor are dynamic so they match each draw call without rebuilding the pipeline.
	hasScissor := request.VpScissorEnable || request.WindowOffsetEnable
	dynStates := []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateBlendConstants}
	if hasScissor {
		dynStates = append(dynStates, vk.DynamicStateScissor)
	}
	if request.LineStippleEnable {
		// dynStates = append(dynStates, vk.DynamicStateLineStippleExt)
	}
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
	if !hasScissor {
		viewportState.PScissors = []vk.Rect2D{{
			Offset: vk.Offset2D{X: 0, Y: 0},
			Extent: vk.Extent2D{Width: 16384, Height: 16384},
		}}
	}

	// Setup rasterization.
	raster := vk.PipelineRasterizationStateCreateInfo{
		SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
		PolygonMode: vk.PolygonModeFill,
		CullMode:    vk.CullModeFlags(vk.CullModeNone),
		FrontFace:   vk.FrontFaceCounterClockwise,
		LineWidth:   1.0,
	}
	multisample := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  translateMsaaSamples(request.MsaaSampleLocations),
		SampleShadingEnable:   vk.False,
		MinSampleShading:      1.0,
		PSampleMask:           nil,
		AlphaToCoverageEnable: vk.Bool32(nstd.Btoi((request.DbKillEnable || request.DbCoverageToMaskEnable) && !request.DbAlphaToMaskDisable)),
		AlphaToOneEnable:      vk.False,
	}
	depthStencil := request.DepthStencilState

	// Setup blending.
	blend := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   request.LogicOpEnable,
		LogicOp:         request.LogicOp,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{request.BlendAttachment},
	}

	// Create pipeline.
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
			Layout:              t.pipelineLayout,
			RenderPass:          request.RenderPass,
		}},
		nil, pipelines)
	if err := as.NewError(result); err != nil {
		return vk.NullPipeline, err
	}

	return pipelines[0], nil
}

func (t *GpuTranslator) createComputePipeline(request ComputePipelineRequest) (vk.Pipeline, error) {
	// Setup stages.
	/* subgroupSizeCs := &VkPipelineShaderStageRequiredSubgroupSizeCreateInfoEXT{
		SType:                StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt,
		RequiredSubgroupSize: t.handles.SubgroupSizeProperties.MaxSubgroupSize,
	} */
	stage := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageComputeBit,
		Module: request.ComputeModule,
		PName:  "main\x00",
		PNext:/* unsafe.Pointer(subgroupSizeCs) */ nil,
	}

	// Create pipeline.
	pipelines := make([]vk.Pipeline, 1)
	result := vk.CreateComputePipelines(t.handles.Device, vk.NullPipelineCache, 1,
		[]vk.ComputePipelineCreateInfo{{
			SType:  vk.StructureTypeComputePipelineCreateInfo,
			Stage:  stage,
			Layout: t.pipelineLayout,
		}},
		nil, pipelines)
	if err := as.NewError(result); err != nil {
		return vk.NullPipeline, err
	}

	return pipelines[0], nil
}
