package renderer

import (
	"fmt"
	"path"
	"sync/atomic"

	"github.com/LamkasDev/sharkie/cmd/goutils"
	"github.com/LamkasDev/sharkie/cmd/structs/gc"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	"github.com/elokore/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/elokore/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/elokore/cimgui-go-vulkan/imgui"
	atomicc "go.uber.org/atomic"
)

var (
	colOverlayBg = imgui.Vec4{X: 12 / 255.0, Y: 12 / 255.0, Z: 14 / 255.0, W: 220 / 255.0}
	colBorder    = imgui.Vec4{X: 50 / 255.0, Y: 55 / 255.0, Z: 70 / 255.0, W: 1}
	colAccent    = imgui.Vec4{X: 82 / 255.0, Y: 130 / 255.0, Z: 255 / 255.0, W: 1}
	colMuted     = imgui.Vec4{X: 120 / 255.0, Y: 125 / 255.0, Z: 140 / 255.0, W: 1}
	colGreen     = imgui.Vec4{X: 80 / 255.0, Y: 220 / 255.0, Z: 120 / 255.0, W: 1}
	colYellow    = imgui.Vec4{X: 250 / 255.0, Y: 200 / 255.0, Z: 60 / 255.0, W: 1}
)

const ImguiOverlayFlags = imgui.WindowFlagsNoDecoration |
	imgui.WindowFlagsNoInputs |
	imgui.WindowFlagsNoMove

type ImguiOverlay struct {
	IconTexture imgui.TextureRef
	Font        *imgui.Font

	ShowOverlay   atomic.Bool
	FrameCount    atomic.Uint64
	FrameLastTime atomic.Int64
	Framerate     atomicc.Float64
}

func NewImguiOverlay(bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) *ImguiOverlay {
	overlay := &ImguiOverlay{}
	overlay.ShowOverlay.Store(true)

	io := imgui.CurrentIO()
	io.SetConfigFlags(io.ConfigFlags() & ^imgui.ConfigFlagsViewportsEnable)

	overlay.Font = io.Fonts().AddFontFromFileTTF(path.Join("data", "JetBrainsMono-Regular.ttf"))
	overlay.Font.SetLegacySize(15)
	io.SetFontDefault(overlay.Font)

	iconImage, err := goutils.LoadImage(path.Join("winres", "icon.png"))
	if err != nil {
		panic(err)
	}
	iconRGBA := backend.ImageToRgba(iconImage)
	overlay.IconTexture = bknd.CreateTextureRgba(iconRGBA, iconRGBA.Bounds().Dx(), iconRGBA.Bounds().Dy())

	return overlay
}

func (overlay *ImguiOverlay) Destroy(bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) {
	bknd.DeleteTexture(overlay.IconTexture)
}

func (overlay *ImguiOverlay) DrawOverlay(width, height uint32) {
	if !overlay.ShowOverlay.Load() {
		return
	}
	frameCount := overlay.FrameCount.Load()
	if frameCount == 0 {
		overlay.DrawWelcomeSplash(width, height)
	} else {
		overlay.DrawHud(frameCount)
	}
}

func (overlay *ImguiOverlay) DrawWelcomeSplash(width, height uint32) {
	const splashW, splashH = float32(420), float32(180)

	imgui.SetNextWindowPos(imgui.Vec2{X: (float32(width) - splashW) / 2, Y: (float32(height) - splashH) / 2})
	imgui.SetNextWindowSize(imgui.Vec2{X: splashW, Y: splashH})
	imgui.PushStyleColorVec4(imgui.ColWindowBg, colOverlayBg)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 48, Y: 48})
	if imgui.BeginV("##splash", nil, ImguiOverlayFlags) {
		// Title.
		imgui.Image(overlay.IconTexture, imgui.Vec2{X: 82, Y: 82})
		imgui.SameLine()

		imgui.SetCursorPosX(imgui.CursorPosX() + 6)
		imgui.SetCursorPosY(imgui.CursorPosY() + 6)
		imgui.PushStyleColorVec4(imgui.ColText, colAccent)
		imgui.PushFont(overlay.Font, 42)
		imgui.Text(" sharkie :3")
		imgui.PopFont()
		imgui.PopStyleColor()

		// Subtitle.
		imgui.SetCursorPosX(imgui.CursorPosX() + 104)
		imgui.SetCursorPosY(imgui.CursorPosY() - 30)
		imgui.PushFont(overlay.Font, 22)
		imgui.PushStyleColorVec4(imgui.ColText, colGreen)
		imgui.Text(" > PS4 emulator")
		imgui.PopFont()
		imgui.PopStyleColor()

		imgui.End()
	}
	imgui.PopStyleVar()
	imgui.PopStyleColor()
}

func (overlay *ImguiOverlay) DrawHud(frameCount uint64) {
	imgui.SetNextWindowPos(imgui.Vec2{X: 8, Y: 8})
	imgui.SetNextWindowSize(imgui.Vec2{X: 340, Y: 220})
	imgui.PushStyleColorVec4(imgui.ColWindowBg, colOverlayBg)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 14, Y: 14})
	if imgui.BeginV("##hud", nil, ImguiOverlayFlags) {
		imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 6, Y: 6})

		// Title.
		imgui.Image(overlay.IconTexture, imgui.Vec2{X: 48, Y: 48})
		imgui.SameLine()

		imgui.SetCursorPosY(imgui.CursorPosY() + 6)
		imgui.PushStyleColorVec4(imgui.ColText, colAccent)
		imgui.PushFont(overlay.Font, 32)
		imgui.Text(" sharkie :3")
		imgui.PopFont()
		imgui.PopStyleColor()

		// Display info.
		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()
		HudRow("frames", colGreen, fmt.Sprint(frameCount))
		HudRow("fps", colGreen, fmt.Sprintf("%.1f", overlay.Framerate.Load()))

		// Graphics card info.
		imgui.Spacing()
		imgui.Separator()
		imgui.Spacing()
		HudRow("ring slot", colGreen, RingSlotText())

		imgui.PopStyleVar()
		imgui.End()
	}
	imgui.PopStyleVar()
	imgui.PopStyleColor()
}

func HudRow(label string, valueColor imgui.Vec4, value string) {
	imgui.PushStyleColorVec4(imgui.ColText, colMuted)
	imgui.Text(label)
	imgui.PopStyleColor()
	imgui.SameLine()
	imgui.PushStyleColorVec4(imgui.ColText, valueColor)
	imgui.Text(value)
	imgui.PopStyleColor()
}

func RingSlotText() string {
	gpu.GlobalLiverpool.RingMutex.Lock()
	defer gpu.GlobalLiverpool.RingMutex.Unlock()
	return fmt.Sprintf("%d (%d & %d pending buffers)",
		gc.GlobalGraphicsController.ActiveRingSlot,
		len(gpu.GlobalLiverpool.ComputeRing.Pending),
		len(gpu.GlobalLiverpool.GraphicsRing.Pending))
}
