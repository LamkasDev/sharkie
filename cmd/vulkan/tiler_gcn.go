package vulkan

const (
	// TileNumPipes is the Liverpool pipe count.
	TileNumPipes = 8

	// TileNumBanks is the bank count for 32 bpp display surfaces.
	TileNumBanks = 4

	// TileMicroWidth and TileMicroHeight are the micro-tile pixel dimensions.
	TileMicroHeight = 8
	TileMicroWidth  = 8

	// TileMicroBytes is the size of one 32 bpp micro-tile.
	TileMicroBytes = TileMicroWidth * TileMicroHeight * 4

	// TileMacroWidth and TileMacroHeight are the macro-tile pixel dimensions.
	TileMacroWidth  = TileNumPipes * TileMicroWidth
	TileMacroHeight = TileNumBanks * 4 * TileMicroHeight

	// TileMacroBytes is the size of one 32 bpp macro-tile.
	TileMacroBytes = TileMacroWidth * TileMacroHeight * 4

	// TileChannels is the number of pipe x bank channels per macro-tile.
	TileChannels = TileNumPipes * TileNumBanks

	// TileMicrosPerChannel is the number of micro-tiles assigned to each channel.
	TileMicrosPerChannel = (TileMacroWidth / TileMicroWidth) * (TileMacroHeight / TileMicroHeight) / TileChannels
)

// GCN tile-mode indices (AmdGpu::TileMode).
const (
	tileModeDisplayLinearAligned = 8
	tileModeDisplay1DThin        = 9
	tileModeDisplay2DThin        = 10
	tileModeThin1DThin           = 13
	tileModeThin2DThin           = 14
	tileModeDisplayLinearGeneral = 31
)

const microTilePixelSize = 8

// macro tile parameters for 32 bpp Thin2DThin (AmdGpu::TileMode::Thin2DThin).
const (
	thin2DPipeBits        = 3
	thin2DPipeInterleave  = 8
	thin2DBankBits        = 2
	thin2DNumPipes        = 8
	thin2DNumBanks        = 4
	thin2DBankWidth       = 1
	thin2DBankHeight      = 2
	thin2DMicroTileBytes  = 256
	thin2DMacroTilePitch  = 128
	thin2DMacroTileHeight = 64
)

func bitAt(v, b int) int {
	return (v >> b) & 1
}

func isLinearTileMode(tilingIndex uint8) bool {
	return tilingIndex == tileModeDisplayLinearAligned || tilingIndex == tileModeDisplayLinearGeneral
}

func is1DTiledMode(tilingIndex uint8) bool {
	switch tilingIndex {
	case 5, 9, 13, 19: // Depth1DThin, Display1DThin, Thin1DThin, Thick1DThick
		return true
	default:
		return false
	}
}

func is2DDisplayTiledMode(tilingIndex uint8) bool {
	switch tilingIndex {
	case 6, 7, 10, 11, 12: // depth/display 2D thin variants
		return true
	default:
		return false
	}
}

func is2DThinTiledMode(tilingIndex uint8) bool {
	switch tilingIndex {
	case 14, 15, 16, 17, 18, 20, 21, 22, 23, 24, 25, 26:
		return true
	default:
		return false
	}
}

func isMacroTiledMode(tilingIndex uint8) bool {
	return !isLinearTileMode(tilingIndex) && !is1DTiledMode(tilingIndex)
}

func usesDisplayMicroTiling(tilingIndex uint8) bool {
	switch tilingIndex {
	case 6, 7, 8, 9, 10, 11, 12, 31:
		return true
	default:
		return false
	}
}

func microTileBytes(bpp int) int {
	return 8 * 8 * bpp
}

func microPixelIndex(x, y int, bpp int, displayMicro bool) int {
	x0, x1, x2 := bitAt(x, 0), bitAt(x, 1), bitAt(x, 2)
	y0, y1, y2 := bitAt(y, 0), bitAt(y, 1), bitAt(y, 2)

	if displayMicro {
		switch bpp {
		case 4:
			return x0 | (x1 << 1) | (y0 << 2) | (x2 << 3) | (y1 << 4) | (y2 << 5)
		case 2:
			return x0 | (x1 << 1) | (x2 << 2) | (y0 << 3) | (y1 << 4) | (y2 << 5)
		case 1:
			return x0 | (x1 << 1) | (x2 << 2) | (y1 << 3) | (y0 << 4) | (y2 << 5)
		default:
			return x0 | (x1 << 1) | (y0 << 2) | (x2 << 3) | (y1 << 4) | (y2 << 5)
		}
	}

	switch bpp {
	default:
		return x0 | (y0 << 1) | (x1 << 2) | (y1 << 3) | (x2 << 4) | (y2 << 5)
	}
}

