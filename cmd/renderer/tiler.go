package renderer

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

// TiledByteOffset returns the byte offset of pixel in a 32 bpp
// ARRAY_2D_TILED_THIN1 surface with the given pitch in pixels.
func TiledByteOffset(x, y, pitchAligned int) int {
	pixelX, pixelY := x&(TileMicroWidth-1), y&(TileMicroHeight-1)
	microTileX, microTileY := x>>3, y>>3

	// Pixel byte offset within the micro-tile.
	pixelOffset := (pixelY*TileMicroWidth + pixelX) * 4

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
	macroTileBase := (macroY*pitchInMacrotiles + macroX) * TileMacroBytes

	return macroTileBase +
		channel*TileMicrosPerChannel*TileMicroBytes +
		localIndex*TileMicroBytes +
		pixelOffset
}

// Detile2D converts a 32 bpp ARRAY_2D_TILED_THIN1 surface stored at src into
// a linear RGBA row-major image written to dst.
func Detile2D(src []byte, dst []byte, width, height, pitchPixels int) {
	for y := range height {
		for x := range width {
			tiledOffset := TiledByteOffset(x, y, pitchPixels)
			linearOffset := (y*width + x) * 4

			// A8R8G8B8 little-endian is [B, G, R, A] in memory.
			// Go image.RGBA.Pix is [R, G, B, A] in memory.
			dst[linearOffset+0] = src[tiledOffset+2]
			dst[linearOffset+1] = src[tiledOffset+1]
			dst[linearOffset+2] = src[tiledOffset+0]
			dst[linearOffset+3] = src[tiledOffset+3]
		}
	}
}

// DetileLinear copies a linear 32 bpp surface to dst, converting A8R8G8B8 to RGBA byte order.
// Used when TilingMode == 1 (linear).
func DetileLinear(src []byte, dst []byte, width, height, pitchPixels int) {
	for y := range height {
		for x := range width {
			srcOffset := (y*pitchPixels + x) * 4
			dstOffset := (y*width + x) * 4

			// A8R8G8B8 little-endian is [B, G, R, A] in memory.
			// Go image.RGBA.Pix is [R, G, B, A] in memory.
			dst[dstOffset+0] = src[srcOffset+2]
			dst[dstOffset+1] = src[srcOffset+1]
			dst[dstOffset+2] = src[srcOffset+0]
			dst[dstOffset+3] = src[srcOffset+3]
		}
	}
}
