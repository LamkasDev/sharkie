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
func (t *GpuTranslator) BindResources(shader *spirv.SpirvShader, userData spirvStructs.UserData) ([]*vulkan.VulkanImage, vk.DescriptorSet, error) {
	// Resolve resources accessed by the shader.
	resources := spirv.StaticResources(shader.StaticLayout)
	accesses := ResolveImageResources(resources, shader, userData[:])
	if len(accesses) == 0 {
		return nil, vk.NullDescriptorSet, nil
	}
	accessByOffset := make(map[uintptr]ResolvedImageAccess, len(accesses))
	for _, access := range accesses {
		accessByOffset[access.InstructionOffset] = access
	}

	// Allocate a static set, bind resources to slots.
	activeStaticSet := t.allocateStaticDescriptorSet()
	boundText := fmt.Sprintf("[Frame %d] Bound slot", t.currentGuestFrame)
	for _, binding := range shader.StaticLayout {
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
			t.updateStaticDescriptorBinding(activeStaticSet, binding.BindingIndex, view.ImageView, vk.NullImageView, sampler)
		case spirvCommon.BindingAccessStorageWrite:
			t.EndRenderPass()
			view.Image.BarrierComputeStorageWrite(t.commandBuffer)
			t.updateStaticDescriptorBinding(activeStaticSet, binding.BindingIndex, vk.NullImageView, view.StorageImageView, sampler)
		}
		boundText += fmt.Sprintf(" %s %d (0x%X/%dx%d)", binding.Access, binding.BindingIndex, access.Descriptor.BaseAddress, access.Descriptor.Width, access.Descriptor.Height)
	}
	boundText += ".\n"
	if len(shader.StaticLayout) > 0 {
		logger.Print(boundText)
	}

	// Find out which resources were written to.
	storeTargets, err := t.StoreTargets(accesses)
	if err != nil {
		return nil, vk.NullDescriptorSet, err
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
	})

	return set
}

func (t *GpuTranslator) updateStaticDescriptorBinding(set vk.DescriptorSet, index uint32, sampledView, storageView vk.ImageView, sampler vk.Sampler) {
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
}