func offset1DTiled(x, y, pitch, bpp int, displayMicro bool) int {
	microTilesPerRow := pitch / 8
	microTileX := x / 8
	microTileY := y / 8
	microTileOffset := (microTileY*microTilesPerRow + microTileX) * microTileBytes(bpp)
	pixelIndex := microPixelIndex(x, y, bpp, displayMicro)
	return microTileOffset + pixelIndex*bpp
}

func copyLinearRows(src, dst []byte, width, height, pitch, bpp int) {
	for y := range height {
		srcRow := y * pitch * bpp
		dstRow := y * width * bpp
		copy(dst[dstRow:dstRow+width*bpp], src[srcRow:srcRow+width*bpp])
	}
}

func copyTiledGeneric(src, dst []byte, width, height, pitch, bpp int, offsetFn func(x, y int) int) {
	for y := range height {
		for x := range width {
			tiledOffset := offsetFn(x, y)
			linearOffset := (y*width + x) * bpp
			if tiledOffset+bpp > len(src) || linearOffset+bpp > len(dst) {
				continue
			}
			copy(dst[linearOffset:linearOffset+bpp], src[tiledOffset:tiledOffset+bpp])
		}
	}
}

func detileMicroTileBlock(srcTile []byte, dst []byte, tileX, tileY, width, height, bpp int, displayMicro bool) {
	for y := range microTilePixelSize {
		py := tileY + y
		if py >= height {
			return
		}
		for x := range microTilePixelSize {
			px := tileX + x
			if px >= width {
				continue
			}
			srcOff := microPixelIndex(x, y, bpp, displayMicro) * bpp
			dstOff := (py*width + px) * bpp
			if srcOff+bpp > len(srcTile) || dstOff+bpp > len(dst) {
				continue
			}
			copy(dst[dstOff:dstOff+bpp], srcTile[srcOff:srcOff+bpp])
		}
	}
}

func copyTiledMicroBlocks(src, dst []byte, width, height, pitch, bpp int, displayMicro bool, offsetFn func(x, y int) int) {
	tileBytes := microTileBytes(bpp)
	for tileY := 0; tileY < height; tileY += microTilePixelSize {
		for tileX := 0; tileX < width; tileX += microTilePixelSize {
			tiledBase := offsetFn(tileX, tileY)
			if tiledBase+tileBytes > len(src) {
				continue
			}
			detileMicroTileBlock(src[tiledBase:tiledBase+tileBytes], dst, tileX, tileY, width, height, bpp, displayMicro)
		}
	}
}

func offsetThin2D(x, y, pitch, height, bpp int) int {
	pixelIndex := microPixelIndex(x, y, bpp, false)
	elementOffset := pixelIndex * bpp

	macroTilePitch := 128
	macroTileHeight := 64
	bankWidth := 1
	bankHeight := 2

	switch bpp {
	case 1:
		macroTilePitch = 256
		macroTileHeight = 128
		bankHeight = 4
	case 2:
		macroTilePitch = 128
		macroTileHeight = 128
		bankHeight = 4
	case 8, 16:
		macroTilePitch = 64
		macroTileHeight = 64
		bankHeight = 2
	}

	microTileBytes := bpp * 64
	macroTileBytes := microTileBytes *
		(macroTilePitch / 8) *
		(macroTileHeight / 8) /
		(thin2DNumPipes * thin2DNumBanks)
	macroTilesPerRow := pitch / macroTilePitch
	macroTileX := x / macroTilePitch
	macroTileY := y / macroTileHeight
	macroTileOffset := (macroTileY*macroTilesPerRow + macroTileX) * macroTileBytes
	macroTilesPerSlice := macroTilesPerRow * (height / macroTileHeight)
	sliceBytes := macroTilesPerSlice * macroTileBytes

	tileRowIndex := (y / 8) % bankHeight
	tileColumnIndex := ((x / 8) / thin2DNumPipes) % bankWidth
	tileIndex := tileRowIndex*bankWidth + tileColumnIndex
	tileOffset := tileIndex * microTileBytes

	totalOffset := sliceBytes*0 + macroTileOffset + elementOffset + tileOffset

	tx, ty := x/8, y/8
	x3, x4, x5 := bitAt(tx, 0), bitAt(tx, 1), bitAt(tx, 2)
	y3, y4, y5 := bitAt(ty, 0), bitAt(ty, 1), bitAt(ty, 2)
	pipe := (x3 ^ y3 ^ x4) | ((x3 ^ y4) << 1) | ((x5 ^ y5) << 2)

	txMacro, tyMacro := x/(8*bankWidth*thin2DNumPipes), y/(8*bankHeight)
	x3m, x4m := bitAt(txMacro, 0), bitAt(txMacro, 1)
	y3m, y4m := bitAt(tyMacro, 0), bitAt(tyMacro, 1)
	bank := (x3m ^ y4m) | ((x4m ^ y3m) << 1)

	pipeInterleaveMask := (1 << thin2DPipeInterleave) - 1
	pipeInterleaveOffset := totalOffset & pipeInterleaveMask
	offset := totalOffset >> thin2DPipeInterleave

	addr := pipeInterleaveOffset
	addr |= pipe << thin2DPipeInterleave
	addr |= bank << (thin2DPipeInterleave + thin2DPipeBits)
	addr |= offset << (thin2DPipeInterleave + thin2DPipeBits + thin2DBankBits)
	return addr
}

