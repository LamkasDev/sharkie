package vulkan

import (
	"encoding/binary"
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) CreateDiscoveryBuffers() error {
	// Setup discovery map buffer.
	var err error
	t.discoveryMapBuffer, t.discoveryMapMem, err = t.AllocBuffer(spirv.DiscoveryMapBufferSize,
		vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	// Setup missing resource buffer.
	t.missingResourceBuffer, t.missingResourceMem, err = t.AllocBuffer(spirv.MissingResourceBufferSize,
		vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	// Zero out the buffers.
	mapData := t.handles.MapMemory(t.discoveryMapMem, spirv.DiscoveryMapBufferSize)
	for i := range mapData {
		mapData[i] = 0
	}
	vk.UnmapMemory(t.handles.Device, t.discoveryMapMem)

	reportData := t.handles.MapMemory(t.missingResourceMem, spirv.MissingResourceBufferSize)
	for i := range reportData {
		reportData[i] = 0
	}
	vk.UnmapMemory(t.handles.Device, t.missingResourceMem)

	return nil
}

func (t *GpuTranslator) FulfillResources(frame uint64) {
	// Read missing resource buffer.
	missingData := t.handles.MapMemory(t.missingResourceMem, spirv.MissingResourceBufferSize)
	defer vk.UnmapMemory(t.handles.Device, t.missingResourceMem)
	count := binary.LittleEndian.Uint32(missingData[0:4])
	if count == 0 {
		return
	}

	discoveryMapData := t.handles.MapMemory(t.discoveryMapMem, spirv.DiscoveryMapBufferSize)
	defer vk.UnmapMemory(t.handles.Device, t.discoveryMapMem)
	for i := range count {
		offset := spirv.MissingResourceBufferHeader + i*spirv.MissingResourceBufferEntrySize
		descriptorData := missingData[offset : offset+spirv.MissingResourceBufferEntrySize]

		// Parse descriptors.
		dwords := make([]uint32, 12)
		for j := range 12 {
			dwords[j] = binary.LittleEndian.Uint32(descriptorData[j*4 : j*4+4])
		}
		imageDescriptor := spirv.NewImageDescriptor(dwords[0:8])
		samplerDescriptor := spirv.NewSamplerDescriptor(dwords[8:12])

		// Calculate hash index (must match SPIR-V hash).
		hash := uint32(0)
		for j := range 12 {
			hash ^= dwords[j]
		}
		hashIndex := hash & 0xFFFF

		// Check if discovered.
		var key [12]uint32
		copy(key[:], dwords)
		if _, ok := t.discoveryMap[key]; ok {
			continue
		}
		logger.Printf("[%s] new resource found (hashIndex=%s, imageDescriptor=%s, samplerDescriptor=%s).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Blue.Sprintf("0x%X", hashIndex),
			color.Blue.Sprintf("0x%X", imageDescriptor),
			color.Blue.Sprintf("0x%X", samplerDescriptor),
		)

		// Create host resources.
		view := t.GetImageView(imageDescriptor)
		sampler := t.GetSampler(samplerDescriptor)
		if view != vk.NullImageView && sampler != vk.NullSampler {
			vulkanIndex := t.discoveryNextVulkanIndex
			t.discoveryNextVulkanIndex++
			t.discoveryMap[key] = vulkanIndex
			logger.Printf("[%s] fulfilled resource from %s (vulkanIndex=%s).\n",
				color.Blue.Sprintf("Frame %d", frame),
				color.Blue.Sprintf("0x%X", imageDescriptor.BaseAddress),
				color.Green.Sprint(vulkanIndex),
			)

			// Upload image to GPU.
			_, bytesPerPixel := TranslateGcnFormat(imageDescriptor.DataFormat, imageDescriptor.NumFormat)
			t.uploadBufferDataToImage(imageDescriptor.BaseAddress, t.images[imageDescriptor.BaseAddress], imageDescriptor.Width, imageDescriptor.Height, bytesPerPixel)

			// Update discovery map.
			binary.LittleEndian.PutUint32(discoveryMapData[hashIndex*4:hashIndex*4+4], vulkanIndex)

			// Update bindless set.
			t.updateBindlessDescriptorSet(vulkanIndex, view, sampler)
		} else {
			logger.Printf("[%s] failed fulfillment for resource from %s.\n",
				color.Blue.Sprintf("Frame %d", frame),
				color.Blue.Sprintf("0x%X", imageDescriptor.BaseAddress),
			)

			// Reset the map entry so it can be retried.
			binary.LittleEndian.PutUint32(discoveryMapData[hashIndex*4:hashIndex*4+4], 0)
		}
	}

	// Reset counter for next time.
	binary.LittleEndian.PutUint32(missingData[0:4], 0)
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
		fmt.Printf("failed to create dummy texture: %v\n", err)
		return
	}

	// Transition dummy image to ShaderReadOnly so it's valid for sampling.
	cb := t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(cb, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	t.imageBarrier(cb, surface.image,
		vk.ImageLayoutUndefined, vk.ImageLayoutShaderReadOnlyOptimal,
		0, vk.AccessFlags(vk.AccessShaderReadBit),
		vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit))
	vk.EndCommandBuffer(cb)
	vk.QueueSubmit(t.handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cb},
	}}, vk.NullFence)
	vk.QueueWaitIdle(t.handles.GraphicsQueue)
	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{cb})

	// Create dummy texture data (magenta).
	dummyData := []uint8{255, 0, 255, 255}
	t.uploadDataToImage(dummyData, surface.image, 1, 1, 4)

	// Initialize the whole bindless array to a valid dummy descriptor so unresolved indices sample it.
	infos := make([]vk.DescriptorImageInfo, BindlessTextureCapacity)
	for i := range infos {
		infos[i] = vk.DescriptorImageInfo{
			Sampler:     surface.sampler,
			ImageView:   surface.imageView,
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
		}
	}
	vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          t.bindlessDescriptorSet,
		DstBinding:      0,
		DstArrayElement: 0,
		DescriptorCount: BindlessTextureCapacity,
		DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
		PImageInfo:      infos,
	}}, 0, nil)
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
