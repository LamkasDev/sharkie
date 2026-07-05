package translation

import (
	"fmt"
	"syscall"
	"unsafe"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) createDummyTexture() {
	// Setup space for dummy texture.
	data, err := syscall.Mmap(-1, 0, int(structs.SystemPageSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		fmt.Printf("failed to map dummy texture memory: %v\n", err)
		return
	}
	baseAddress := uintptr(unsafe.Pointer(&data[0]))
	structs.GlobalMemoryManager.Guest().MapAnonymous(baseAddress, uint64(structs.SystemPageSize), structs.PROT_READ|structs.PROT_WRITE|structs.PROT_GPU_READ, structs.VMATypeAnon)
	structs.GlobalMemoryManager.OnMapGuest(baseAddress, structs.SystemPageSize)
	*(*uint32)(unsafe.Pointer(baseAddress)) = 0xFFFF00FF // magenta

	surface, err := t.GetSurface(spirvStructs.ImageDescriptor{
		BaseAddress: baseAddress,
		Width:       1, Height: 1,
		DataFormat: 10, NumFormat: 0,
		DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
		Depth: 1, Pitch: 1,
	}, vk.FormatR8g8b8a8Unorm)
	if err != nil {
		fmt.Printf("failed to create dummy texture: %v\n", err)
		return
	}
	t.defaultSampler = surface.Sampler

	structs.GlobalMemoryManager.MarkRegionCpuModified(baseAddress, 4)
	surface.ImageView.Image.UploadToVkImage(&t.handles, t.GetLinearBuffer)

	t.initStaticDescriptorSet(surface.ImageView.ImageView, surface.Sampler)
}

func (t *GpuTranslator) initStaticDescriptorSet(imageView vk.ImageView, sampler vk.Sampler) {
	if t.staticDescriptorSet == vk.NullDescriptorSet || imageView == vk.NullImageView {
		return
	}
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
			DstSet:          t.staticDescriptorSet,
			DstBinding:      spirvStructs.StaticBindingSampledImages,
			DstArrayElement: 0,
			DescriptorCount: spirvStructs.MaxStaticBindings,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      sampledInfos,
		},
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          t.staticDescriptorSet,
			DstBinding:      spirvStructs.StaticBindingStorageImages,
			DstArrayElement: 0,
			DescriptorCount: spirvStructs.MaxStaticBindings,
			DescriptorType:  vk.DescriptorTypeStorageImage,
			PImageInfo:      storageInfos,
		},
	}, 0, nil)
}
