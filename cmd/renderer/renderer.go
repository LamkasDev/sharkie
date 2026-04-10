package renderer

import (
	"runtime"
	"time"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
	vk "github.com/goki/vulkan"
)

type Renderer struct {
	Handles       VulkanHandles
	Backend       backend.Backend[glfwvulkanbackend.GLFWWindowFlags]
	GpuTranslator *GpuTranslator
	FrameSource   *FrameSource
	Overlay       *ImguiOverlay

	SwapchainDimensions   *as.SwapchainDimensions
	Depth                 *Depth
	RenderPass            vk.RenderPass
	PipelineCache         vk.PipelineCache
	PendingCommandBuffers chan vk.CommandBuffer

	DisplayTextureId imgui.TextureRef
}

func NewRenderer(context as.Context, dimensions *as.SwapchainDimensions) *Renderer {
	r := &Renderer{
		Handles:               NewVulkanHandles(context),
		SwapchainDimensions:   dimensions,
		FrameSource:           NewFrameSource(),
		PendingCommandBuffers: make(chan vk.CommandBuffer),
	}

	var err error
	if r.Backend, err = backend.CreateBackend(glfwvulkanbackend.NewGLFWBackend()); err != nil {
		panic(err)
	}
	if r.GpuTranslator, err = NewGpuTranslator(r.Handles, r.Backend); err != nil {
		panic(err)
	}

	r.Depth = NewDepth(r)
	r.prepareRenderPass()
	r.preparePipelineCache()
	r.prepareFramebuffers()
	for _, res := range r.Handles.Context.SwapchainImageResources() {
		vk.BeginCommandBuffer(res.CommandBuffer(), &vk.CommandBufferBeginInfo{
			SType: vk.StructureTypeCommandBufferBeginInfo,
			Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageSimultaneousUseBit),
		})
		vk.EndCommandBuffer(res.CommandBuffer())
	}

	return r
}

func (r *Renderer) Destroy() {
	vk.DeviceWaitIdle(r.Handles.Device)
	r.Backend.Cleanup()
	if r.GpuTranslator != nil {
		r.GpuTranslator.Destroy()
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
		if r.DisplayTextureId.CData != nil {
			imgui.Image(r.DisplayTextureId, imgui.Vec2{
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

	for range r.FrameSource.Channel {
		r.UpdateCounters()

		gpu.GlobalLiverpool.Walk()
		draws := gpu.GlobalLiverpool.FlushDrawCalls()
		if len(draws) > 0 && r.GpuTranslator != nil {
			commandBuffer := r.GpuTranslator.Translate(draws)
			if commandBuffer == nil {
				continue
			}
			r.PendingCommandBuffers <- *commandBuffer
		}
	}
}

func (r *Renderer) UpdateCounters() {
	r.Overlay.FrameCount.Add(1)
	now := time.Now().UnixNano()
	last := r.Overlay.FrameLastTime.Swap(now)
	delta := float64(now-last) / float64(time.Second)
	if delta <= 0 {
		return
	}
	instantFramerate := 1.0 / delta
	alpha := 0.1
	oldFramerate := r.Overlay.Framerate.Load()
	newFramerate := (instantFramerate * alpha) + (oldFramerate * (1.0 - alpha))
	r.Overlay.Framerate.Store(newFramerate)
}

func (r *Renderer) RegisterFramebuffer(address uintptr, attribute *VideoOutBufferAttribute) {
	if r.GpuTranslator == nil {
		return
	}
	textureId, err := r.GpuTranslator.RegisterSurface(address, attribute.Width, attribute.Height)
	if err != nil {
		panic("renderer: could not register GPU surface: " + err.Error())
	}

	if r.DisplayTextureId.CData == nil && textureId.CData != nil {
		r.DisplayTextureId = textureId
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
			Format:         vk.Format(r.SwapchainDimensions.Format),
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
