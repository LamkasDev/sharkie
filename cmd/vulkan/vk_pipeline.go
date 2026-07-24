package vulkan

import (
	"unsafe"

	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

type GraphicsPipelineRequest struct {
	VertexModule      vk.ShaderModule
	TessControlModule vk.ShaderModule
	TessEvalModule    vk.ShaderModule
	GeometryModule    vk.ShaderModule
	FragmentModule    vk.ShaderModule
	RenderPass        vk.RenderPass

	GraphicsPipelineKey
}

type GraphicsPipelineKey struct {
	VertexModuleAddress   uintptr
	FetchShaderAddress    uintptr
	FragmentModuleAddress uintptr
	RenderTargetAddress   uintptr
	DepthTargetAddress    uintptr

	// Pipeline options.
	Width    uint32
	Height   uint32
	PrimType uint32

	// Culling and polygon mode.
	CullFront             bool
	CullBack              bool
	Face                  bool
	PolyMode              uint32
	PolyModeFrontPtype    uint32
	PolyModeBackPtype     uint32
	PolyOffsetFrontEnable bool
	PolyOffsetBackEnable  bool
	PolyOffsetParaEnable  bool
	ProvokingVertexLast   bool

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

func CreateGraphicsPipeline(handles *VulkanHandles, request GraphicsPipelineRequest, renderPass vk.RenderPass, layout vk.PipelineLayout, cache vk.PipelineCache) (vk.Pipeline, error) {
	// Setup stages.
	subgroupSizeVs := &VkPipelineShaderStageRequiredSubgroupSizeCreateInfoEXT{
		SType:                StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt,
		RequiredSubgroupSize: handles.SubgroupSizeProperties.MaxSubgroupSize,
	}
	subgroupSizeFs := &VkPipelineShaderStageRequiredSubgroupSizeCreateInfoEXT{
		SType:                StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt,
		RequiredSubgroupSize: handles.SubgroupSizeProperties.MaxSubgroupSize,
	}
	stages := []vk.PipelineShaderStageCreateInfo{
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageVertexBit,
			Module: request.VertexModule,
			PName:  "main\x00",
			PNext:  unsafe.Pointer(subgroupSizeVs),
		},
	}
	if request.TessControlModule != vk.NullShaderModule {
		stages = append(stages, vk.PipelineShaderStageCreateInfo{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageTessellationControlBit,
			Module: request.TessControlModule,
			PName:  "main\x00",
		})
	}
	if request.TessEvalModule != vk.NullShaderModule {
		stages = append(stages, vk.PipelineShaderStageCreateInfo{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageTessellationEvaluationBit,
			Module: request.TessEvalModule,
			PName:  "main\x00",
		})
	}
	if request.GeometryModule != vk.NullShaderModule {
		stages = append(stages, vk.PipelineShaderStageCreateInfo{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageGeometryBit,
			Module: request.GeometryModule,
			PName:  "main\x00",
		})
	}
	stages = append(stages, vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageFragmentBit,
		Module: request.FragmentModule,
		PName:  "main\x00",
		PNext:  unsafe.Pointer(subgroupSizeFs),
	})

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
					topology == vk.PrimitiveTopologyTriangleFan ||
					topology == vk.PrimitiveTopologyPatchList)),
		),
	}

	var tessellationState *vk.PipelineTessellationStateCreateInfo
	if request.PrimType == 17 { // RECTLIST
		patchPoints := uint32(3)
		tessellationState = &vk.PipelineTessellationStateCreateInfo{
			SType:              vk.StructureTypePipelineTessellationStateCreateInfo,
			PatchControlPoints: patchPoints,
		}
	}

	// Viewport and scissor are dynamic so they match each draw call without rebuilding the pipeline.
	dynStates := []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateScissor, vk.DynamicStateBlendConstants}
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
		PScissors: []vk.Rect2D{{
			Offset: vk.Offset2D{X: 0, Y: 0},
			Extent: vk.Extent2D{Width: 16384, Height: 16384},
		}},
	}

	// Setup rasterizer.
	frontFace := vk.FrontFaceCounterClockwise
	if request.Face {
		frontFace = vk.FrontFaceClockwise
	}
	cullMode := vk.CullModeNone
	if request.CullFront {
		cullMode |= vk.CullModeFrontBit
	}
	if request.CullBack {
		cullMode |= vk.CullModeBackBit
	}
	polygonMode := vk.PolygonModeFill
	switch request.PolyMode {
	case 1:
		polygonMode = vk.PolygonModeLine
	case 2:
		polygonMode = vk.PolygonModePoint
	}

	provokingVertex := vk.PipelineRasterizationProvokingVertexStateCreateInfo{
		SType:               vk.StructureTypePipelineRasterizationProvokingVertexStateCreateInfo,
		ProvokingVertexMode: vk.ProvokingVertexModeFirstVertex,
	}
	if request.ProvokingVertexLast {
		provokingVertex.ProvokingVertexMode = vk.ProvokingVertexModeLastVertex
	}

	raster := vk.PipelineRasterizationStateCreateInfo{
		SType:            vk.StructureTypePipelineRasterizationStateCreateInfo,
		PNext:            unsafe.Pointer(&provokingVertex),
		DepthClampEnable: vk.False,
		PolygonMode:      polygonMode,
		CullMode:         vk.CullModeFlags(cullMode),
		FrontFace:        frontFace,
		LineWidth:        1.0,
	}
	depthStencil := request.DepthStencilState
	if request.DepthTargetAddress == 0 {
		depthStencil.DepthTestEnable = vk.False
		depthStencil.DepthWriteEnable = vk.False
		depthStencil.StencilTestEnable = vk.False
	}

	// Setup anti-aliasing.
	multisample := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  translateMsaaSamples(request.MsaaSampleLocations),
		SampleShadingEnable:   vk.False,
		MinSampleShading:      1.0,
		PSampleMask:           nil,
		AlphaToCoverageEnable: vk.Bool32(nstd.Btoi((request.DbKillEnable || request.DbCoverageToMaskEnable) && !request.DbAlphaToMaskDisable)),
		AlphaToOneEnable:      vk.False,
	}

	blendAttachments := make([]vk.PipelineColorBlendAttachmentState, 8)
	blendAttachments[0] = request.BlendAttachment
	for i := 1; i < 8; i++ {
		blendAttachments[i] = request.BlendAttachment
		blendAttachments[i].ColorWriteMask = 0
	}

	// Setup blending.
	blend := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   request.LogicOpEnable,
		LogicOp:         request.LogicOp,
		AttachmentCount: 8,
		PAttachments:    blendAttachments,
	}

	// Create pipeline.
	pipelineInfo := vk.GraphicsPipelineCreateInfo{
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
		Layout:              layout,
		RenderPass:          renderPass,
	}
	if tessellationState != nil {
		pipelineInfo.PTessellationState = tessellationState
	}

	pipelines := make([]vk.Pipeline, 1)
	result := vk.CreateGraphicsPipelines(handles.Device, cache, 1,
		[]vk.GraphicsPipelineCreateInfo{pipelineInfo},
		nil, pipelines)
	if err := NewError(result); err != nil {
		return vk.NullPipeline, err
	}

	return pipelines[0], nil
}

func CreateComputePipeline(handles *VulkanHandles, request ComputePipelineRequest, layout vk.PipelineLayout, cache vk.PipelineCache) (vk.Pipeline, error) {
	// Setup stages.
	/* subgroupSizeCs := &VkPipelineShaderStageRequiredSubgroupSizeCreateInfoEXT{
		SType:                StructureTypePipelineShaderStageRequiredSubgroupSizeCreateInfoExt,
		RequiredSubgroupSize: handles.SubgroupSizeProperties.MaxSubgroupSize,
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
	result := vk.CreateComputePipelines(handles.Device, cache, 1,
		[]vk.ComputePipelineCreateInfo{{
			SType:  vk.StructureTypeComputePipelineCreateInfo,
			Stage:  stage,
			Layout: layout,
		}},
		nil, pipelines)
	if err := NewError(result); err != nil {
		return vk.NullPipeline, err
	}

	return pipelines[0], nil
}
