package vulkan

import (
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
	isBlock := request.Descriptor.DataFormat >= 35 && request.Descriptor.DataFormat <= 41
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
	if storageLayout {
		components = vk.ComponentMapping{
			R: vk.ComponentSwizzleIdentity,
			G: vk.ComponentSwizzleIdentity,
			B: vk.ComponentSwizzleIdentity,
			A: vk.ComponentSwizzleIdentity,
		}
	}

	viewType := vk.ImageViewType2d
	switch request.Descriptor.Type {
	case 8: // Color1D
		viewType = vk.ImageViewType1d
	case 10: // Color3D
		viewType = vk.ImageViewType3d
	case 11: // Cube
		viewType = vk.ImageViewTypeCube
	case 12: // Color1DArray
		viewType = vk.ImageViewType1dArray
	case 13, 15: // Color2DArray, Color2DMsaaArray
		viewType = vk.ImageViewType2dArray
	}

	var view vk.ImageView
	result := vk.CreateImageView(handles.Device, &vk.ImageViewCreateInfo{
		SType:      vk.StructureTypeImageViewCreateInfo,
		Image:      request.Image.Image,
		ViewType:   viewType,
		Format:     request.Image.ImageFormat,
		Components: components,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:   request.Image.ImageAspect,
			BaseMipLevel: min(uint32(request.Descriptor.BaseLevel), 15),
			LevelCount:   1,
			LayerCount:   1,
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
