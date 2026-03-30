package gpu

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/structs/video"
)

type LiverpoolDisplaySurface struct {
	GpuAddress     uintptr
	PixelFormat    VideoOutPixelFormat
	TilingMode     VideoOutTilingMode
	Width          uint32
	Height         uint32
	PitchPixels    uint32
	AttributeIndex uint32
}

const LiverpoolDisplaySurfaceSize = unsafe.Sizeof(LiverpoolDisplaySurface{})
