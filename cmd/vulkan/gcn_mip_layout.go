package vulkan

import (
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
)

type MipLayout struct {
	Offset int
	Size   int
	Pitch  int
	Height int
	Width  int
}

func bitCeil(v uint32) uint32 {
	if v <= 1 {
		return 1
	}
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	return v + 1
}

func imageSizeLinearAligned(pitch, height, bpp, numSamples uint32) (pitchAligned, heightAligned uint32, size int) {
	pitchAlign := max(uint32(8), 64/((bpp+7)/8))
	pitchAligned = (pitch + pitchAlign - 1) &^ (pitchAlign - 1)
	heightAligned = height
	logSz := pitchAligned * heightAligned * numSamples
	sliceAlign := max(uint32(64), 256/((bpp+7)/8))
	for logSz%sliceAlign != 0 {
		pitchAligned += pitchAlign
		logSz = pitchAligned * heightAligned * numSamples
	}

	return pitchAligned, heightAligned, int((logSz*bpp + 7) / 8)
}

func imageSizeMicroTiled(pitch, height, thickness, bpp, numSamples uint32) (pitchAligned, heightAligned uint32, size int) {
	const pitchAlign, heightAlign uint32 = 8, 8
	pitchAligned = (pitch + pitchAlign - 1) &^ (pitchAlign - 1)
	heightAligned = (height + heightAlign - 1) &^ (heightAlign - 1)
	logSz := (pitchAligned * heightAligned * bpp * numSamples) / 8
	for (logSz*thickness)%256 != 0 {
		pitchAligned += pitchAlign
		logSz = (pitchAligned * heightAligned * bpp * numSamples) / 8
	}

	return pitchAligned, heightAligned, int(logSz)
}

func imageSizeMacroTiled(pitch, height, thickness, bpp, numSamples uint32) (pitchAligned, heightAligned uint32, size int) {
	var pitchAlign, heightAlign uint32
	switch bpp {
	case 8: // 1 byte
		pitchAlign, heightAlign = 256, 128
	case 16: // 2 bytes
		pitchAlign, heightAlign = 128, 128
	case 32: // 4 bytes
		pitchAlign, heightAlign = 128, 64
	default:
		pitchAlign, heightAlign = 64, 64
	}
	pitchAligned = (pitch + pitchAlign - 1) &^ (pitchAlign - 1)
	heightAligned = (height + heightAlign - 1) &^ (heightAlign - 1)
	logSz := (pitchAligned * heightAligned * bpp * numSamples) / 8

	return pitchAligned, heightAligned, int(logSz)
}

func mipTexelSize(baseWidth, baseHeight, basePitch uint16, level uint8, pow2Pad bool) (uint32, uint32) {
	mipW := uint32(basePitch) >> level
	mipH := uint32(baseHeight) >> level
	if mipW < 1 {
		mipW = 1
	}
	if mipH < 1 {
		mipH = 1
	}
	if pow2Pad {
		mipW = bitCeil(mipW)
		mipH = bitCeil(mipH)
	}

	return mipW, mipH
}

func computeMipLayouts(descriptor spirvStructs.ImageDescriptor, numLevels uint8) []MipLayout {
	if numLevels < 1 {
		numLevels = 1
	}
	if numLevels > 16 {
		numLevels = 16
	}

	bpp := uint32(gcn.GetBytesPerPixel(descriptor.DataFormat) * 8)
	linear := isLinearTileMode(descriptor.TilingIndex)
	isBlock := descriptor.DataFormat >= 35 && descriptor.DataFormat <= 41

	layouts := make([]MipLayout, numLevels)
	guestSize := 0
	for mip := range numLevels {
		mipW, mipH := mipTexelSize(descriptor.Width, descriptor.Height, descriptor.Pitch, uint8(mip), descriptor.Pow2Pad)
		if isBlock {
			mipW = (mipW + 3) / 4
			mipH = (mipH + 3) / 4
		}
		if mipW < 1 {
			mipW = 1
		}
		if mipH < 1 {
			mipH = 1
		}

		var pitchAligned, heightAligned uint32
		var size int
		if linear {
			pitchAligned, heightAligned, size = imageSizeLinearAligned(mipW, mipH, bpp, 1)
		} else if is1DTiledMode(descriptor.TilingIndex) {
			pitchAligned, heightAligned, size = imageSizeMicroTiled(mipW, mipH, 1, bpp, 1)
		} else if isMacroTiledMode(descriptor.TilingIndex) {
			pitchAligned, heightAligned, size = imageSizeMacroTiled(mipW, mipH, 1, bpp, 1)
		} else {
			pitchAligned, heightAligned, size = imageSizeMicroTiled(mipW, mipH, 1, bpp, 1)
		}

		if isBlock {
			pitchAligned = max(pitchAligned*4, 32)
			heightAligned = max(heightAligned*4, 32)
		}

		texW := max(uint32(descriptor.Width)>>uint(mip), 1)
		texH := max(uint32(descriptor.Height)>>uint(mip), 1)
		texHeight := int(heightAligned)
		if int(texH) < texHeight {
			texHeight = int(texH)
		}

		if isBlock {
			// for texture height limitation
			texHeight = max(texHeight*4, 32)
		}

		layouts[mip] = MipLayout{
			Offset: guestSize,
			Size:   size,
			Pitch:  int(pitchAligned),
			Height: texHeight,
			Width:  int(texW),
		}
		guestSize += size
	}

	return layouts
}

func mipLevelCount(descriptor spirvStructs.ImageDescriptor) uint8 {
	maxDim := max(descriptor.Width, descriptor.Height)
	computed := 1
	for maxDim > 1 {
		maxDim /= 2
		computed++
	}
	fromDesc := int(descriptor.LastLevel) + 1
	count := max(computed, fromDesc)
	if count > 16 {
		count = 16
	}

	return uint8(count)
}
