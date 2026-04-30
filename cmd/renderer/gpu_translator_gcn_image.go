package renderer

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) BindImageSamplers(commandBuffer vk.CommandBuffer, userData []uint32, stageOffset uint32) {
	// The texture table pointer is in s[2:3]
	textureTablePtr := uintptr(userData[stageOffset+2]) | (uintptr(userData[stageOffset+3]) << 32)
	if textureTablePtr == 0 {
		return
	}

	logger.Printf("[GPU] Texture Table Ptr: 0x%X\n", textureTablePtr)

	// The table contains 32-bit indices.
	// We'll iterate through the first 16 entries and bind any that are non-zero.
	var writes []vk.WriteDescriptorSet
	var imageInfos []vk.DescriptorImageInfo

	for i := range uint32(16) {
		// Load index from table.
		indexPtr := (*uint32)(unsafe.Pointer(textureTablePtr + uintptr(i*4)))
		index := *indexPtr
		if index == 0 {
			continue
		}

		logger.Printf("[GPU] Texture Index %d: 0x%X\n", i, index)

		// The index points to a 256-bit image descriptor followed by a 128-bit sampler descriptor?
		// Or maybe the descriptors are in a separate global table?
		// For now, let's assume the descriptors are in memory at some fixed location.
		// PS4 games often use a global table at a fixed address.
		// Let's assume the index is into a global table of (ImageDesc[8], SamplerDesc[4]).
		globalTableBase := uintptr(0x200000000)         // Placeholder
		descAddr := globalTableBase + uintptr(index)*48 // 32 + 16 bytes

		imageDwords := unsafe.Slice((*uint32)(unsafe.Pointer(descAddr)), 8)
		samplerDwords := unsafe.Slice((*uint32)(unsafe.Pointer(descAddr+32)), 4)

		imageDesc := spirv.NewImageDescriptor(imageDwords)
		samplerDesc := spirv.NewSamplerDescriptor(samplerDwords)

		// Create/Get Vulkan resources.
		view := t.GetImageView(imageDesc)
		sampler := t.GetSampler(samplerDesc)

		if view != vk.NullImageView && sampler != vk.NullSampler {
			imageInfos = append(imageInfos, vk.DescriptorImageInfo{
				Sampler:     sampler,
				ImageView:   view,
				ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			})

			writes = append(writes, vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          t.texelDescriptorSets[0], // Using the first set for now
				DstBinding:      0,
				DstArrayElement: uint32(i),
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo:      []vk.DescriptorImageInfo{imageInfos[len(imageInfos)-1]},
			})
		}
	}

	if len(writes) > 0 {
		// For bindless textures, we typically use a single global descriptor set.
		// However, for simplicity here, we'll just bind the first one we allocated.
		// In a real implementation, we would have a dedicated set for bindless textures.
		descriptorSet := t.texelDescriptorSets[0]
		vk.UpdateDescriptorSets(t.handles.Device, uint32(len(writes)), writes, 0, nil)
		vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, t.stubPipelineLayout, 0, 1, []vk.DescriptorSet{descriptorSet}, 0, nil)
	}
}

func translateClampMode(mode uint8) vk.SamplerAddressMode {
	switch mode {
	case 0: // wrap
		return vk.SamplerAddressModeRepeat
	case 1: // mirror
		return vk.SamplerAddressModeMirroredRepeat
	case 2: // clamp to edge
		return vk.SamplerAddressModeClampToEdge
	case 3: // mirror once
		return vk.SamplerAddressModeMirrorClampToEdge
	case 4: // clamp to border
		return vk.SamplerAddressModeClampToBorder
	default:
		return vk.SamplerAddressModeRepeat
	}
}

func translateFilter(mode uint8) vk.Filter {
	switch mode {
	case 0: // point
		return vk.FilterNearest
	case 1: // linear
		return vk.FilterLinear
	default:
		return vk.FilterNearest
	}
}

func translateMipmapMode(mode uint8) vk.SamplerMipmapMode {
	switch mode {
	case 0: // none
		return vk.SamplerMipmapModeNearest
	case 1: // point
		return vk.SamplerMipmapModeNearest
	case 2: // linear
		return vk.SamplerMipmapModeLinear
	default:
		return vk.SamplerMipmapModeNearest
	}
}

