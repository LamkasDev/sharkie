package renderer

import (
	"fmt"
	"unsafe"

	as "github.com/LamkasDev/asche"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/structs/spirv"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) createCommandPool() error {
	var pool vk.CommandPool
	result := vk.CreateCommandPool(t.handles.Device, &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: t.handles.GraphicsQueueFamilyIndex,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}, nil, &pool)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.pool = pool

	return nil
}

func (t *GpuTranslator) allocSurface(s *GpuSurface) error {
	// Create the render-target image.
	var image vk.Image
	result := vk.CreateImage(t.handles.Device, &vk.ImageCreateInfo{
		SType:         vk.StructureTypeImageCreateInfo,
		ImageType:     vk.ImageType2d,
		Format:        s.Format,
		Extent:        vk.Extent3D{Width: s.Width, Height: s.Height, Depth: 1},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        vk.ImageTilingOptimal,
		Usage:         vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageSampledBit | vk.ImageUsageTransferSrcBit),
		SharingMode:   vk.SharingModeExclusive,
		InitialLayout: vk.ImageLayoutUndefined,
	}, nil, &image)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateImage: %w", err)
	}
	s.image = image

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(t.handles.Device, s.image, &memReqs)
	memReqs.Deref()

	var imageMem vk.DeviceMemory
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyDeviceLocalBit),
	}, nil, &imageMem)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkAllocateMemory: %w", err)
	}
	s.imageMem = imageMem
	vk.BindImageMemory(t.handles.Device, s.image, s.imageMem, 0)

	var imageView vk.ImageView
	result = vk.CreateImageView(t.handles.Device, &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    s.image,
		ViewType: vk.ImageViewType2d,
		Format:   s.Format,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			LevelCount: 1,
			LayerCount: 1,
		},
	}, nil, &imageView)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateImageView: %w", err)
	}
	s.imageView = imageView

	var sampler vk.Sampler
	result = vk.CreateSampler(t.handles.Device, &vk.SamplerCreateInfo{
		SType:        vk.StructureTypeSamplerCreateInfo,
		MagFilter:    vk.FilterNearest,
		MinFilter:    vk.FilterNearest,
		AddressModeU: vk.SamplerAddressModeClampToEdge,
		AddressModeV: vk.SamplerAddressModeClampToEdge,
		AddressModeW: vk.SamplerAddressModeClampToEdge,
	}, nil, &sampler)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateSampler: %w", err)
	}
	s.sampler = sampler

	var renderPass vk.RenderPass
	result = vk.CreateRenderPass(t.handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments: []vk.AttachmentDescription{{
			Format:         s.Format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutColorAttachmentOptimal,
			FinalLayout:    vk.ImageLayoutShaderReadOnlyOptimal,
		}},
		SubpassCount: 1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
		}},
	}, nil, &renderPass)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateRenderPass: %w", err)
	}
	s.renderPass = renderPass

	var framebuffer vk.Framebuffer
	result = vk.CreateFramebuffer(t.handles.Device, &vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		RenderPass:      s.renderPass,
		AttachmentCount: 1,
		PAttachments:    []vk.ImageView{s.imageView},
		Width:           s.Width,
		Height:          s.Height,
		Layers:          1,
	}, nil, &framebuffer)
	if err := as.NewError(result); err != nil {
		return fmt.Errorf("vkCreateFramebuffer: %w", err)
	}
	s.framebuffer = framebuffer

	return nil
}

func (t *GpuTranslator) loadStubShaders() error {
	var err error
	var vertModule vk.ShaderModule
	vertModule, err = loadShaderModule(t.handles.Device, "data/shaders/stub_vert.spv")
	if err != nil {
		return fmt.Errorf("stub_vert.spv: %w", err)
	}
	t.stubVertShader = vertModule
	var fragModule vk.ShaderModule
	fragModule, err = loadShaderModule(t.handles.Device, "data/shaders/stub_frag.spv")
	if err != nil {
		return fmt.Errorf("stub_frag.spv: %w", err)
	}
	t.stubFragShader = fragModule

	return nil
}

func (t *GpuTranslator) createStubPipelineLayout() error {
	var layout vk.PipelineLayout
	result := vk.CreatePipelineLayout(t.handles.Device, &vk.PipelineLayoutCreateInfo{
		SType: vk.StructureTypePipelineLayoutCreateInfo,
		PPushConstantRanges: []vk.PushConstantRange{{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       16,
		}},
		PushConstantRangeCount: 1,
	}, nil, &layout)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.stubPipelineLayout = layout

	return nil
}

