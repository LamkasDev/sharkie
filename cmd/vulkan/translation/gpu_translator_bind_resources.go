package translation

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"

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
		fetchShaderAddress := GetFetchShaderPC(bind.VertexShader, userData[:])
		if fetchShaderAddress != 0 {
			fetchShader := gpu.GlobalLiverpool.GetShader(gcn.GcnShaderStageVertex, fetchShaderAddress)
			if fetchShader != nil {
				fetchInstructions = ParseFetchShaderInstructions(fetchShader)
			}
		}
		bind.VertexContext.FetchShaderAddress = fetchShaderAddress
		bind.VertexContext.FetchShaderInstructions = fetchInstructions

		vsSpirv, vsKey := t.GetShaderWithContext(bind.VertexShader, bind.VertexContext)
		t.activeVertexShader = vsSpirv
		t.activeVertexShaderKey = vsKey
		shaders = append(shaders, vsSpirv)
	} else {
		t.activeVertexShader = nil
		t.activeVertexShaderKey = SpirvShaderKey{}
	}
	if bind.FragmentShader != nil {
		psSpirv, psKey := t.GetShaderWithContext(bind.FragmentShader, bind.FragmentContext)
		t.activeFragmentShader = psSpirv
		t.activeFragmentShaderKey = psKey
		shaders = append(shaders, psSpirv)
	} else {
		t.activeFragmentShader = nil
		t.activeFragmentShaderKey = SpirvShaderKey{}
	}
	if bind.GeometryShader != nil {
		gsSpirv, gsKey := t.GetShaderWithContext(bind.GeometryShader, spirvCommon.SpirvShaderContext(nil))
		t.activeGeometryShader = gsSpirv
		t.activeGeometryShaderKey = gsKey
		shaders = append(shaders, gsSpirv)
	}
	if bind.ComputeShader != nil {
		csSpirv, csKey := t.GetShaderWithContext(bind.ComputeShader, bind.ComputeContext)
		t.activeComputeShader = csSpirv
		t.activeComputeShaderKey = csKey
		shaders = append(shaders, csSpirv)
	}

	// Bind resources.
	staticSetToBind := t.staticDescriptorPool.DefaultSet(frame)
	storeTargets, storeBufferTargets, activeStaticSet, err := t.GetBindDescriptorSet(shaders, userData)
	if err != nil {
		return
	}
	if activeStaticSet != vk.NullDescriptorSet {
		staticSetToBind = activeStaticSet
	}
	t.activeComputeStoreTargets = storeTargets
	t.activeComputeStoreBufferTargets = storeBufferTargets

	vk.CmdBindDescriptorSets(
		t.commandBuffer.CommandBuffer, vk.PipelineBindPointGraphics,
		t.pipelineLayout, spirvStructs.DescriptorSetSlotStatic,
		1, []vk.DescriptorSet{staticSetToBind},
		0, nil,
	)
	vk.CmdBindDescriptorSets(
		t.commandBuffer.CommandBuffer, vk.PipelineBindPointCompute,
		t.pipelineLayout, spirvStructs.DescriptorSetSlotStatic,
		1, []vk.DescriptorSet{staticSetToBind},
		0, nil,
	)
}