// TranslateGcnFormat maps GCN DataFormat and NumFormat to Vulkan VkFormat and byte size.
func TranslateGcnFormat(dataFormat, numFormat uint8) (vk.Format, uint32) {
	switch dataFormat {
	case 1: // 8-bit (1 byte)
		switch numFormat {
		case 0:
			return vk.FormatR8Unorm, 1
		case 1:
			return vk.FormatR8Snorm, 1
		case 2:
			return vk.FormatR8Uscaled, 1
		case 3:
			return vk.FormatR8Sscaled, 1
		case 4:
			return vk.FormatR8Uint, 1
		case 5:
			return vk.FormatR8Sint, 1
		}
	case 2: // 16-bit (2 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16Unorm, 2
		case 1:
			return vk.FormatR16Snorm, 2
		case 2:
			return vk.FormatR16Uscaled, 2
		case 3:
			return vk.FormatR16Sscaled, 2
		case 4:
			return vk.FormatR16Uint, 2
		case 5:
			return vk.FormatR16Sint, 2
		case 7:
			return vk.FormatR16Sfloat, 2
		}
	case 3: // 8_8 (2 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR8g8Unorm, 2
		case 1:
			return vk.FormatR8g8Snorm, 2
		case 2:
			return vk.FormatR8g8Uscaled, 2
		case 3:
			return vk.FormatR8g8Sscaled, 2
		case 4:
			return vk.FormatR8g8Uint, 2
		case 5:
			return vk.FormatR8g8Sint, 2
		}
	case 4: // 32-bit (4 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32Uint, 4
		case 5:
			return vk.FormatR32Sint, 4
		case 7:
			return vk.FormatR32Sfloat, 4
		}
	case 5: // 16_16 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16g16Unorm, 4
		case 1:
			return vk.FormatR16g16Snorm, 4
		case 2:
			return vk.FormatR16g16Uscaled, 4
		case 3:
			return vk.FormatR16g16Sscaled, 4
		case 4:
			return vk.FormatR16g16Uint, 4
		case 5:
			return vk.FormatR16g16Sint, 4
		case 7:
			return vk.FormatR16g16Sfloat, 4
		}
	case 6: // 10_11_11 (4 bytes)
		if numFormat == 7 {
			return vk.FormatB10g11r11UfloatPack32, 4
		}
	case 8: // 10_10_10_2 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatA2b10g10r10UnormPack32, 4
		case 4:
			return vk.FormatA2b10g10r10UintPack32, 4
		}
	case 10: // 8_8_8_8 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR8g8b8a8Unorm, 4
		case 1:
			return vk.FormatR8g8b8a8Snorm, 4
		case 2:
			return vk.FormatR8g8b8a8Uscaled, 4
		case 3:
			return vk.FormatR8g8b8a8Sscaled, 4
		case 4:
			return vk.FormatR8g8b8a8Uint, 4
		case 5:
			return vk.FormatR8g8b8a8Sint, 4
		}
	case 11: // 32_32 (8 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32Uint, 8
		case 5:
			return vk.FormatR32g32Sint, 8
		case 7:
			return vk.FormatR32g32Sfloat, 8
		}
	case 12: // 16_16_16_16 (8 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16g16b16a16Unorm, 8
		case 1:
			return vk.FormatR16g16b16a16Snorm, 8
		case 2:
			return vk.FormatR16g16b16a16Uscaled, 8
		case 3:
			return vk.FormatR16g16b16a16Sscaled, 8
		case 4:
			return vk.FormatR16g16b16a16Uint, 8
		case 5:
			return vk.FormatR16g16b16a16Sint, 8
		case 7:
			return vk.FormatR16g16b16a16Sfloat, 8
		}
	case 13: // 32_32_32 (12 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32b32Uint, 12
		case 5:
			return vk.FormatR32g32b32Sint, 12
		case 7:
			return vk.FormatR32g32b32Sfloat, 12
		}
	case 14: // 32_32_32_32 (16 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32b32a32Uint, 16
		case 5:
			return vk.FormatR32g32b32a32Sint, 16
		case 7:
			return vk.FormatR32g32b32a32Sfloat, 16
		}
	}

	return vk.FormatUndefined, 0
}
