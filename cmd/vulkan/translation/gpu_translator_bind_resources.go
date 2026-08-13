package translation

import (
	"fmt"
	"maps"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"

	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvCommon "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) BindResources(frame uint64, bind *gpu.LiverpoolBindResources) {
	t.EndRenderPass()

	// Get buffer addresses.
	t.userDataBuffersMutex.Lock()
	userData, _ := gpu.GlobalUserDataSnapshots[bind.UserDataHash]
	t.userDataBuffersMutex.Unlock()

	// Gather shaders.
	var shaders []*spirv.SpirvShader
	if bind.VertexShader != nil {
		// Parse fetch shader layout.
		var fetchInstructions []*gcnSpec.Instruction
		fetchShaderAddress := t.GetFetchShaderPC(bind.VertexShader, userData[:])
		if fetchShaderAddress != 0 {
			fetchShader := gpu.GlobalLiverpool.GetShader(gcn.GcnShaderStageFetch, fetchShaderAddress)
			if fetchShader != nil {
				fetchInstructions = ParseFetchShaderInstructions(fetchShader)
			}
		}
		bind.VertexContext.FetchShaderAddress = fetchShaderAddress
		bind.VertexContext.FetchShaderInstructions = fetchInstructions
		bind.VertexContext.MubufFormats = t.GetMubufFormats(bind.VertexShader, userData[:])

		vsSpirv, vsKey := t.GetShaderWithContext(bind.VertexShader, bind.VertexContext)
		t.activeVertexShader = vsSpirv
		t.activeVertexShaderKey = vsKey
		shaders = append(shaders, vsSpirv)
	} else {
		t.activeVertexShader = nil
		t.activeVertexShaderKey = spirvCommon.SpirvShaderKey{}
	}
	if bind.FragmentShader != nil {
		bind.FragmentContext.MubufFormats = t.GetMubufFormats(bind.FragmentShader, userData[:])
		psSpirv, psKey := t.GetShaderWithContext(bind.FragmentShader, bind.FragmentContext)
		t.activeFragmentShader = psSpirv
		t.activeFragmentShaderKey = psKey
		shaders = append(shaders, psSpirv)
	} else {
		t.activeFragmentShader = nil
		t.activeFragmentShaderKey = spirvCommon.SpirvShaderKey{}
	}
	if bind.GeometryShader != nil {
		gsSpirv, gsKey := t.GetShaderWithContext(bind.GeometryShader, spirvCommon.SpirvShaderContext(nil))
		t.activeGeometryShader = gsSpirv
		t.activeGeometryShaderKey = gsKey
		shaders = append(shaders, gsSpirv)
	}
	if bind.ComputeShader != nil {
		bind.ComputeContext.MubufFormats = t.GetMubufFormats(bind.ComputeShader, userData[:])
		csSpirv, csKey := t.GetShaderWithContext(bind.ComputeShader, bind.ComputeContext)
		t.activeComputeShader = csSpirv
		t.activeComputeShaderKey = csKey
		shaders = append(shaders, csSpirv)
	}

	// Bind resources.
	globalSetToBind := t.globalDescriptorPool.DefaultSet(frame)
	imageSetToBind := t.imageDescriptorPool.DefaultSet(frame)
	storeTargets, storeBufferTargets, activeGlobalSet, activeImageSet, err := t.GetBindDescriptorSet(shaders, userData)
	if err != nil {
		panic(err)
	}
	if activeGlobalSet != vk.NullDescriptorSet {
		globalSetToBind = activeGlobalSet
	}
	if activeImageSet != vk.NullDescriptorSet {
		imageSetToBind = activeImageSet
	}
	t.activeComputeStoreTargets = storeTargets
	t.activeComputeStoreBufferTargets = storeBufferTargets

	vk.CmdBindDescriptorSets(
		t.commandBuffer.CommandBuffer, vk.PipelineBindPointGraphics,
		t.pipelineLayout, spirvStructs.DescriptorSetSlotGlobal,
		2, []vk.DescriptorSet{globalSetToBind, imageSetToBind},
		0, nil,
	)
	vk.CmdBindDescriptorSets(
		t.commandBuffer.CommandBuffer, vk.PipelineBindPointCompute,
		t.pipelineLayout, spirvStructs.DescriptorSetSlotGlobal,
		2, []vk.DescriptorSet{globalSetToBind, imageSetToBind},
		0, nil,
	)
}

