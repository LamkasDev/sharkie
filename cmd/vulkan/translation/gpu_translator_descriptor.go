package translation

import (
	"fmt"
	"syscall"
	"unsafe"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) createDummyTexture() {
	// Calculate how much space to allocate.
	descriptor := spirvStructs.ImageDescriptor{
		BaseAddress: 0,
		Width:       1, Height: 1,
		DataFormat: 10, NumFormat: 0,
		DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
		Depth: 1, Pitch: 1,
	}
	size := vulkan.DescriptorGuestSize(descriptor)

	// Setup space for dummy texture.
	data, err := syscall.Mmap(-1, 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		fmt.Printf("failed to map dummy texture memory: %v\n", err)
		return
	}
	descriptor.BaseAddress = uintptr(unsafe.Pointer(&data[0]))
	*(*uint32)(unsafe.Pointer(descriptor.BaseAddress)) = 0xFFFF00FF // magenta

	// Register it as surface.
	surface, err := t.GetSurface(descriptor, vk.FormatR8g8b8a8Unorm)
	if err != nil {
		fmt.Printf("failed to create dummy texture: %v\n", err)
		return
	}
	t.defaultSampler = surface.Sampler

	t.initStaticDescriptorSet(surface.ImageView.ImageView, surface.Sampler)
}

func (t *GpuTranslator) initStaticDescriptorSet(imageView vk.ImageView, sampler vk.Sampler) {
	sampledInfos := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	storageInfos := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	for i := range sampledInfos {
		sampledInfos[i] = vk.DescriptorImageInfo{
			Sampler:     sampler,
			ImageView:   imageView,
			ImageLayout: vk.ImageLayoutGeneral,
		}
		storageInfos[i] = vk.DescriptorImageInfo{
			ImageView:   imageView,
			ImageLayout: vk.ImageLayoutGeneral,
		}
	}
	vk.UpdateDescriptorSets(t.handles.Device, 2, []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.staticDescriptorPool.Pools[0].DefaultSet,
			DstBinding:      spirvStructs.StaticBindingSampledImages,
			DstArrayElement: 0,
			DescriptorCount: spirvStructs.MaxStaticBindings,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      sampledInfos,
		},
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.staticDescriptorPool.Pools[0].DefaultSet,
			DstBinding:      spirvStructs.StaticBindingStorageImages,
			DstArrayElement: 0,
			DescriptorCount: spirvStructs.MaxStaticBindings,
			DescriptorType:  vk.DescriptorTypeStorageImage,
			PImageInfo:      storageInfos,
		},
	}, 0, nil)
}
