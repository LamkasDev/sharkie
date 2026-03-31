package renderer

import (
	"image"
	"image/color"
	"unsafe"
)

// GuestFramebuffer describes a registered guest display surface.
type GuestFramebuffer struct {
	// GpuAddress is the guest virtual address of the surface.
	GpuAddress uintptr

	// Width and Height in pixels.
	Width, Height int

	// PitchPixels is the padded row stride in pixels.
	PitchPixels int

	// Tiled is true when the surface uses ARRAY_2D_TILED_THIN1 layout.
	Tiled bool

	// reuseBuf is a scratch buffer reused between Snapshot calls.
	reuseBuf []byte
}

func NewGuestFramebuffer(gpuAddress uintptr, width, height, pitchPixels, tilingMode int) *GuestFramebuffer {
	// Align to macro-tile width.
	if pitchPixels == 0 {
		pitchPixels = (width + TileMacroWidth - 1) &^ (TileMacroWidth - 1)
	}

	return &GuestFramebuffer{
		GpuAddress:  gpuAddress,
		Width:       width,
		Height:      height,
		PitchPixels: pitchPixels,
		Tiled:       tilingMode == 0,
		reuseBuf:    make([]byte, width*height*4),
	}
}

// Snapshot reads the current frame from guest memory and returns a *image.RGBA.
func (fb *GuestFramebuffer) Snapshot() *image.RGBA {
	// Align height for tiled framebuffers (macro-tiles need to be 128x128).
	alignedHeight := fb.Height
	if fb.Tiled {
		alignedHeight = (fb.Height + 127) &^ 127
	}
	totalBytes := fb.PitchPixels * alignedHeight * 4

	// Detile frame into our scratch buffer.
	src := unsafe.Slice((*byte)(unsafe.Pointer(fb.GpuAddress)), totalBytes)
	if fb.Tiled {
		Detile2D(src, fb.reuseBuf, fb.Width, fb.Height, fb.PitchPixels)
	} else {
		DetileLinear(src, fb.reuseBuf, fb.Width, fb.Height, fb.PitchPixels)
	}

	// If nothing drawn, use test pattern.
	if allZero(fb.reuseBuf) {
		return testPattern(fb.Width, fb.Height)
	}

	// Build image.RGBA backed by our scratch buffer.
	img := &image.RGBA{
		Pix:    make([]byte, len(fb.reuseBuf)),
		Stride: fb.Width * 4,
		Rect:   image.Rect(0, 0, fb.Width, fb.Height),
	}
	copy(img.Pix, fb.reuseBuf)

	return img
}

// allZero returns true when every byte is 0.
func allZero(data []byte) bool {
	// Check in 8-byte chunks first for speed on large buffers.
	words := len(data) / 8
	for i := range words {
		word := *(*uint64)(unsafe.Pointer(&data[i*8]))
		if word != 0 {
			return false
		}
	}
	for i := words * 8; i < len(data); i++ {
		if data[i] != 0 {
			return false
		}
	}

	return true
}

// testPattern generates a test pattern image.
func testPattern(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	centerX, centerY := width/2, height/2
	for y := range height {
		for x := range width {
			var c color.RGBA
			switch {
			case x == centerX || y == centerY:
				c = color.RGBA{R: 255, G: 0, B: 200, A: 255}
			case x%64 == 0 && y%64 == 0:
				c = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			default:
				bVal := uint8(20 + (x*60)/width)
				gVal := uint8(20 + (y*30)/height)
				c = color.RGBA{R: 18, G: gVal, B: bVal, A: 255}
			}

			img.SetRGBA(x, y, c)
		}
	}

	return img
}
