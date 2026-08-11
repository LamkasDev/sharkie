package vulkan

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	"github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
)

type GraphicsPipelineRequest struct {
	VertexModule      vk.ShaderModule
	TessControlModule vk.ShaderModule
	TessEvalModule    vk.ShaderModule
	GeometryModule    vk.ShaderModule
	FragmentModule    vk.ShaderModule
	ColorFormat       vk.Format
	DepthFormat       vk.Format

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

	// GCN registers.
	PaSuScModeCntl            reg.PaSuScModeCntl
	PaScModeCntl0             reg.PaScModeCntl0
	DbShaderControl           reg.DbShaderControl
	PaScAaConfig              reg.PaScAaConfig
	VgtMultiPrimIbResetEn     reg.VgtMultiPrimIbResetEn
	PaSuLineCntl              reg.PaSuLineCntl
	PaScAaMaskX0y0X1y0        reg.PaScAaMaskX0y0X1y0
	PaScAaMaskX0y1X1y1        reg.PaScAaMaskX0y1X1y1
	SpiShaderColFormat        reg.SpiShaderColFormat
	SpiShaderZFormat          reg.SpiShaderZFormat
	SpiVsOutConfig            reg.SpiVsOutConfig
	PaClVsOutCntl             reg.PaClVsOutCntl
	CbColorInfo0              reg.CbColorInfo
	CbTargetMask              reg.CbTargetMask
	CbShaderMask              reg.CbShaderMask
	CbColorControl            reg.CbColorControl
	CbColorAttrib             reg.CbColorAttrib
	PaSuPolyOffsetClamp       reg.PaSuPolyOffsetClamp
	PaSuPolyOffsetFrontScale  reg.PaSuPolyOffsetFrontScale
	PaSuPolyOffsetFrontOffset reg.PaSuPolyOffsetFrontOffset
	PaSuPolyOffsetBackScale   reg.PaSuPolyOffsetBackScale
	PaSuPolyOffsetBackOffset  reg.PaSuPolyOffsetBackOffset
	DbRenderControl           reg.DbRenderControl

	// Render control flags.
	BlendAttachment   vk.PipelineColorBlendAttachmentState
	DepthStencilState vk.PipelineDepthStencilStateCreateInfo
	LogicOpEnable     vk.Bool32
	LogicOp           vk.LogicOp
}

type ComputePipelineRequest struct {
	ComputeModule vk.ShaderModule

	ComputePipelineKey
}

type ComputePipelineKey struct {
	ComputeModuleAddress uintptr
}

func CreateGraphicsPipeline(handles *VulkanHandles, request GraphicsPipelineRequest, layout vk.PipelineLayout, cache vk.PipelineCache) (vk.Pipeline, error) {
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
	if request.FragmentModule != vk.NullShaderModule {
		stages = append(stages, vk.PipelineShaderStageCreateInfo{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageFragmentBit,
			Module: request.FragmentModule,
			PName:  "main\x00",
			PNext:  unsafe.Pointer(subgroupSizeFs),
		})
	}

	// No vertex input.
	vertexInput := vk.PipelineVertexInputStateCreateInfo{
		SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
	}
	inputAssembly := gcn.CreateInputAssemblyState(request.PrimType, request.VgtMultiPrimIbResetEn)

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
	if request.PaScModeCntl0.LineStippleEnable() {
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

	// Setup rasterization state.
	raster, provokingVertex := gcn.CreateRasterizationState(request.PaSuScModeCntl, request.PaSuLineCntl, request.PaSuPolyOffsetClamp, request.PaSuPolyOffsetFrontScale, request.PaSuPolyOffsetFrontOffset, request.PaSuPolyOffsetBackScale, request.PaSuPolyOffsetBackOffset)
	raster.PNext = unsafe.Pointer(&provokingVertex)
	depthStencil := request.DepthStencilState
	if request.DepthTargetAddress == 0 {
		depthStencil.DepthTestEnable = vk.False
		depthStencil.DepthWriteEnable = vk.False
		depthStencil.StencilTestEnable = vk.False
	}

	// Setup anti-aliasing.
	multisample := gcn.CreateMultisampleState(request.PaScAaConfig, request.PaScModeCntl0, request.DbShaderControl, request.PaScAaMaskX0y0X1y0, request.PaScAaMaskX0y1X1y1)

	// Setup blend attachments.
	blendAttachments := gcn.CreateBlendAttachments(request.BlendAttachment, request.CbTargetMask, request.CbShaderMask, request.SpiShaderColFormat, request.CbColorControl)

	// Setup blending.
	blend := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   request.LogicOpEnable,
		LogicOp:         request.LogicOp,
		AttachmentCount: 8,
		PAttachments:    blendAttachments,
	}

	// Setup dynamic rendering info.
	colorFormats := make([]vk.Format, 8)
	for i := 0; i < 8; i++ {
		colorFormats[i] = vk.FormatUndefined
	}
	if request.ColorFormat != vk.FormatUndefined {
		colorFormats[0] = request.ColorFormat
	}
	renderingInfo := &vk.PipelineRenderingCreateInfo{
		SType:                   vk.StructureTypePipelineRenderingCreateInfo,
		ColorAttachmentCount:    8,
		PColorAttachmentFormats: colorFormats,
	}
	if request.DepthFormat != vk.FormatUndefined {
		renderingInfo.DepthAttachmentFormat = request.DepthFormat
		renderingInfo.StencilAttachmentFormat = request.DepthFormat
	}

	cRenderingInfo, freeRenderingInfo := CreatePipelineRenderingCreateInfoC(renderingInfo)
	defer freeRenderingInfo()

	// Create pipeline.
	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		PNext:               cRenderingInfo,
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
		RenderPass:          vk.NullRenderPass,
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
