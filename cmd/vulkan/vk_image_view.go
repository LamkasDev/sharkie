package vulkan

import (
	"unsafe"

	gcn2 "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
)

type VulkanImageView struct {
	Descriptor       spirvStructs.ImageDescriptor
	Image            *VulkanImage
	ImageView        vk.ImageView
	StorageImageView vk.ImageView
}

type VulkanImageViewRequest struct {
	Image      *VulkanImage
	Descriptor spirvStructs.ImageDescriptor
}

func CreateImageView(handles *VulkanHandles, request VulkanImageViewRequest) (*VulkanImageView, error) {
	imageView, err := CreateVkImageView(handles, request, false)
	if err != nil {
		return nil, err
	}
	var storageImageView vk.ImageView
	isBlock := request.Descriptor.DataFormat >= gcn2.GcnDataFormatBC1 && request.Descriptor.DataFormat <= gcn2.GcnDataFormatBC7
	if !isBlock {
		storageImageView, err = CreateVkImageView(handles, request, true)
		if err != nil {
			vk.DestroyImageView(handles.Device, imageView, nil)
			return nil, err
		}
	} else {
		storageImageView = vk.NullImageView
	}

	return &VulkanImageView{
		Descriptor:       request.Descriptor,
		Image:            request.Image,
		ImageView:        imageView,
		StorageImageView: storageImageView,
	}, nil
}

func CreateVkImageView(handles *VulkanHandles, request VulkanImageViewRequest, storageLayout bool) (vk.ImageView, error) {
	components := gcn.TranslateComponentMapping(
		request.Descriptor.DataFormat, request.Descriptor.NumFormat,
		request.Descriptor.DstSelX, request.Descriptor.DstSelY, request.Descriptor.DstSelZ, request.Descriptor.DstSelW,
	)
	if storageLayout {
		components = vk.ComponentMapping{
			R: vk.ComponentSwizzleIdentity,
			G: vk.ComponentSwizzleIdentity,
			B: vk.ComponentSwizzleIdentity,
			A: vk.ComponentSwizzleIdentity,
		}
	}

	viewType := vk.ImageViewType2d
	baseArray := uint32(request.Descriptor.BaseArray)
	layerCount := uint32(1)
	switch request.Descriptor.Type {
	case gcn2.GcnImageTypeColor1D:
		viewType = vk.ImageViewType1d
	case gcn2.GcnImageTypeColor3D:
		viewType = vk.ImageViewType3d
		baseArray = 0
	case gcn2.GcnImageTypeCubeOrArray:
		viewType = vk.ImageViewType2dArray
		layerCount = (uint32(request.Descriptor.LastArray) + 1) - baseArray
	case gcn2.GcnImageTypeColor1DArray:
		viewType = vk.ImageViewType1dArray
		layerCount = (uint32(request.Descriptor.LastArray) + 1) - baseArray
	case gcn2.GcnImageTypeColor2DMsaa:
		viewType = vk.ImageViewType2d
	case gcn2.GcnImageTypeColor2DArray, gcn2.GcnImageTypeColor2DMsaaArray:
		viewType = vk.ImageViewType2dArray
		layerCount = (uint32(request.Descriptor.LastArray) + 1) - baseArray
	}

	// Filter unsupported usage bits based on format properties.
	var formatProps vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(handles.PhysicalDevice, request.Image.ImageFormat, &formatProps)
	formatProps.Deref()
	imageUsage := request.Image.ImageUsage
	if (formatProps.OptimalTilingFeatures & vk.FormatFeatureFlags(vk.FormatFeatureStorageImageBit)) == 0 {
		if imageUsage&vk.ImageUsageFlags(vk.ImageUsageStorageBit) != 0 {
			logger.Printf("Failed assigning storage bit to format %d.\n", request.Image.ImageFormat)
		}
		imageUsage &^= vk.ImageUsageFlags(vk.ImageUsageStorageBit)
	}

	usageInfo := vk.ImageViewUsageCreateInfo{
		SType: vk.StructureTypeImageViewUsageCreateInfo,
		Usage: imageUsage,
	}

	var view vk.ImageView
	result := vk.CreateImageView(handles.Device, &vk.ImageViewCreateInfo{
		SType:      vk.StructureTypeImageViewCreateInfo,
		PNext:      unsafe.Pointer(&usageInfo),
		Image:      request.Image.Image,
		ViewType:   viewType,
		Format:     request.Image.ImageFormat,
		Components: components,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     request.Image.ImageAspect,
			BaseMipLevel:   min(uint32(request.Descriptor.BaseLevel), 15),
			LevelCount:     1,
			BaseArrayLayer: baseArray,
			LayerCount:     layerCount,
		},
	}, nil, &view)
	if err := NewError(result); err != nil {
		return vk.NullImageView, err
	}

	return view, nil
}

func (view *VulkanImageView) Destroy(device vk.Device) {
	if view.ImageView != vk.NullImageView {
		vk.DestroyImageView(device, view.ImageView, nil)
		view.ImageView = vk.NullImageView
	}
	if view.StorageImageView != vk.NullImageView {
		vk.DestroyImageView(device, view.StorageImageView, nil)
		view.StorageImageView = vk.NullImageView
	}
}
