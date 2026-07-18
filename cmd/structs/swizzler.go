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
	case 2, 3, 4, 5, 6, 25: // R16, R8G8, B5G6R5, etc
		return 2
	case 8, 10, 11, 26, 34: // R8G8B8A8, B8G8R8A8, R10G10B10A2, Format5_9_9_9
		return 4
	case 12, 13, 14, 35, 38: // R16G16B16A16, R32G32, BC1, BC4
		return 8
	case 15, 36, 37, 39, 40, 41: // R32G32B32A32, BC2, BC3, BC5, BC6, BC7
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
	swizzleWidth := uint32(descriptor.Width)
	swizzleHeight := uint32(descriptor.Height)
	swizzlePitch := uint32(descriptor.Pitch)
	isBlock := descriptor.DataFormat >= 35 && descriptor.DataFormat <= 41
	if isBlock {
		swizzleWidth = (swizzleWidth + 3) / 4
		swizzleHeight = (swizzleHeight + 3) / 4
		swizzlePitch = (swizzlePitch + 3) / 4
	}
	swizzledData := make([]byte, swizzlePitch*swizzleHeight*bpp)

	// Stub for Tiling logic
	// In the next refinement we will add Morton order (Z-order) curve translation.
	// For now, we will perform a direct linear copy to verify the framework works.
	for y := uint32(0); y < swizzleHeight; y++ {
		for x := uint32(0); x < swizzleWidth; x++ {
			linearOffset := (y*swizzleWidth + x) * bpp
			swizzledOffset := (y*swizzlePitch + x) * bpp

			if linearOffset+bpp <= uint32(len(linearData)) && swizzledOffset+bpp <= uint32(len(swizzledData)) {
				copy(swizzledData[swizzledOffset:swizzledOffset+bpp], linearData[linearOffset:linearOffset+bpp])
			}
		}
	}

	return swizzledData
}