// GetBindDescriptorSet resolves compute operands, prepares VkImages and returns a descriptor set to bind.
func (t *GpuTranslator) GetBindDescriptorSet(shaders []*spirv.SpirvShader, userData spirvStructs.UserData) ([]*vulkan.VulkanImage, []*vulkan.VulkanImage, vk.DescriptorSet, vk.DescriptorSet, error) {
	// Resolve resources accessed by the shaders.
	var imageAccesses []ResolvedImageAccess
	var bufferAccesses []ResolvedBufferAccess
	var memoryAccesses []ResolvedMemoryAccess
	var allLayouts = make(map[*gcnSpec.Instruction]spirvCommon.ShaderResourceBinding)
	for _, shader := range shaders {
		shaderImageAccesses, shaderBufferAccesses, shaderMemoryAccesses := t.ResolveResources(shader.GcnShader, userData[:])
		imageAccesses = append(imageAccesses, shaderImageAccesses...)
		bufferAccesses = append(bufferAccesses, shaderBufferAccesses...)
		memoryAccesses = append(memoryAccesses, shaderMemoryAccesses...)
		maps.Copy(allLayouts, shader.StaticLayout)
	}
	if len(imageAccesses) == 0 && len(bufferAccesses) == 0 && len(memoryAccesses) == 0 {
		return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, nil
	}

	// Get a descriptor set for each pool.
	activeGlobalSet, err := t.globalDescriptorPool.Get(t.handles, t.currentGuestFrame)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, err
	}
	activeImageSet, err := t.imageDescriptorPool.Get(t.handles, t.currentGuestFrame)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, err
	}

	// Resolve addresses.
	bufferAccessByInstr := make(map[*gcnSpec.Instruction]ResolvedBufferAccess, len(bufferAccesses))
	for _, access := range bufferAccesses {
		bufferAccessByInstr[access.Instruction] = access
	}
	accessByInstr := make(map[*gcnSpec.Instruction]ResolvedImageAccess, len(imageAccesses))
	for _, access := range imageAccesses {
		accessByInstr[access.Instruction] = access
	}
	memoryAccessByInstr := make(map[*gcnSpec.Instruction]ResolvedMemoryAccess, len(memoryAccesses))
	for _, access := range memoryAccesses {
		memoryAccessByInstr[access.Instruction] = access
	}

	// Update address translation buffer.
	ringBuffer := t.addressTranslationRing.Get(t.currentGuestFrame)
	slice := unsafe.Slice((*uint64)(unsafe.Pointer(&ringBuffer.Mapped[t.addressTranslationOffset])), spirvStructs.AddressTranslationBlockEntries)
	for i := range slice {
		slice[i] = 0 // Clear previous bindings.
	}
	var boundText string
	if logger.LogRenderer {
		boundText = fmt.Sprintf("[Frame %d] Bound memory slot", t.currentGuestFrame)
	}
	for instr, binding := range allLayouts {
		var baseAddress uint64
		switch binding.Type {
		case spirvCommon.SpirvShaderResourceTypeBuffer:
			if access, ok := bufferAccessByInstr[instr]; ok {
				baseAddress = uint64(access.Descriptor.BaseAddress)
			}
		case spirvCommon.SpirvShaderResourceTypeMemory:
			if access, ok := memoryAccessByInstr[instr]; ok {
				baseAddress = uint64(access.BaseAddress)
			}
		}
		if baseAddress != 0 {
			translated := t.TranslateToDeviceAddress(uintptr(baseAddress))
			slice[binding.BindingIndex] = translated
		}
		if logger.LogRenderer {
			boundText += fmt.Sprintf(" %s %d (0x%X->0x%X)", binding.Kind, binding.BindingIndex, baseAddress, slice[binding.BindingIndex])
		}
	}
	vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          activeGlobalSet,
		DstBinding:      spirvStructs.GlobalBindingAddressTranslation,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  vk.DescriptorTypeStorageBuffer,
		PBufferInfo: []vk.DescriptorBufferInfo{{
			Buffer: ringBuffer.Buffer,
			Offset: vk.DeviceSize(t.addressTranslationOffset),
			Range:  vk.DeviceSize(spirvStructs.AddressTranslationBlockSize),
		}},
	}}, 0, nil)
	t.addressTranslationOffset += spirvStructs.AddressTranslationBlockSize
	if len(allLayouts) > 0 && logger.LogRenderer {
		boundText += ".\n"
		logger.Print(boundText)
	}

	// Bind images to slots.
	if logger.LogRenderer {
		boundText = fmt.Sprintf("[Frame %d] Bound image slot", t.currentGuestFrame)
	}
	for instr, binding := range allLayouts {
		if binding.Type != spirvCommon.SpirvShaderResourceTypeImage {
			continue
		}
		access := accessByInstr[instr]
		if access.Descriptor.BaseAddress == 0 {
			continue
		}

		// Get image view and sampler.
		view, err, _ := t.GetImageView(access.Descriptor)
		if err != nil {
			return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, err
		}
		sampler := t.defaultSampler
		if access.Kind == spirvCommon.ImageAccessSample {
			sampler, err = t.GetSampler(*access.Sampler)
			if err != nil {
				return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, err
			}
		}

		// Transition image layout, update descriptor set.
		switch binding.Kind {
		case spirvCommon.ImageAccessLoad, spirvCommon.ImageAccessSample:
			view.Image.BarrierSampledRead(t.commandBuffer)
			t.updateImageDescriptorBinding(activeImageSet, binding.BindingIndex, view, nil, sampler)
		case spirvCommon.ImageAccessStore:
			view.Image.BarrierComputeStorageWrite(t.commandBuffer)
			t.updateImageDescriptorBinding(activeImageSet, binding.BindingIndex, nil, view, sampler)
		}
		if logger.LogRenderer {
			boundText += fmt.Sprintf(" %s %d (0x%X/%dx%d)", binding.Kind, binding.BindingIndex, access.Descriptor.BaseAddress, access.Descriptor.Width, access.Descriptor.Height)
		}
	}
	if len(allLayouts) > 0 && logger.LogRenderer {
		boundText += ".\n"
		logger.Print(boundText)
	}

	// Bind buffers to slots (not really).
	if logger.LogRenderer {
		boundText = fmt.Sprintf("[Frame %d] Bound buffer slot", t.currentGuestFrame)
	}
	for instr, binding := range allLayouts {
		if binding.Type != spirvCommon.SpirvShaderResourceTypeBuffer {
			continue
		}
		access := bufferAccessByInstr[instr]
		if access.Descriptor.BaseAddress == 0 {
			continue
		}
		if logger.LogRenderer {
			boundText += fmt.Sprintf(" %s %d (0x%X/%d)", binding.Kind, binding.BindingIndex, access.Descriptor.BaseAddress, access.Descriptor.NumRecords)
		}
	}
	if len(allLayouts) > 0 && logger.LogRenderer {
		boundText += ".\n"
		logger.Print(boundText)
	}

	/* for _, shader := range shaders {
		offset := spirvStructs.GcnStageToUserDataOffset[shader.GcnShader.Stage]
		logger.Printf("Using user data for vv: %+v\n", userData[offset:offset+16])
	} */

	// Download accessed buffer images.
	loadBufferTargets, err := t.ResolveBufferTargets(bufferAccesses, spirvCommon.BufferAccessLoad)
	for _, image := range loadBufferTargets {
		if !image.ShouldDownloadFromVkImage() {
			continue
		}
		if err = image.DownloadFromVkImage(t.handles, t.commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
			return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, err
		}
	}

	// Find out which resources were written to.
	storeTargets, err := t.ResolveImageTargets(imageAccesses, spirvCommon.ImageAccessStore)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, err
	}
	storeBufferTargets, err := t.ResolveBufferTargets(bufferAccesses, spirvCommon.BufferAccessStore)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, vk.NullDescriptorSet, err
	}

	return storeTargets, storeBufferTargets, activeGlobalSet, activeImageSet, nil
}

