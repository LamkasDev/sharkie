package renderer

import (
	"encoding/binary"
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) FulfillResources(frame uint64) {
	// Read MissingResourceBuffer.
	data := t.handles.MapMemory(t.discoveryReportMem, 4+1024*48)
	count := binary.LittleEndian.Uint32(data[0:4])
	if count == 0 {
		vk.UnmapMemory(t.handles.Device, t.discoveryReportMem)
		return
	}
	if count > 1024 {
		count = 1024
	}

	mapData := t.handles.MapMemory(t.discoveryMapMem, 65536*4)
	for i := uint32(0); i < count; i++ {
		offset := 4 + i*48
		descriptorData := data[offset : offset+48]

		// Parse descriptors.
		dwords := make([]uint32, 12)
		for j := 0; j < 12; j++ {
			dwords[j] = binary.LittleEndian.Uint32(descriptorData[j*4 : j*4+4])
		}

		var key [12]uint32
		copy(key[:], dwords)

		if _, ok := t.discoveryMap[key]; ok {
			continue
		}

		// Calculate hash index (must match SPIR-V hash).
		hash := uint32(0)
		for j := 0; j < 12; j++ {
			hash ^= dwords[j]
		}
		hashIndex := hash & 0xFFFF

		// Create resource.
		imageDescriptor := spirv.NewImageDescriptor(dwords[0:8])
		samplerDescriptor := spirv.NewSamplerDescriptor(dwords[8:12])
		logger.Printf("[%s] new resource found (hashIndex=%s, imageDescriptor=%s, samplerDescriptor=%s).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Blue.Sprintf("0x%X", hashIndex),
			color.Blue.Sprintf("0x%X", imageDescriptor),
			color.Blue.Sprintf("0x%X", samplerDescriptor),
		)

		view := t.GetImageView(imageDescriptor)
		sampler := t.GetSampler(samplerDescriptor)
		if view != vk.NullImageView && sampler != vk.NullSampler {
			vulkanIndex := t.discoveryNextIndex
			t.discoveryNextIndex++
			t.discoveryMap[key] = vulkanIndex

			// Update map on GPU.
			binary.LittleEndian.PutUint32(mapData[hashIndex*4:hashIndex*4+4], vulkanIndex)

			// Update bindless set.
			t.updateBindlessDescriptorSet(vulkanIndex, view, sampler)
		}
	}

	// Reset counter for next time.
	binary.LittleEndian.PutUint32(data[0:4], 0)
	vk.UnmapMemory(t.handles.Device, t.discoveryMapMem)
	vk.UnmapMemory(t.handles.Device, t.discoveryReportMem)
}

func (t *GpuTranslator) createDummyTexture() {
	surface := &GpuSurface{
		GPUAddress: 0,
		Width:      1,
		Height:     1,
		Format:     vk.FormatR8g8b8a8Unorm,
		firstUse:   true,
	}
	if err := t.allocSurface(surface); err != nil {
		fmt.Printf("Discovery: Failed to create dummy texture: %v\n", err)
		return
	}
	t.updateBindlessDescriptorSet(0, surface.imageView, surface.sampler)
	fmt.Println("Discovery: Created dummy texture at index 0.")
}

func (t *GpuTranslator) updateBindlessDescriptorSet(index uint32, view vk.ImageView, sampler vk.Sampler) {
	vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          t.bindlessDescriptorSet,
		DstBinding:      0,
		DstArrayElement: index,
		DescriptorCount: 1,
		DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
		PImageInfo: []vk.DescriptorImageInfo{{
			Sampler:     sampler,
			ImageView:   view,
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
		}},
	}}, 0, nil)
}
