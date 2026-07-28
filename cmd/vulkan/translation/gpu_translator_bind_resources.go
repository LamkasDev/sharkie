package translation

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvCommon "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

// BindResources resolves compute operands, prepares VkImages, updates bindless and/or set 2.
func (t *GpuTranslator) BindResources(shaders []*spirv.SpirvShader, userData spirvStructs.UserData) ([]*vulkan.VulkanImage, []*vulkan.VulkanImage, vk.DescriptorSet, error) {
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

	// Get a static descriptor set, bind images to slots.
	activeStaticSet, err := t.staticDescriptorPool.Get(t.handles, t.currentGuestFrame)
	if err != nil {
		return nil, nil, vk.NullDescriptorSet, err
	}
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
			t.EndRenderPass()
			view.Image.BarrierSampledRead(t.commandBuffer)
			t.updateStaticDescriptorBinding(activeStaticSet, binding.BindingIndex, view.ImageView, vk.NullImageView, sampler, vk.NullBufferView)
		case spirvCommon.BindingAccessStorageWrite:
			t.EndRenderPass()
			view.Image.BarrierComputeStorageWrite(t.commandBuffer)
			t.updateStaticDescriptorBinding(activeStaticSet, binding.BindingIndex, vk.NullImageView, view.StorageImageView, sampler, vk.NullBufferView)
		}
		boundText += fmt.Sprintf(" %s %d (0x%X/%dx%d)", binding.Access, binding.BindingIndex, access.Descriptor.BaseAddress, access.Descriptor.Width, access.Descriptor.Height)
	}
	boundText += ".\n"
	if len(allLayouts) > 0 && logger.LogRenderer {
		logger.Print(boundText)
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