// DetileToLinear converts guest GCN texture memory into a tightly-packed linear buffer.
func DetileToLinear(src, dst []byte, width, height, pitch int, tilingIndex uint8, bpp int) {
	if width <= 0 || height <= 0 || pitch <= 0 || bpp <= 0 {
		return
	}

	switch {
	case isLinearTileMode(tilingIndex):
		copyLinearRows(src, dst, width, height, pitch, bpp)
	case is1DTiledMode(tilingIndex):
		displayMicro := usesDisplayMicroTiling(tilingIndex)
		copyTiledMicroBlocks(src, dst, width, height, pitch, bpp, displayMicro, func(x, y int) int {
			return offset1DTiled(x, y, pitch, bpp, displayMicro)
		})
	case is2DDisplayTiledMode(tilingIndex):
		copyTiledMicroBlocks(src, dst, width, height, pitch, bpp, true, func(x, y int) int {
			return TiledByteOffset(x, y, pitch, bpp)
		})
	case is2DThinTiledMode(tilingIndex):
		offsetFn := func(x, y int) int {
			return offsetThin2D(x, y, pitch, height, bpp)
		}
		copyTiledGeneric(src, dst, width, height, pitch, bpp, offsetFn)
	default:
		copyLinearRows(src, dst, width, height, pitch, bpp)
	}
}

// TiledByteOffset returns the byte offset of pixel in a
// ARRAY_2D_TILED_DISPLAY surface with the given pitch in pixels.
func TiledByteOffset(x, y, pitchAligned, bpp int) int {
	pixelX, pixelY := x&(TileMicroWidth-1), y&(TileMicroHeight-1)
	microTileX, microTileY := x>>3, y>>3

	// Pixel byte offset within the micro-tile.
	pixelOffset := microPixelIndex(pixelX, pixelY, bpp, true) * bpp

	// Pipe is 3 bits (XOR of micro-tile coordinates modulo 8).
	pipe := (microTileX ^ microTileY) & 7

	// Bank is 2 bits (uses bits above the micro-tile index level).
	// bank[0] = microTileY[3] XOR microTileX[0]  =>  (y>>6 ^ x>>3) & 1
	// bank[1] = microTileY[4] XOR microTileX[1]  =>  (y>>7 ^ x>>4) & 1
	bank := ((y>>6 ^ x>>3) & 1) | (((y>>7 ^ x>>4) & 1) << 1)

	// Channel index.
	channel := bank*TileNumPipes + pipe

	// Local micro-tile x within macro-tile.
	localMicroTileX := microTileX % (TileMacroWidth / TileMicroWidth)
	localIndex := (localMicroTileX & 1) | ((localMicroTileX >> 2 & 1) << 1)

	// Macro-tile position.
	macroX, macroY := x/TileMacroWidth, y/TileMacroHeight
	pitchInMacrotiles := pitchAligned / TileMacroWidth
	macroTileBytes := TileMacroWidth * TileMacroHeight * bpp
	microTileBytes := bpp * 64

	return macroY*pitchInMacrotiles*macroTileBytes + macroX*macroTileBytes +
		channel*TileMicrosPerChannel*microTileBytes +
		localIndex*microTileBytes +
		pixelOffset
}