func (t *GpuTranslator) GetShaderModule(shader *SpirvShader) (vk.ShaderModule, error) {
	// Get already created shader module.
	t.shaderModulesMutex.Lock()
	mod, ok := t.shaderModules[shader.Address]
	t.shaderModulesMutex.Unlock()
	if ok {
		return mod, nil
	}

	// Create the shader module.
	var module vk.ShaderModule
	result := vk.CreateShaderModule(t.handles.Device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(shader.Code) * 4),
		PCode:    shader.Code,
	}, nil, &module)
	if err := as.NewError(result); err != nil {
		return vk.NullShaderModule, fmt.Errorf("vkCreateShaderModule 0x%X: %w", shader.Address, err)
	}
	t.shaderModulesMutex.Lock()
	t.shaderModules[shader.Address] = module
	t.shaderModulesMutex.Unlock()

	return module, nil
}

func (t *GpuTranslator) GetPipeline(key GpuTranslatorPipelineKey, psModule vk.ShaderModule, renderPass vk.RenderPass, width, height uint32) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.pipelinesMutex.Lock()
	pipeline, ok := t.pipelines[key]
	t.pipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := t.createPipelineFromModules(t.stubVertShader, psModule, renderPass, width, height)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createCompiledPipeline 0x%X: %w", key.PixelShaderAddress, err)
	}
	t.pipelinesMutex.Lock()
	t.pipelines[key] = pipeline
	t.pipelinesMutex.Unlock()

	return pipeline, nil
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

	raster := vk.PipelineRasterizationStateCreateInfo{
		SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
		PolygonMode: vk.PolygonModeFill,
		CullMode:    vk.CullModeFlags(vk.CullModeNone),
		FrontFace:   vk.FrontFaceCounterClockwise,
		LineWidth:   1.0,
	}

	multisample := vk.PipelineMultisampleStateCreateInfo{
		SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples: vk.SampleCount1Bit,
	}

	// Opaque blend.
	blendAttach := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(
			vk.ColorComponentRBit | vk.ColorComponentGBit |
				vk.ColorComponentBBit | vk.ColorComponentABit),
	}
	blend := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
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

func (t *GpuTranslator) UpdateConstRamBuffers(draws []LiverpoolDrawCall) {
	t.constRamBuffersMutex.Lock()
	defer t.constRamBuffersMutex.Unlock()

	// Find unique hashes in current draw calls.
	activeHashes := make(map[uint32]bool)
	for i := range draws {
		activeHashes[draws[i].ConstRamHash] = true
	}

	// Delete buffers that are no longer active.
	for h, b := range t.constRamBuffers {
		if !activeHashes[h] {
			vk.DestroyBuffer(t.handles.Device, b, nil)
			vk.FreeMemory(t.handles.Device, t.constRamBufferMems[h], nil)
			delete(t.constRamBuffers, h)
			delete(t.constRamBufferMems, h)
		}
	}

	// Create buffers for new active hashes.
	for hash := range activeHashes {
		if _, ok := t.constRamBuffers[hash]; ok {
			continue
		}

		// Get contents from global state.
		contents, ok := GlobalConstRamSnapshots[hash]
		if !ok {
			continue
		}

		// Allocate and upload.
		size := vk.DeviceSize(len(contents) * 4)
		buffer, mem, err := t.allocBuffer(size,
			vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageUniformBufferBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
		if err != nil {
			panic(fmt.Errorf("allocConstRamBuffer: %w", err))
		}

		data := t.handles.MapMemory(mem, size)
		copy(data, unsafe.Slice((*byte)(unsafe.Pointer(&contents[0])), size))
		vk.UnmapMemory(t.handles.Device, mem)

		t.constRamBuffers[hash] = buffer
		t.constRamBufferMems[hash] = mem
	}
}

func (t *GpuTranslator) allocBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, props vk.MemoryPropertyFlags) (vk.Buffer, vk.DeviceMemory, error) {
	var buffer vk.Buffer
	result := vk.CreateBuffer(t.handles.Device, &vk.BufferCreateInfo{
		SType: vk.StructureTypeBufferCreateInfo,
		Size:  size,
		Usage: usage,
	}, nil, &buffer)
	if err := as.NewError(result); err != nil {
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkCreateBuffer: %w", err)
	}

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(t.handles.Device, buffer, &memReqs)
	memReqs.Deref()

	var mem vk.DeviceMemory
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyFlagBits(props)),
	}, nil, &mem)
	if err := as.NewError(result); err != nil {
		vk.DestroyBuffer(t.handles.Device, buffer, nil)
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkAllocateMemory: %w", err)
	}

	vk.BindBufferMemory(t.handles.Device, buffer, mem, 0)

	return buffer, mem, nil
}

func (t *GpuTranslator) imageBarrier(commandBuffer vk.CommandBuffer, image vk.Image,
	oldLayout, newLayout vk.ImageLayout,
	srcAccess, dstAccess vk.AccessFlags,
	srcStage, dstStage vk.PipelineStageFlags,
) {
	vk.CmdPipelineBarrier(commandBuffer,
		srcStage, dstStage,
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           oldLayout,
			NewLayout:           newLayout,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               image,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1,
				LayerCount: 1,
			},
			SrcAccessMask: srcAccess,
			DstAccessMask: dstAccess,
		}})
}
