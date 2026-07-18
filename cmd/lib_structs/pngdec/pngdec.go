package pngdec

type OrbisPngDecParseParam struct {
	PngMemAddr uintptr
	PngMemSize uint32
	Reserved   uint32
}

type OrbisPngDecImageInfo struct {
	ImageWidth  uint32
	ImageHeight uint32
	ColorSpace  uint16
	BitDepth    uint16
	ImageFlag   uint32
}

type PngDecDecodeParam struct {
	PngMemAddr   uintptr
	ImageMemAddr uintptr
	PngMemSize   uint32
	ImageMemSize uint32
	PixelFormat  uint16
	AlphaValue   uint16
	ImagePitch   uint32
}
