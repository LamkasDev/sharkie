package vulkan

import vk "github.com/goki/vulkan"

func IsDepthFormat(format vk.Format) bool {
	switch format {
	case vk.FormatD16Unorm, vk.FormatD16UnormS8Uint, vk.FormatD24UnormS8Uint, vk.FormatD32Sfloat, vk.FormatD32SfloatS8Uint:
		return true
	default:
		return false
	}
}

func GetFormatAspectFlags(format vk.Format) vk.ImageAspectFlags {
	if !IsDepthFormat(format) {
		return vk.ImageAspectFlags(vk.ImageAspectColorBit)
	}
	if format == vk.FormatD32Sfloat || format == vk.FormatD16Unorm {
		return vk.ImageAspectFlags(vk.ImageAspectDepthBit)
	}

	return vk.ImageAspectFlags(vk.ImageAspectDepthBit | vk.ImageAspectStencilBit)
}
