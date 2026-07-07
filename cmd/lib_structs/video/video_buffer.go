package video

import "unsafe"

type VideoOutBuffer struct {
	GpuAddress     uintptr
	AttributeIndex uint32
	Registered     bool
}

const VideoOutBufferSize = unsafe.Sizeof(VideoOutBuffer{})

type VideoOutBufferAttribute struct {
	PixelFormat  VideoOutPixelFormat
	TilingMode   VideoOutTilingMode
	AspectRatio  VideoOutAspectRatio
	Width        uint32
	Height       uint32
	PitchInPixel uint32
	Option       uint32
	Reserved     [16]byte
}

const VideoOutBufferAttributeSize = unsafe.Sizeof(VideoOutBufferAttribute{})
