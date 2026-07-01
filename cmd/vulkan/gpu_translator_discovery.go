package vulkan

import (
	"encoding/binary"
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv/gcn"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) CreateDiscoveryBuffers() error {
	// Setup discovery map buffer.
	var err error
	t.discoveryMapBuffer, t.discoveryMapMem, err = t.AllocBuffer(gcn.DiscoveryMapBufferSize,
		vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	// Setup missing resource buffer.
	t.missingResourceBuffer, t.missingResourceMem, err = t.AllocBuffer(gcn.MissingResourceBufferSize,
		vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	// Zero out the buffers.
	mapData := t.handles.MapMemory(t.discoveryMapMem, gcn.DiscoveryMapBufferSize)
	for i := range mapData {
		mapData[i] = 0
	}
	vk.UnmapMemory(t.handles.Device, t.discoveryMapMem)

	reportData := t.handles.MapMemory(t.missingResourceMem, gcn.MissingResourceBufferSize)
	for i := range reportData {
		reportData[i] = 0
	}
	vk.UnmapMemory(t.handles.Device, t.missingResourceMem)

	return nil
}

func (t *GpuTranslator) FulfillResources(frame uint64) uint32 {
	// Read missing resource buffer.
	missingData := t.handles.MapMemory(t.missingResourceMem, gcn.MissingResourceBufferSize)
	defer vk.UnmapMemory(t.handles.Device, t.missingResourceMem)
	count := binary.LittleEndian.Uint32(missingData[0:4])
	if count == 0 {
		return count
	}

	discoveryMapData := t.handles.MapMemory(t.discoveryMapMem, gcn.DiscoveryMapBufferSize)
	defer vk.UnmapMemory(t.handles.Device, t.discoveryMapMem)
	for i := range count {
		offset := gcn.MissingResourceBufferHeader + i*gcn.MissingResourceBufferEntrySize
		descriptorData := missingData[offset : offset+gcn.MissingResourceBufferEntrySize]

		// Parse descriptors.
		dwords := make([]uint32, 12)
		for j := range 12 {
			dwords[j] = binary.LittleEndian.Uint32(descriptorData[j*4 : j*4+4])
		}
		imageDescriptor := spirvStructs.NewImageDescriptor(dwords[0:8])
		samplerDescriptor := spirvStructs.NewSamplerDescriptor(dwords[8:12])

		// Calculate hash index (must match SPIR-V hash).
		hash := uint32(0)
		for j := range 12 {
			hash ^= dwords[j]
		}
		hashIndex := hash & 0xFFFF

		// Check if discovered.
		var samplerKey spirvStructs.ImageSamplerKey
		copy(samplerKey[:], dwords)
		if _, ok := t.discoveryImageSamplerMap[samplerKey]; ok {
			continue
		}
		logger.Printf("[%s] new resource found (hashIndex=%s, imageDescriptor=%s, samplerDescriptor=%s).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Blue.Sprintf("0x%X", hashIndex),
			color.Blue.Sprintf("%+v", imageDescriptor),
			color.Blue.Sprintf("%+v", samplerDescriptor),
		)

		// Alias if it was first discovered without sampler.
		var noSamplerKey spirvStructs.ImageNoSamplerKey
		copy(noSamplerKey[:], dwords[0:8])
		if vkIndex, ok := t.discoveryImageNoSamplerMap[noSamplerKey]; ok {
			t.discoveryImageSamplerMap[samplerKey] = vkIndex
			binary.LittleEndian.PutUint32(discoveryMapData[hashIndex*4:hashIndex*4+4], vkIndex)
			logger.Printf("[%s] aliased resource from %s to existing vulkanIndex=%s.\n",
				color.Blue.Sprintf("Frame %d", frame),
				color.Blue.Sprintf("0x%X", imageDescriptor.BaseAddress),
				color.Green.Sprint(vkIndex),
			)

			// Add the new sampler.
			if samplerDescriptor == nil {
				panic("no sampler to alias")
			}
			view, storageView, err, _ := t.GetImageView(imageDescriptor)
			if err != nil {
				panic(fmt.Errorf("failed to create image view resource: %w", err))
			}
			sampler, err := t.GetSampler(*samplerDescriptor)
			if err != nil {
				panic(fmt.Errorf("failed to create sampler resource: %w", err))
			}
			t.updateBindlessDescriptorSet(vkIndex, view, storageView, sampler)
			continue
		}

		// Create host resources.
		view, storageView, err, _ := t.GetImageView(imageDescriptor)
		if err != nil {
			panic(fmt.Errorf("failed to create image view resource: %w", err))
		}
		sampler := t.defaultSampler
		if samplerDescriptor != nil {
			sampler, err = t.GetSampler(*samplerDescriptor)
			if err != nil {
				panic(fmt.Errorf("failed to create sampler resource: %w", err))
			}
		}

		// Update our maps.
		vkIndex := t.discoveryNextVulkanIndex
		t.discoveryNextVulkanIndex++
		t.discoveryImageNoSamplerMap[noSamplerKey] = vkIndex
		t.discoveryImageSamplerMap[samplerKey] = vkIndex
		logger.Printf("[%s] fulfilled resource from %s (vulkanIndex=%s).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Blue.Sprintf("0x%X", imageDescriptor.BaseAddress),
			color.Green.Sprint(vkIndex),
		)

		// Update discovery map.
		binary.LittleEndian.PutUint32(discoveryMapData[hashIndex*4:hashIndex*4+4], vkIndex)

		// Update bindless set.
		t.updateBindlessDescriptorSet(vkIndex, view, storageView, sampler)
	}

	// Reset counter for next time.
	binary.LittleEndian.PutUint32(missingData[0:4], 0)

	return count
}

func (t *GpuTranslator) createDummyTexture() {
	surface, err := t.GetSurface(SurfaceRequest{
		SurfaceKey: SurfaceKey{
			GpuAddress: spirvStructs.GetPhysicalGpuAddress(0),
		},
		Width:  1,
		Height: 1,
		Format: vk.FormatR8g8b8a8Unorm,
	})
	if err != nil {
		fmt.Printf("failed to create dummy texture: %v\n", err)
		return
	}
	t.defaultSampler = surface.Value.sampler

	// Create dummy texture data (magenta).
	dummyData := []uint8{255, 0, 255, 255}
	t.uploadDataToImage(dummyData, surface.Value.image, 1, 1, 4)

	// Initialize the whole bindless array to a valid dummy descriptor so unresolved indices sample it.
	sampledInfos := make([]vk.DescriptorImageInfo, BindlessTextureCapacity)
	storageInfos := make([]vk.DescriptorImageInfo, BindlessTextureCapacity)
	for i := range sampledInfos {
		sampledInfos[i] = vk.DescriptorImageInfo{
			Sampler:     surface.Value.sampler,
			ImageView:   surface.Value.imageView,
			ImageLayout: vk.ImageLayoutGeneral,
		}
		storageInfos[i] = vk.DescriptorImageInfo{
			ImageView:   surface.Value.imageView,
			ImageLayout: vk.ImageLayoutGeneral,
		}
	}
	vk.UpdateDescriptorSets(t.handles.Device, 2, []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.bindlessDescriptorSet,
			DstBinding:      spirvStructs.BindlessBindingSampledImages,
			DstArrayElement: 0,
			DescriptorCount: BindlessTextureCapacity,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      sampledInfos,
		},
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.bindlessDescriptorSet,
			DstBinding:      spirvStructs.BindlessBindingStorageImages,
			DstArrayElement: 0,
			DescriptorCount: BindlessTextureCapacity,
			DescriptorType:  vk.DescriptorTypeStorageImage,
			PImageInfo:      storageInfos,
		},
	}, 0, nil)
}

func (t *GpuTranslator) updateBindlessDescriptorSet(index uint32, view vk.ImageView, storageView vk.ImageView, sampler vk.Sampler) {
	vk.UpdateDescriptorSets(t.handles.Device, 2, []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.bindlessDescriptorSet,
			DstBinding:      spirvStructs.BindlessBindingSampledImages,
			DstArrayElement: index,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo: []vk.DescriptorImageInfo{{
				Sampler:     sampler,
				ImageView:   view,
				ImageLayout: vk.ImageLayoutGeneral,
			}},
		},
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.bindlessDescriptorSet,
			DstBinding:      spirvStructs.BindlessBindingStorageImages,
			DstArrayElement: index,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeStorageImage,
			PImageInfo: []vk.DescriptorImageInfo{{
				ImageView:   storageView,
				ImageLayout: vk.ImageLayoutGeneral,
			}},
		},
	}, 0, nil)
}
