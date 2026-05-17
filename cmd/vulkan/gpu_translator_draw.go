package vulkan

import (
	"math"
	"os"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
	"go101.org/nstd"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func (t *GpuTranslator) recordDraw(frame uint64, commandBuffer vk.CommandBuffer, draw *gpu.LiverpoolDrawCall) {
	rtAddress := draw.RtGpuAddress()
	if rtAddress == 0 {
		return
	}

	// Get or create surface and framebuffer.
	width := draw.RtPitchPixels()
	height := draw.VpHeight()
	if height == 0 {
		height = 1080 // Default to something if viewport height is 0
	}
	surface, err := t.GetSurface(SurfaceRequest{
		SurfaceKey: SurfaceKey{
			GpuAddress: rtAddress,
		},
		Format: translateColorFormat(draw.RtFormat, draw.RtNumberType, draw.RtCompSwap),
		Width:  width,
		Height: height,
	})
	if err != nil {
		return
	}
	fb, err := t.GetFramebuffer(FramebufferRequest{
		ImageView: surface.Value.imageView,
		FramebufferKey: FramebufferKey{
			GpuAddress: rtAddress,
			Format:     surface.Value.format,
			Width:      surface.Value.width,
			Height:     surface.Value.height,
		},
	})
	if err != nil {
		return
	}

	// Force load SPIR-V shaders.
	vsSpirv := t.GetShader(draw.VertexShader)
	if draw.EvalShader != nil {
		t.GetShader(draw.EvalShader)
	}
	if draw.HullShader != nil {
		t.GetShader(draw.HullShader)
	}
	if draw.GeometryShader != nil {
		t.GetShader(draw.GeometryShader)
	}
	psSpirv := t.GetShader(draw.PixelShader)

	// Get shader modules.
	vsModule, err := t.GetShaderModule(vsSpirv)
	if err != nil {
		return
	}
	psModule, err := t.GetShaderModule(psSpirv)
	if err != nil {
		return
	}

	var gsModule vk.ShaderModule
	if draw.PrimType == 17 { // RECTLIST
		bytes, err := os.ReadFile("temp/shaders/shader_rectlist.spv")
		if err != nil {
			return
		}
		gsModule, err = t.GetShaderModuleFromBytes(common.SpvBytesToWords(bytes))
		if err != nil {
			return
		}
	} else if draw.GeometryShader != nil {
		gsSpirv := t.GetShader(draw.GeometryShader)
		gsModule, err = t.GetShaderModule(gsSpirv)
		if err != nil {
			return
		}
	}

	// Get pipeline for shader modules and parameters.
	colorWriteMask := draw.RtTargetMask
	if (draw.RtColorControl>>4)&0x7 == 0 { // CB_DISABLE
		colorWriteMask = 0
	}

	logicOpEnable := vk.Bool32(vk.False)
	logicOp := vk.LogicOpCopy
	if (draw.RtBlendControl>>30)&1 == 0 { // Blending disabled
		rop3 := (draw.RtColorControl >> 16) & 0xFF
		if rop3 != 0xCC {
			logicOpEnable = vk.Bool32(vk.True)
			logicOp = translateLogicOp(rop3)
		}
	}

	pipeline, err := t.GetPipeline(GraphicsPipelineRequest{
		VertexModule:   vsModule,
		GeometryModule: gsModule,
		FragmentModule: psModule,
		RenderPass:     fb.RenderPass,
		GraphicsPipelineKey: GraphicsPipelineKey{
			VertexModuleAddress:   draw.VertexShader.Address,
			FragmentModuleAddress: draw.PixelShader.Address,
			RenderTargetAddress:   rtAddress,
			RenderPass:            fb.RenderPass,

			Width:             surface.Value.width,
			Height:            surface.Value.height,
			PrimType:          draw.PrimType,
			BlendAttachment:   translateBlendControl(draw.RtBlendControl, colorWriteMask, draw.RtBlendBypass),
			DepthStencilState: translateDepthControl(draw.DbDepthControl, draw.DbStencilControl),
			LogicOpEnable:     logicOpEnable,
			LogicOp:           logicOp,

			DbKillEnable:           draw.DbKillEnable,
			DbCoverageToMaskEnable: draw.DbCoverageToMaskEnable,
			DbAlphaToMaskDisable:   draw.DbAlphaToMaskDisable,
		},
	})
	if err != nil {
		return
	}

	// Select render pass and clear on first use in frame.
	var renderPass vk.RenderPass
	var clearValueCount uint32
	var clearValues []vk.ClearValue
	if surface.frameUsed < frame {
		renderPass = fb.RenderPass
		clearValueCount = 1
		clearColor := vk.ClearValue{}
		clearColor.SetColor([]float32{0.8, 0.8, 0.8, 1.0})
		clearValues = []vk.ClearValue{clearColor}
		surface.frameUsed = frame

		// Transition surface to General if it's the first use.
		if surface.firstUse {
			t.imageBarrier(commandBuffer, surface.Value.image,
				vk.ImageLayoutUndefined, vk.ImageLayoutGeneral,
				0, vk.AccessFlags(vk.AccessColorAttachmentWriteBit|vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
				vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit|vk.PipelineStageAllGraphicsBit|vk.PipelineStageComputeShaderBit))
			surface.firstUse = false
		}
	} else {
		renderPass = fb.RenderPassNoClear
		clearValueCount = 0
	}
	vk.CmdBeginRenderPass(commandBuffer, &vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      renderPass,
		Framebuffer:     fb.Framebuffer,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: surface.Value.width, Height: surface.Value.height}},
		ClearValueCount: clearValueCount,
		PClearValues:    clearValues,
	}, vk.SubpassContentsInline)

	// Bind pipeline and setup scissor/viewport.
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, pipeline)
	t.setDynamicState(commandBuffer, draw, surface)

	// Bind bindless descriptor set.
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, t.pipelineLayout, spirvStructs.DescriptorSetSlotBindless, 1, []vk.DescriptorSet{t.bindlessDescriptorSet}, 0, nil)

	// Bind discovery descriptor set.
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, t.pipelineLayout, spirvStructs.DescriptorSetSlotDiscovery, 1, []vk.DescriptorSet{t.discoveryDescriptorSet}, 0, nil)

	// Get buffer addresses.
	t.userDataBuffersMutex.Lock()
	userDataOffset := t.userDataOffsets[draw.UserDataHash]
	t.userDataBuffersMutex.Unlock()
	userData := gpu.GlobalUserDataSnapshots[draw.UserDataHash]

	// Bind texel buffers for vertex.
	vsFormatSizes, vsFormatStrides := t.BindTexelBuffers(commandBuffer, userData[:], gcn.GcnShaderStageVertex, spirvStructs.DescriptorSetSlotTexel, vk.PipelineBindPointGraphics)

	// Push constants to vertex shader.
	pushDataVs := spirvStructs.PushConstants{
		UserDataAddress:          t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:   GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress:  GlobalGpuAllocator.DeviceAddress,
		TexelBuffer0FormatSize:   vsFormatSizes[0],
		TexelBuffer1FormatSize:   vsFormatSizes[1],
		TexelBuffer2FormatSize:   vsFormatSizes[2],
		TexelBuffer3FormatSize:   vsFormatSizes[3],
		TexelBuffer0FormatStride: vsFormatStrides[0],
		TexelBuffer1FormatStride: vsFormatStrides[1],
		TexelBuffer2FormatStride: vsFormatStrides[2],
		TexelBuffer3FormatStride: vsFormatStrides[3],
		UserSgprCount:            gpu.DecodeUserSgprCount(draw.VertexShRsrc2),
	}
	vk.CmdPushConstants(
		commandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageComputeBit),
		0, spirvStructs.PushConstantsSize,
		unsafe.Pointer(&pushDataVs),
	)

	// Bind texel buffers for fragment.
	fsFormatSizes, fsFormatStrides := t.BindTexelBuffers(commandBuffer, userData[:], gcn.GcnShaderStageFragment, spirvStructs.DescriptorSetSlotTexelSecondary, vk.PipelineBindPointGraphics)

	// Push constants to fragment shader.
	pushDataFs := spirvStructs.PushConstants{
		UserDataAddress:          t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:   GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress:  GlobalGpuAllocator.DeviceAddress,
		TexelBuffer0FormatSize:   fsFormatSizes[0],
		TexelBuffer1FormatSize:   fsFormatSizes[1],
		TexelBuffer2FormatSize:   fsFormatSizes[2],
		TexelBuffer3FormatSize:   fsFormatSizes[3],
		TexelBuffer0FormatStride: fsFormatStrides[0],
		TexelBuffer1FormatStride: fsFormatStrides[1],
		TexelBuffer2FormatStride: fsFormatStrides[2],
		TexelBuffer3FormatStride: fsFormatStrides[3],
		UserSgprCount:            gpu.DecodeUserSgprCount(draw.PixelShRsrc2),
		ShaderRsrc2:              draw.PixelShRsrc2,
	}
	vk.CmdPushConstants(
		commandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		spirvStructs.PushConstantsSize, spirvStructs.PushConstantsSize,
		unsafe.Pointer(&pushDataFs),
	)

	// Draw.
	logger.Printf("[%s] Drawing %s vertices (vertex=%s, fragment=%s, userData=%s, topology=%s, indexed=%s).\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Green.Sprint(draw.VertexCount),
		color.Yellow.Sprintf("0x%X", draw.VertexShader.Address),
		color.Yellow.Sprintf("0x%X", draw.PixelShader.Address),
		color.Yellow.Sprintf("0x%X", draw.UserDataHash),
		color.Green.Sprint(draw.PrimType),
		color.Green.Sprint(nstd.Btoi(draw.IsIndexed)),
	)
	if draw.IsIndexed {
		targetBuffer, relativeOffset, err := t.GetBufferFromAddress(draw.IndexBaseAddress)
		if err != nil {
			panic(err)
		}

		if targetBuffer != vk.NullBuffer {
			indexType := vk.IndexTypeUint16
			if draw.IndexType == 1 {
				indexType = vk.IndexTypeUint32
			}
			vk.CmdBindIndexBuffer(commandBuffer, targetBuffer, vk.DeviceSize(relativeOffset), indexType)
			vk.CmdDrawIndexed(commandBuffer, draw.IndexCount, draw.InstanceCount, 0, int32(draw.BaseVertexOffset), 0)
		} else {
			vk.CmdDraw(commandBuffer, draw.VertexCount, draw.InstanceCount, 0, 0)
		}
	} else {
		vk.CmdDraw(commandBuffer, draw.VertexCount, draw.InstanceCount, 0, 0)
	}

	vk.CmdEndRenderPass(commandBuffer)
}