// GetBindDescriptorSet resolves compute operands, prepares VkImages and returns a descriptor set to bind.
func (t *GpuTranslator) GetBindDescriptorSet(shaders []*spirv.SpirvShader, userData spirvStructs.UserData) ([]*vulkan.VulkanImage, []*vulkan.VulkanImage, vk.DescriptorSet, error) {
	// Resolve resources accessed by the shaders.
	var imageAccesses []ResolvedImageAccess
	var bufferAccesses []ResolvedBufferAccess
	var allLayouts []spirvCommon.ShaderResourceBinding
	for _, shader := range shaders {
		shaderImageAccesses := ResolveImageResources(shader, userData[:])
		shaderBufferAccesses := ResolveBufferResources(shader, userData[:])
		imageAccesses = append(imageAccesses, shaderImageAccesses...)
		bufferAccesses = append(bufferAccesses, shaderBufferAccesses...)
		allLayouts = append(allLayouts, shader.StaticLayout...)
	}
	if len(imageAccesses) == 0 && len(bufferAccesses) == 0 {
		return nil, nil, vk.NullDescriptorSet, nil
	}

	// Get a static descriptor set.
	activeStaticSet, err := t.staticDescriptorPool.Get(t.handles, t.currentGuestFrame)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, err
	}

	// Update address translation buffer.
	vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          activeStaticSet,
		DstBinding:      spirvStructs.StaticBindingAddressTranslation,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  vk.DescriptorTypeStorageBuffer,
		PBufferInfo: []vk.DescriptorBufferInfo{{
			Buffer: t.addressTranslationBuffer,
			Offset: 0,
			Range:  vk.DeviceSize(vk.WholeSize),
		}},
	}}, 0, nil)

	// Bind images to slots.
	accessByOffset := make(map[uintptr]ResolvedImageAccess, len(imageAccesses))
	for _, access := range imageAccesses {
		accessByOffset[access.InstructionOffset] = access
	}
	boundText := fmt.Sprintf("[Frame %d] Bound slot", t.currentGuestFrame)
	for _, binding := range allLayouts {
		access := accessByOffset[binding.InstructionOffset]

		// Get image view and sampler.
		view, err, _ := t.GetImageView(access.Descriptor)
		if err != nil {
			return nil, nil, vk.NullDescriptorSet, err
		}
		sampler := t.defaultSampler
		if access.Kind == spirvCommon.ImageAccessSample {
			sampler, err = t.GetSampler(*access.Sampler)
			if err != nil {
				return nil, nil, vk.NullDescriptorSet, err
			}
		}

		// Transition image layout, update descriptor set.
		switch binding.Access {
		case spirvCommon.BindingAccessSampledRead:
			view.Image.BarrierSampledRead(t.commandBuffer)
			t.updateStaticDescriptorBinding(activeStaticSet, binding.BindingIndex, view.ImageView, vk.NullImageView, sampler, vk.NullBufferView)
		case spirvCommon.BindingAccessStorageWrite:
			view.Image.BarrierComputeStorageWrite(t.commandBuffer)
			t.updateStaticDescriptorBinding(activeStaticSet, binding.BindingIndex, vk.NullImageView, view.StorageImageView, sampler, vk.NullBufferView)
		}
		boundText += fmt.Sprintf(" %s %d (0x%X/%dx%d)", binding.Access, binding.BindingIndex, access.Descriptor.BaseAddress, access.Descriptor.Width, access.Descriptor.Height)
	}
	boundText += ".\n"
	if len(allLayouts) > 0 && logger.LogRenderer {
		logger.Print(boundText)
	}

	accessText := fmt.Sprintf("[Frame %d] Accessed buffers", t.currentGuestFrame)
	for i, access := range bufferAccesses {
		data := unsafe.Slice((*uint32)(unsafe.Pointer(access.Descriptor.BaseAddress)), 8)
		accessText += fmt.Sprintf(" %d (0x%X/%d) + %+v", i, access.Descriptor.BaseAddress, access.Descriptor.NumRecords, data)
	}
	accessText += ".\n"
	if len(bufferAccesses) > 0 && logger.LogRenderer {
		logger.Print(accessText)
	}

	// Download accessed buffer images.
	loadBufferTargets, err := t.ResolveBufferTargets(bufferAccesses, spirvCommon.BufferAccessLoad)
	for _, image := range loadBufferTargets {
		if !image.ShouldDownloadFromVkImage() {
			continue
		}
		if err = image.DownloadFromVkImage(t.handles, t.commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
			return nil, nil, vk.NullDescriptorSet, err
		}
	}

	// Find out which resources were written to.
	storeTargets, err := t.ResolveImageTargets(imageAccesses, spirvCommon.ImageAccessStore)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, err
	}
	storeBufferTargets, err := t.ResolveBufferTargets(bufferAccesses, spirvCommon.BufferAccessStore)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, err
	}

	return storeTargets, storeBufferTargets, activeStaticSet, nil
}

func (t *GpuTranslator) updateStaticDescriptorBinding(set vk.DescriptorSet, index uint32, sampledView, storageView vk.ImageView, sampler vk.Sampler, bufferView vk.BufferView) {
	if sampledView != vk.NullImageView {
		vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          set,
			DstBinding:      spirvStructs.StaticBindingSampledImages,
			DstArrayElement: index,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo: []vk.DescriptorImageInfo{{
				Sampler:     sampler,
				ImageView:   sampledView,
				ImageLayout: vk.ImageLayoutGeneral,
			}},
		}}, 0, nil)
	}
	if storageView != vk.NullImageView {
		vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          set,
			DstBinding:      spirvStructs.StaticBindingStorageImages,
			DstArrayElement: index,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeStorageImage,
			PImageInfo: []vk.DescriptorImageInfo{{
				ImageView:   storageView,
				ImageLayout: vk.ImageLayoutGeneral,
			}},
		}}, 0, nil)
	}
	if bufferView != vk.NullBufferView {
		vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
			SType:            vk.StructureTypeWriteDescriptorSet,
			DstSet:           set,
			DstBinding:       spirvStructs.StaticBindingSampledBuffers,
			DstArrayElement:  index,
			DescriptorCount:  1,
			DescriptorType:   vk.DescriptorTypeUniformTexelBuffer,
			PTexelBufferView: []vk.BufferView{bufferView},
		}}, 0, nil)
	}
}
