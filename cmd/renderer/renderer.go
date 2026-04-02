package renderer

import (
	"runtime"
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/structs/video"
	"github.com/elokore/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/elokore/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/elokore/cimgui-go-vulkan/imgui"
	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
)

type Renderer struct {
	Handles     VulkanHandles
	Backend     backend.Backend[glfwvulkanbackend.GLFWWindowFlags]
	FrameSource *FrameSource
	Overlay     *ImguiOverlay

	SwapchainDimensions *as.SwapchainDimensions
	Depth               *Depth
	RenderPass          vk.RenderPass
	PipelineCache       vk.PipelineCache

	// TODO: framebuffer texture is already double-buffered, so
	//   	 maybe we should split it so it makes more sense?
	FramebufferMutex   sync.RWMutex
	Framebuffers       map[uintptr]*GuestFramebuffer
	FramebufferTexture *FramebufferTexture
}

func NewRenderer(context as.Context, dimensions *as.SwapchainDimensions) *Renderer {
	r := &Renderer{
		Handles:             NewVulkanHandles(context),
		SwapchainDimensions: dimensions,
		FrameSource:         NewFrameSource(),
		Framebuffers:        make(map[uintptr]*GuestFramebuffer),
	}
	r.Backend, _ = backend.CreateBackend(glfwvulkanbackend.NewGLFWBackend())
	r.Depth = NewDepth(r)
	r.prepareRenderPass()
	r.preparePipelineCache()
	r.prepareFramebuffers()

	return r
}

func (r *Renderer) Destroy() {
	vk.DeviceWaitIdle(r.Handles.Device)
	r.Backend.Cleanup()
	if r.FramebufferTexture != nil {
		r.FramebufferTexture.Destroy(&r.Handles)
		r.FramebufferTexture = nil
	}
	vk.DestroyPipelineCache(r.Handles.Device, r.PipelineCache, nil)
	vk.DestroyRenderPass(r.Handles.Device, r.RenderPass, nil)
	r.Depth.Destroy(r)
	r.Handles.Destroy()
}

func (r *Renderer) Render() {
	r.DrawFramebuffer()
	r.Overlay.DrawOverlay(r.SwapchainDimensions.Width, r.SwapchainDimensions.Height)
}

func (r *Renderer) DrawFramebuffer() {
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowSize(imgui.Vec2{X: float32(r.SwapchainDimensions.Width), Y: float32(r.SwapchainDimensions.Height)})
	imgui.PushStyleColorVec4(imgui.ColWindowBg, imgui.Vec4{X: 10 / 255.0, Y: 10 / 255.0, Z: 12 / 255.0, W: 1.0})
	if imgui.BeginV("##fb", nil, ImguiOverlayFlags|imgui.WindowFlagsNoBringToFrontOnFocus) {
		if r.FramebufferTexture != nil && r.FramebufferTexture.TextureId.CData != nil {
			imgui.Image(r.FramebufferTexture.TextureId, imgui.Vec2{
				X: float32(r.SwapchainDimensions.Width),
				Y: float32(r.SwapchainDimensions.Height),
			})
		}
		imgui.End()
	}
	imgui.PopStyleColor()
}

func (r *Renderer) ConsumeFrames(done chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(done)

	for rawFrame := range r.FrameSource.Channel {
		frame := rawFrame
		r.Overlay.LastFlip.Store(&frame)
		r.Overlay.FrameCount.Add(1)

		// TODO: gpu.GlobalLiverpool.Walk()
		r.FramebufferMutex.RLock()
		framebuffer, ok := r.Framebuffers[frame.GpuAddress]
		r.FramebufferMutex.RUnlock()

		// Detile from guest memory and copy to staging buffer.
		if ok && r.FramebufferTexture != nil {
			texture := framebuffer.Snapshot()
			r.FramebufferTexture.WritePixels(texture.Pix)
		}
	}
}

func (r *Renderer) RegisterFramebuffer(address uintptr, attribute *VideoOutBufferAttribute) {
	framebuffer := NewGuestFramebuffer(
		address, int(attribute.Width), int(attribute.Height),
		int(attribute.PitchInPixel), int(attribute.TilingMode),
	)
	r.FramebufferMutex.Lock()
	r.Framebuffers[address] = framebuffer
	r.FramebufferMutex.Unlock()
	r.Overlay.LatestFramebuffer.Store(framebuffer)
	if r.FramebufferTexture == nil {
		var err error
		r.FramebufferTexture, err = NewFramebufferTexture(&r.Handles, r.Backend, attribute.Width, attribute.Height)
		if err != nil {
			panic("renderer: could not allocate framebuffer texture: " + err.Error())
		}
	}
}

func (r *Renderer) prepareFramebuffers() {
	swapchainImageResources := r.Handles.Context.SwapchainImageResources()
	for _, res := range swapchainImageResources {
		var framebuffer vk.Framebuffer
		result := vk.CreateFramebuffer(r.Handles.Device, &vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      r.RenderPass,
			AttachmentCount: 2,
			PAttachments:    []vk.ImageView{res.View(), r.Depth.view},
			Width:           r.SwapchainDimensions.Width,
			Height:          r.SwapchainDimensions.Height,
			Layers:          1,
		}, nil, &framebuffer)
		if err := as.NewError(result); err != nil {
			panic(err)
		}
		res.SetFramebuffer(framebuffer)
	}
}

func (r *Renderer) preparePipelineCache() {
	var pipelineCache vk.PipelineCache
	vk.CreatePipelineCache(r.Handles.Device, &vk.PipelineCacheCreateInfo{
		SType: vk.StructureTypePipelineCacheCreateInfo,
	}, nil, &pipelineCache)
	r.PipelineCache = pipelineCache
}

func (r *Renderer) prepareRenderPass() {
	var renderPass vk.RenderPass
	result := vk.CreateRenderPass(r.Handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 2,
		PAttachments: []vk.AttachmentDescription{{
			Format:         r.SwapchainDimensions.Format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutPresentSrc,
		}, {
			Format:         r.Depth.format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpDontCare,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
		}},
		SubpassCount: 1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
			PDepthStencilAttachment: &vk.AttachmentReference{
				Attachment: 1,
				Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
			},
		}},
	}, nil, &renderPass)
	if err := as.NewError(result); err != nil {
		panic(err)
	}
	r.RenderPass = renderPass
}
