package vulkan

import (
	gcn2 "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
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
	components := vk.ComponentMapping{
		R: gcn.TranslateDstSelToVkSwizzle(request.Descriptor.DstSelX),
		G: gcn.TranslateDstSelToVkSwizzle(request.Descriptor.DstSelY),
		B: gcn.TranslateDstSelToVkSwizzle(request.Descriptor.DstSelZ),
		A: gcn.TranslateDstSelToVkSwizzle(request.Descriptor.DstSelW),
	}

	switch request.Descriptor.DataFormat {
	case gcn2.GcnDataFormat5_6_5, gcn2.GcnDataFormat1_5_5_5, gcn2.GcnDataFormat11_11_10:
		components = vk.ComponentMapping{
			R: components.B,
			G: components.G,
			B: components.R,
			A: components.A,
		}
	case gcn2.GcnDataFormat10_10_10_2:
		components = vk.ComponentMapping{
			R: components.A,
			G: components.B,
			B: components.G,
			A: components.R,
		}
	case gcn2.GcnDataFormat4_4_4_4:
		components = vk.ComponentMapping{
			R: components.G,
			G: components.B,
			B: components.A,
			A: components.R,
		}
	case gcn2.GcnDataFormat8, gcn2.GcnDataFormat8_8, gcn2.GcnDataFormat16_16:
		if components.R == vk.ComponentSwizzleA {
			components.R = vk.ComponentSwizzleR
		}
		if components.G == vk.ComponentSwizzleA {
			components.G = vk.ComponentSwizzleR
		}
		if components.B == vk.ComponentSwizzleA {
			components.B = vk.ComponentSwizzleR
		}
		if components.A == vk.ComponentSwizzleA {
			components.A = vk.ComponentSwizzleR
		}
	}
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

	var view vk.ImageView
	result := vk.CreateImageView(handles.Device, &vk.ImageViewCreateInfo{
		SType:      vk.StructureTypeImageViewCreateInfo,
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
