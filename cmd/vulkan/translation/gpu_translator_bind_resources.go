package translation

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvCommon "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

// BindResources resolves compute operands, prepares VkImages, updates bindless and/or set 2.
func (t *GpuTranslator) BindResources(shaders []*spirv.SpirvShader, userData spirvStructs.UserData) ([]*vulkan.VulkanImage, vk.DescriptorSet, error) {
	// Resolve resources accessed by the shaders.
	var allAccesses []ResolvedImageAccess
	var allLayouts []spirvCommon.ShaderResourceBinding
	for _, shader := range shaders {
		resources := spirv.StaticResources(shader.StaticLayout)
		accesses := ResolveImageResources(resources, shader, userData[:])
		allAccesses = append(allAccesses, accesses...)
		allLayouts = append(allLayouts, shader.StaticLayout...)
	}
	if len(allAccesses) == 0 {
		return nil, vk.NullDescriptorSet, nil
	}
	accessByOffset := make(map[uintptr]ResolvedImageAccess, len(allAccesses))
	for _, access := range allAccesses {
		accessByOffset[access.InstructionOffset] = access
	}

	// Allocate a static set, bind resources to slots.
	activeStaticSet := t.allocateStaticDescriptorSet()
	boundText := fmt.Sprintf("[Frame %d] Bound slot", t.currentGuestFrame)
	for _, binding := range allLayouts {
		access := accessByOffset[binding.InstructionOffset]

		// Get image view and sampler.
		view, err, _ := t.GetImageView(access.Descriptor)
		if err != nil {
			return nil, vk.NullDescriptorSet, err
		}
		sampler := t.defaultSampler
		if access.Kind == spirvCommon.ImageAccessSample {
			sampler, err = t.GetSampler(*access.Sampler)
			if err != nil {
				return nil, vk.NullDescriptorSet, err
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

	// Find out which resources were written to.
	storeTargets, err := t.StoreTargets(allAccesses)
	if err != nil {
		return nil, vk.NullDescriptorSet, err
	}

	// Bind buffer resources if a fetch shader is present.
	if t.activeVertexShader != nil && t.activeVertexShader.GcnShader != nil {
		fetchPC := GetFetchShaderPC(t.activeVertexShader.GcnShader, userData[:])
		if fetchPC != 0 {
			fetchShader := gpu.GlobalLiverpool.GetShader(gcn.GcnShaderStageVertex, fetchPC)
			if fetchShader != nil {
				bufferAccesses := ResolveBufferResources(fetchShader, userData[:])
				for i, access := range bufferAccesses {
					view, err := t.GetBufferView(access.Descriptor)
					if err != nil {
						return nil, vk.NullDescriptorSet, err
					}
					t.updateStaticDescriptorBinding(activeStaticSet, uint32(i), vk.NullImageView, vk.NullImageView, vk.NullSampler, view)
				}
			}
		}
	}

	return storeTargets, activeStaticSet, nil
}

func (t *GpuTranslator) allocateStaticDescriptorSet() vk.DescriptorSet {
	if t.staticDescriptorSetIdx >= len(t.staticDescriptorSets) {
		var newSet vk.DescriptorSet
		result := vk.AllocateDescriptorSets(t.handles.Device, &vk.DescriptorSetAllocateInfo{
			SType:              vk.StructureTypeDescriptorSetAllocateInfo,
			DescriptorPool:     t.descriptorPool,
			DescriptorSetCount: 1,
			PSetLayouts:        []vk.DescriptorSetLayout{t.staticDescriptorSetLayout},
		}, &newSet)
		if result != vk.Success {
			fmt.Printf("WARNING: out of static descriptor sets (%v), falling back to shared set\n", result)
			return t.staticDescriptorSet // fallback to the shared one if out of memory
		}
		t.staticDescriptorSets = append(t.staticDescriptorSets, newSet)
	}
	set := t.staticDescriptorSets[t.staticDescriptorSetIdx]
	t.staticDescriptorSetIdx++

	vk.UpdateDescriptorSets(t.handles.Device, 0, nil, 2, []vk.CopyDescriptorSet{
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcSet:          t.staticDescriptorSet,
			SrcBinding:      spirvStructs.StaticBindingSampledImages,
			SrcArrayElement: 0,
			DstSet:          set,
			DstBinding:      spirvStructs.StaticBindingSampledImages,
			DstArrayElement: 0,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcSet:          t.staticDescriptorSet,
			SrcBinding:      spirvStructs.StaticBindingStorageImages,
			SrcArrayElement: 0,
			DstSet:          set,
			DstBinding:      spirvStructs.StaticBindingStorageImages,
			DstArrayElement: 0,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcSet:          t.staticDescriptorSet,
			SrcBinding:      spirvStructs.StaticBindingSampledBuffers,
			SrcArrayElement: 0,
			DstSet:          set,
			DstBinding:      spirvStructs.StaticBindingSampledBuffers,
			DstArrayElement: 0,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
	})

	return set
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
