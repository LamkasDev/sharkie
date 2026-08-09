package translation

import (
	"fmt"
	"syscall"
	"unsafe"

	gcn2 "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) createDummyTextures() {
	createDummy := func(addr uintptr, imgType uint32) *vulkan.VulkanSurface {
		descriptor := spirvStructs.ImageDescriptor{
			BaseAddress: addr,
			Width:       1, Height: 1,
			DataFormat: gcn2.GcnDataFormat8_8_8_8, NumFormat: gcn2.GcnNumFormatUnorm,
			DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
			Depth: 1, Pitch: 1,
			Type: uint8(imgType),
		}
		size := vulkan.DescriptorGuestSize(descriptor)
		data, err := syscall.Mmap(-1, 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			fmt.Printf("failed to map dummy texture memory: %v\n", err)
			return nil
		}
		descriptor.BaseAddress = uintptr(unsafe.Pointer(&data[0]))
		*(*uint32)(unsafe.Pointer(descriptor.BaseAddress)) = 0xFFFF00FF // magenta

		surface, err := t.GetSurface(descriptor, vk.FormatR8g8b8a8Unorm)
		if err != nil {
			fmt.Printf("failed to create dummy texture: %v\n", err)
			return nil
		}
		return surface
	}

	surface1D := createDummy(1, gcn2.GcnImageTypeColor1D)
	surface2D := createDummy(0, gcn2.GcnImageTypeColor2D)
	surface3D := createDummy(2, gcn2.GcnImageTypeColor3D)
	surface2DArray := createDummy(3, gcn2.GcnImageTypeCubeOrArray)
	if surface2D == nil || surface1D == nil || surface3D == nil || surface2DArray == nil {
		return
	}
	t.defaultSampler = surface2D.Sampler

	t.initDescriptorSets(
		surface2D.ImageView.ImageView,
		surface1D.ImageView.ImageView,
		surface3D.ImageView.ImageView,
		surface2DArray.ImageView.ImageView,
		surface2D.Sampler,
	)
}

func (t *GpuTranslator) initDescriptorSets(view2D, view1D, view3D, view2DArray vk.ImageView, sampler vk.Sampler) {
	sampledInfos1D := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	sampledInfos2D := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	storageInfos2D := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	storageInfos1D := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	storageInfos3D := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	storageInfos2DArray := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	sampledInfos3D := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	sampledInfos2DArray := make([]vk.DescriptorImageInfo, spirvStructs.MaxStaticBindings)
	for i := range sampledInfos2D {
		sampledInfos1D[i] = vk.DescriptorImageInfo{Sampler: sampler, ImageView: view1D, ImageLayout: vk.ImageLayoutGeneral}
		sampledInfos2D[i] = vk.DescriptorImageInfo{Sampler: sampler, ImageView: view2D, ImageLayout: vk.ImageLayoutGeneral}
		storageInfos2D[i] = vk.DescriptorImageInfo{ImageView: view2D, ImageLayout: vk.ImageLayoutGeneral}
		storageInfos1D[i] = vk.DescriptorImageInfo{ImageView: view1D, ImageLayout: vk.ImageLayoutGeneral}
		storageInfos3D[i] = vk.DescriptorImageInfo{ImageView: view3D, ImageLayout: vk.ImageLayoutGeneral}
		storageInfos2DArray[i] = vk.DescriptorImageInfo{ImageView: view2DArray, ImageLayout: vk.ImageLayoutGeneral}
		sampledInfos3D[i] = vk.DescriptorImageInfo{Sampler: sampler, ImageView: view3D, ImageLayout: vk.ImageLayoutGeneral}
		sampledInfos2DArray[i] = vk.DescriptorImageInfo{Sampler: sampler, ImageView: view2DArray, ImageLayout: vk.ImageLayoutGeneral}
	}

	for _, pool := range t.globalDescriptorPool.Pools {
		vk.UpdateDescriptorSets(t.handles.Device, 1, []vk.WriteDescriptorSet{
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.GlobalBindingAddressTranslation,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeStorageBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{{
					Buffer: t.addressTranslationBuffer,
					Offset: 0,
					Range:  vk.DeviceSize(vk.WholeSize),
				}},
			},
		}, 0, nil)
	}

	for _, pool := range t.imageDescriptorPool.Pools {
		writes := []vk.WriteDescriptorSet{
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingSampledImages1D,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo:      sampledInfos1D,
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingSampledImages2D,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo:      sampledInfos2D,
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingStorageImages2D,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeStorageImage,
				PImageInfo:      storageInfos2D,
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingStorageImages1D,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeStorageImage,
				PImageInfo:      storageInfos1D,
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingStorageImages3D,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeStorageImage,
				PImageInfo:      storageInfos3D,
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingStorageImages2DArray,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeStorageImage,
				PImageInfo:      storageInfos2DArray,
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingSampledImages3D,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo:      sampledInfos3D,
			},
			{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstSet:          pool.DefaultSet,
				DstBinding:      spirvStructs.ImageBindingSampledImages2DArray,
				DstArrayElement: 0,
				DescriptorCount: spirvStructs.MaxStaticBindings,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo:      sampledInfos2DArray,
			},
		}
		vk.UpdateDescriptorSets(t.handles.Device, uint32(len(writes)), writes, 0, nil)
	}
}
