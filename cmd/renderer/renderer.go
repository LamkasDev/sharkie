package renderer

import (
	"image"
	"image/color"
	"os"
	"path"
	"runtime"

	"github.com/AllenDang/cimgui-go/imgui"
	g "github.com/AllenDang/giu"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
)

var GlobalRenderer *Renderer

type Renderer struct {
	Window      *g.MasterWindow
	FrameSource *FrameSource
	Overlay     *Overlay

	IconImage          image.Image
	IconTexture        *g.Texture
	FramebufferTexture *g.ReflectiveBoundTexture
}

func NewRenderer() *Renderer {
	r := &Renderer{
		Window: g.NewMasterWindow(
			"sharkie",
			1280,
			720,
			0,
		),
		FrameSource:        NewFrameSource(),
		Overlay:            NewOverlay(),
		FramebufferTexture: &g.ReflectiveBoundTexture{},
	}
	io := imgui.CurrentIO()
	io.SetConfigFlags(io.ConfigFlags() & ^imgui.ConfigFlagsViewportsEnable)
	r.Window.SetBgColor(color.RGBA{R: 10, G: 10, B: 12, A: 255})
	r.Window.SetTargetFPS(60)

	var err error
	r.IconImage, err = g.LoadImage(path.Join("winres", "icon.png"))
	if err != nil {
		panic(err)
	}
	r.Window.SetIcon(r.IconImage)

	return r
}

func (r *Renderer) LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (r *Renderer) Run() {
	go r.ConsumeFrames()
	r.Window.Run(r.Loop)
}

func (r *Renderer) Loop() {
	w, h := r.Window.GetSize()
	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	r.DrawFramebuffer(float32(w), float32(h))
	imgui.PopStyleVarV(3)
	r.DrawOverlay()
}

func (r *Renderer) ConsumeFrames() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for rawFrame := range r.FrameSource.ch {
		frame := rawFrame
		r.Overlay.LastFlip.Store(&frame)
		r.Overlay.FrameCount.Add(1)

		// Snapshot the guest framebuffer and push it to the texture.
		if framebuffer := r.Overlay.GuestFramebuffers.Load(); framebuffer != nil {
			texture := framebuffer.Snapshot()
			if err := r.FramebufferTexture.SetSurfaceFromRGBA(texture, false); err == nil {
				g.Update()
			}
		} else {
			g.Update()
		}
	}
}

func (r *Renderer) RegisterFramebuffer(address uintptr, attribute *VideoOutBufferAttribute) {
	r.Overlay.GuestFramebuffers.Store(NewGuestFramebuffer(
		address, int(attribute.Width), int(attribute.Height),
		int(attribute.PitchInPixel), int(attribute.TilingMode),
	))
}

func SetupRenderer() {
	GlobalRenderer = NewRenderer()
}
