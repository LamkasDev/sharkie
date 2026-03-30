package renderer

import (
	"image"
	"image/color"
	"os"
	"path"
	"runtime"

	"github.com/AllenDang/cimgui-go/imgui"
	g "github.com/AllenDang/giu"
)

var GlobalRenderer *Renderer

type Renderer struct {
	Window      *g.MasterWindow
	FrameSource *FrameSource
	Overlay     *Overlay

	IconImage   image.Image
	IconTexture *g.Texture
}

func NewRenderer() *Renderer {
	r := &Renderer{
		Window: g.NewMasterWindow(
			"sharkie",
			1280,
			720,
			0,
		),
		FrameSource: NewFrameSource(),
		Overlay:     NewOverlay(),
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
	r.DrawOverlay()
}

func (r *Renderer) ConsumeFrames() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for frame := range r.FrameSource.ch {
		f := frame
		r.Overlay.LastFlip.Store(&f)
		r.Overlay.FrameCount.Add(1)
		g.Update()
	}
}

func SetupRenderer() {
	GlobalRenderer = NewRenderer()
}
