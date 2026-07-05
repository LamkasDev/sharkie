package structs

import "github.com/LamkasDev/sharkie/cmd/spirv/structs"

// Texture metadata for swizzling
type TextureInfo struct {
	Width         uint32
	Height        uint32
	Pitch         uint32
	Format        uint8
	TilingIndex   uint8
	BytesPerPixel uint32
}

// GetBytesPerPixel returns the size of a single pixel based on the GCN DataFormat.
func GetBytesPerPixel(format uint8) uint32 {
	switch format {
	case 1: // R8_UNORM
		return 1
	case 2, 4, 5, 6, 25: // R16, R8G8, B5G6R5, etc
		return 2
	case 10, 11, 26: // R8G8B8A8, B8G8R8A8, R10G10B10A2
		return 4
	case 13, 14: // R16G16B16A16, R32G32
		return 8
	case 15: // R32G32B32A32
		return 16
	default:
		return 4 // Fallback
	}
}

// SwizzleTexture takes a linear pixel buffer and swizzles it into the PS4 tiled layout.
func SwizzleTexture(linearData []byte, descriptor structs.ImageDescriptor) []byte {
	// If it's linear (TilingIndex 0), no swizzling needed.
	if descriptor.TilingIndex == 0 {
		result := make([]byte, len(linearData))
		copy(result, linearData)
		return result
	}

	bpp := GetBytesPerPixel(descriptor.DataFormat)
	swizzledData := make([]byte, uint32(descriptor.Pitch*descriptor.Height)*bpp)

	// Stub for Tiling logic
	// In the next refinement we will add Morton order (Z-order) curve translation.
	// For now, we will perform a direct linear copy to verify the framework works.
	for y := uint32(0); y < uint32(descriptor.Height); y++ {
		for x := uint32(0); x < uint32(descriptor.Width); x++ {
			linearOffset := (y*uint32(descriptor.Width) + x) * bpp
			swizzledOffset := (y*uint32(descriptor.Pitch) + x) * bpp

			if linearOffset+bpp <= uint32(len(linearData)) && swizzledOffset+bpp <= uint32(len(swizzledData)) {
				copy(swizzledData[swizzledOffset:swizzledOffset+bpp], linearData[linearOffset:linearOffset+bpp])
			}
		}
	}

	return swizzledData
}

// DeswizzleTexture takes a swizzled PS4 tiled buffer and converts it to linear layout.
func DeswizzleTexture(swizzledData []byte, descriptor structs.ImageDescriptor) []byte {
	if descriptor.TilingIndex == 0 {
		result := make([]byte, len(swizzledData))
		copy(result, swizzledData)
		return result
	}

	bpp := GetBytesPerPixel(descriptor.DataFormat)
	linearData := make([]byte, uint32(descriptor.Width)*uint32(descriptor.Height)*bpp)

	for y := uint32(0); y < uint32(descriptor.Height); y++ {
		for x := uint32(0); x < uint32(descriptor.Width); x++ {
			linearOffset := (y*uint32(descriptor.Width) + x) * bpp
			swizzledOffset := (y*uint32(descriptor.Pitch) + x) * bpp

			if linearOffset+bpp <= uint32(len(linearData)) && swizzledOffset+bpp <= uint32(len(swizzledData)) {
				copy(linearData[linearOffset:linearOffset+bpp], swizzledData[swizzledOffset:swizzledOffset+bpp])
			}
		}
	}

	return linearData
}
