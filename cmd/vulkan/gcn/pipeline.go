package gcn

import (
	"math"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

func CreateInputAssemblyState(primType uint32, resetEn reg.VgtMultiPrimIbResetEn) vk.PipelineInputAssemblyStateCreateInfo {
	topology := TranslateTopology(primType)
	return vk.PipelineInputAssemblyStateCreateInfo{
		SType:    vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology: topology,
		PrimitiveRestartEnable: vk.Bool32(
			nstd.Btoi(resetEn.Enable() &&
				(topology == vk.PrimitiveTopologyLineStrip ||
					topology == vk.PrimitiveTopologyTriangleStrip ||
					topology == vk.PrimitiveTopologyTriangleFan ||
					topology == vk.PrimitiveTopologyPatchList)),
		),
	}
}

func CreateRasterizationState(paSuScModeCntl reg.PaSuScModeCntl, paSuLineCntl reg.PaSuLineCntl, clamp reg.PaSuPolyOffsetClamp, fScale reg.PaSuPolyOffsetFrontScale, fOffset reg.PaSuPolyOffsetFrontOffset, bScale reg.PaSuPolyOffsetBackScale, bOffset reg.PaSuPolyOffsetBackOffset) (vk.PipelineRasterizationStateCreateInfo, vk.PipelineRasterizationProvokingVertexStateCreateInfo) {
	frontFace := vk.FrontFaceCounterClockwise
	if paSuScModeCntl.Face() {
		frontFace = vk.FrontFaceClockwise
	}
	cullMode := vk.CullModeNone
	if paSuScModeCntl.CullFront() {
		cullMode |= vk.CullModeFrontBit
	}
	if paSuScModeCntl.CullBack() {
		cullMode |= vk.CullModeBackBit
	}
	polygonMode := vk.PolygonModeFill
	switch paSuScModeCntl.PolyMode() {
	case 1:
		polygonMode = vk.PolygonModeLine
	case 2:
		polygonMode = vk.PolygonModePoint
	}

	provokingVertex := vk.PipelineRasterizationProvokingVertexStateCreateInfo{
		SType:               vk.StructureTypePipelineRasterizationProvokingVertexStateCreateInfo,
		ProvokingVertexMode: vk.ProvokingVertexModeFirstVertex,
	}
	if paSuScModeCntl.ProvokingVertexLast() {
		provokingVertex.ProvokingVertexMode = vk.ProvokingVertexModeLastVertex
	}

	lineWidth := float32(paSuLineCntl.Width()) / 8.0 // 1/2 width in 1/16th subpixels? Or 1/8th? Usually /8.0 works.
	if lineWidth <= 0.0 {
		lineWidth = 1.0 // Vulkan requires >= 1.0, though wideLines feature is needed for > 1.0
	}

	depthBiasEnable := vk.Bool32(vk.False)
	var depthBiasConstantFactor, depthBiasClamp, depthBiasSlopeFactor float32

	if paSuScModeCntl.PolyOffsetFrontEnable() && !paSuScModeCntl.CullFront() {
		depthBiasEnable = vk.Bool32(vk.True)
		depthBiasConstantFactor = math.Float32frombits(fOffset.Offset())
		depthBiasSlopeFactor = math.Float32frombits(fScale.Scale())
		depthBiasClamp = math.Float32frombits(clamp.Clamp())
	} else if paSuScModeCntl.PolyOffsetBackEnable() && !paSuScModeCntl.CullBack() {
		depthBiasEnable = vk.Bool32(vk.True)
		depthBiasConstantFactor = math.Float32frombits(bOffset.Offset())
		depthBiasSlopeFactor = math.Float32frombits(bScale.Scale())
		depthBiasClamp = math.Float32frombits(clamp.Clamp())
	}

	raster := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        vk.False,
		PolygonMode:             polygonMode,
		CullMode:                vk.CullModeFlags(cullMode),
		FrontFace:               frontFace,
		LineWidth:               lineWidth,
		DepthBiasEnable:         depthBiasEnable,
		DepthBiasConstantFactor: depthBiasConstantFactor,
		DepthBiasClamp:          depthBiasClamp,
		DepthBiasSlopeFactor:    depthBiasSlopeFactor,
	}
	return raster, provokingVertex
}

func CreateMultisampleState(aaConfig reg.PaScAaConfig, dbShaderControl reg.DbShaderControl, aaMask1 reg.PaScAaMaskX0y0X1y0, aaMask2 reg.PaScAaMaskX0y1X1y1) vk.PipelineMultisampleStateCreateInfo {
	mask := aaMask1.AaMaskX0y0() & aaMask1.AaMaskX1y0() & aaMask2.AaMaskX0y1() & aaMask2.AaMaskX1y1()
	var pSampleMask []vk.SampleMask
	if mask != 0xFFFF && mask != 0 {
		pSampleMask = []vk.SampleMask{vk.SampleMask(mask)}
	}

	return vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  TranslateMsaaSamples(aaConfig.MsaaNumSamples()),
		SampleShadingEnable:   vk.False,
		MinSampleShading:      1.0,
		PSampleMask:           pSampleMask,
		AlphaToCoverageEnable: vk.Bool32(nstd.Btoi((dbShaderControl.KillEnable() || dbShaderControl.CoverageToMaskEnable()) && !dbShaderControl.AlphaToMaskDisable())),
		AlphaToOneEnable:      vk.False,
	}
}
