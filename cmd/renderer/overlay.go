package renderer

import (
	"fmt"
	"image/color"
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
	g "github.com/AllenDang/giu"
	"github.com/LamkasDev/sharkie/cmd/structs/gc"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
)

var (
	colOverlayBg    = color.RGBA{R: 12, G: 12, B: 14, A: 220}
	colBorder       = color.RGBA{R: 50, G: 55, B: 70, A: 255}
	colAccent       = color.RGBA{R: 82, G: 130, B: 255, A: 255}
	colMuted        = color.RGBA{R: 120, G: 125, B: 140, A: 255}
	colGreen        = color.RGBA{R: 80, G: 220, B: 120, A: 255}
	colYellow       = color.RGBA{R: 250, G: 200, B: 60, A: 255}
	colCanvasBg     = color.RGBA{R: 15, G: 15, B: 18, A: 255}
	colWelcomeTitle = color.RGBA{R: 82, G: 130, B: 255, A: 255}
	colWelcomeSub   = color.RGBA{R: 160, G: 165, B: 180, A: 255}
)

const OverlayFlags = g.WindowFlagsNoDecoration |
	g.WindowFlagsNoInputs |
	g.WindowFlagsNoMove

type Overlay struct {
	FrameCount  atomic.Uint64
	LastFlip    atomic.Pointer[Frame]
	ShowOverlay atomic.Bool
}

func NewOverlay() *Overlay {
	s := &Overlay{}
	s.ShowOverlay.Store(true)

	return s
}

func (r *Renderer) DrawOverlay() {
	if r.IconTexture == nil {
		g.NewTextureFromRgba(r.IconImage, func(tex *g.Texture) {
			r.IconTexture = tex
		})
	}

	frameCount := r.Overlay.FrameCount.Load()
	if frameCount == 0 {
		r.DrawWelcomeSplash()
	} else {
		r.DrawHud(frameCount)
	}
}

func (r *Renderer) DrawWelcomeSplash() {
	const splashW, splashH = float32(420), float32(200)
	width, height := r.Window.GetSize()
	cx := (float32(width) - splashW) / 2
	cy := (float32(height) - splashH) / 2

	g.Window("##splash").
		Pos(cx, cy).
		Size(splashW, splashH).
		Flags(OverlayFlags).
		Layout(
			// Title.
			g.Row(
				g.Custom(func() {
					if r.IconTexture != nil {
						imgui.AlignTextToFramePadding()
						g.Image(r.IconTexture).Size(24, 24).Build()
						imgui.SameLine()
					}
				}),
				g.Style().SetColor(g.StyleColorText, colAccent).To(
					g.Label("sharkie :3"),
				),
			),
		)
}

func (r *Renderer) DrawHud(frameCount uint64) {
	const (
		padX     = float32(8)
		padY     = float32(8)
		overlayW = float32(340)
		overlayH = float32(190)
	)

	flip := r.Overlay.LastFlip.Load()
	g.Window("##hud").
		Pos(padX, padY).
		Size(overlayW, overlayH).
		Flags(OverlayFlags).
		Layout(
			g.Style().
				SetStyle(g.StyleVarItemSpacing, 6, 6).
				To(
					// Title.
					g.Row(
						g.Custom(func() {
							if r.IconTexture != nil {
								g.Image(r.IconTexture).Size(48, 48).Build()
							}
						}),
						g.Custom(func() {
							imgui.SetCursorPosY(imgui.CursorPosY() + 6)
							g.Style().SetFontSize(36).SetColor(g.StyleColorText, colAccent).
								To(g.Label(" sharkie :3")).Build()
						}),
					),

					// Frame counter.
					g.Spacing(),
					g.Separator(),
					g.Spacing(),
					g.Row(
						g.Style().SetColor(g.StyleColorText, colMuted).To(g.Label("frames")),
						g.Style().SetColor(g.StyleColorText, colGreen).To(g.Labelf("%d", frameCount)),
					),

					// Last flip address.
					g.Row(
						g.Style().SetColor(g.StyleColorText, colMuted).To(g.Label("last flip address")),
						g.Style().SetColor(g.StyleColorText, colYellow).To(FlipAddressWidget(flip)),
					),

					// Last flip arg.
					g.Row(
						g.Style().SetColor(g.StyleColorText, colMuted).To(g.Label("flip arg")),
						g.Style().SetColor(g.StyleColorText, colYellow).To(FlipArgWidget(flip)),
					),

					// Ring slot.
					g.Spacing(),
					g.Separator(),
					g.Spacing(),
					g.Row(
						g.Style().SetColor(g.StyleColorText, colMuted).To(g.Label("ring slot")),
						g.Style().SetColor(g.StyleColorText, colGreen).To(RingSlotWidget()),
					),
				),
		)
}

func FlipAddressWidget(f *Frame) g.Widget {
	if f == nil {
		return g.Label("-")
	}

	return g.Labelf("0x%X", f.GpuAddress)
}

func FlipArgWidget(f *Frame) g.Widget {
	if f == nil {
		return g.Label("-")
	}

	return g.Labelf("0x%X", f.FlipArg)
}

func RingSlotWidget() g.Widget {
	return g.Label(fmt.Sprintf(
		"%d (%d & %d pending buffers)",
		gc.GlobalGraphicsController.ActiveRingSlot,
		len(gpu.GlobalLiverpool.ComputeRing.Pending),
		len(gpu.GlobalLiverpool.GraphicsRing.Pending),
	))
}