func (t *GpuTranslator) setDynamicState(commandBuffer vk.CommandBuffer, draw *gpu.LiverpoolDrawCall, surface *GpuSurface) {
	// Derive viewport from GCN scale/offset registers.
	vpWidth := draw.VpXScale * 2
	vpHeight := draw.VpYScale * 2
	vpX, vpY := draw.VpXOffset-vpWidth/2, draw.VpYOffset-vpHeight/2

	if vpWidth <= 0 || vpHeight == 0 {
		vpWidth, vpHeight = float32(surface.Value.width), float32(surface.Value.height)
		vpX, vpY = 0, 0
	}
	vk.CmdSetViewport(commandBuffer, 0, 1, []vk.Viewport{{
		X: vpX, Y: vpY,
		Width: vpWidth, Height: vpHeight,
		MinDepth: draw.VpZMin,
		MaxDepth: draw.VpZMax,
	}})

	// Setup the scissor from TL/BR registers.
	scissorX, scissorY, scissorW, scissorH := draw.ScissorRect()
	if scissorW <= 0 || scissorH <= 0 {
		scissorW = int(surface.Value.width)
		scissorH = int(surface.Value.height)
		scissorX, scissorY = 0, 0
	}
	vk.CmdSetScissor(commandBuffer, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{X: int32(scissorX), Y: int32(scissorY)},
		Extent: vk.Extent2D{Width: uint32(scissorW), Height: uint32(scissorH)},
	}})

	// Setup blend constants.
	vk.CmdSetBlendConstants(commandBuffer, &[4]float32{
		math.Float32frombits(draw.BlendRed),
		math.Float32frombits(draw.BlendGreen),
		math.Float32frombits(draw.BlendBlue),
		math.Float32frombits(draw.BlendAlpha),
	})
}