func (t *GpuTranslator) updateImageDescriptorBinding(set vk.DescriptorSet, index uint32, sampledView *vulkan.VulkanImageView, storageView *vulkan.VulkanImageView, sampler vk.Sampler) {
	if sampledView != nil && sampledView.ImageView != vk.NullImageView {
		dstBinding := uint32(spirvStructs.ImageBindingSampledImages2D)
		if sampledView.Image != nil {
			switch sampledView.Image.FirstDescriptor.InferredType() {
			case gcn.GcnImageTypeBuffer, gcn.GcnImageTypeColor1D:
				dstBinding = spirvStructs.ImageBindingSampledImages1D
			case gcn.GcnImageTypeColor3D:
				dstBinding = spirvStructs.ImageBindingSampledImages3D
			case gcn.GcnImageTypeCubeOrArray, gcn.GcnImageTypeColor1DArray, gcn.GcnImageTypeColor2DArray, gcn.GcnImageTypeColor2DMsaaArray:
				// Cube and Arrays use 2DArray view type.
				dstBinding = spirvStructs.ImageBindingSampledImages2DArray
			}
		}

		vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          set,
			DstBinding:      dstBinding,
			DstArrayElement: index,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo: []vk.DescriptorImageInfo{{
				Sampler:     sampler,
				ImageView:   sampledView.ImageView,
				ImageLayout: vk.ImageLayoutGeneral,
			}},
		}}, 0, nil)
	}
	if storageView != nil && storageView.StorageImageView != vk.NullImageView {
		dstBinding := uint32(spirvStructs.ImageBindingStorageImages2D)
		if storageView.Image != nil {
			switch storageView.Image.FirstDescriptor.InferredType() {
			case gcn.GcnImageTypeBuffer, gcn.GcnImageTypeColor1D:
				dstBinding = spirvStructs.ImageBindingStorageImages1D
			case gcn.GcnImageTypeColor3D:
				dstBinding = spirvStructs.ImageBindingStorageImages3D
			case gcn.GcnImageTypeCubeOrArray, gcn.GcnImageTypeColor1DArray, gcn.GcnImageTypeColor2DArray, gcn.GcnImageTypeColor2DMsaaArray:
				dstBinding = spirvStructs.ImageBindingStorageImages2DArray
			}
		}

		vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          set,
			DstBinding:      dstBinding,
			DstArrayElement: index,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeStorageImage,
			PImageInfo: []vk.DescriptorImageInfo{{
				ImageView:   storageView.StorageImageView,
				ImageLayout: vk.ImageLayoutGeneral,
			}},
		}}, 0, nil)
	}
}
