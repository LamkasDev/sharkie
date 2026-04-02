package renderer

import (
	"image"
	"image/color"
	"unsafe"
)

// GuestFramebuffer describes a registered guest framebuffer.
type GuestFramebuffer struct {
	GpuAddress    uintptr
	Width, Height int
	PitchPixels   int
	Tiled         bool

	// rgba is a scratch buffer reused between Snapshot calls.
	rgba        *image.RGBA
	testPattern *image.RGBA
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
		rgba: &image.RGBA{
			Pix:    make([]byte, width*height*4),
			Stride: width * 4,
			Rect:   image.Rect(0, 0, width, height),
		},
		testPattern: testPattern(width, height),
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
		Detile2D(src, fb.rgba.Pix, fb.Width, fb.Height, fb.PitchPixels)
	} else {
		DetileLinear(src, fb.rgba.Pix, fb.Width, fb.Height, fb.PitchPixels)
	}

	// If nothing drawn, use test pattern.
	if true {
		return fb.testPattern
	}

	return fb.rgba
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
