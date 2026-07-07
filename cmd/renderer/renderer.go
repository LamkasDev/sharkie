package renderer

import (
	"runtime"
	"sync"
	"time"

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	"github.com/LamkasDev/sharkie/cmd/vulkan/translation"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

type Renderer struct {
	Handles       vulkan.VulkanHandles
	Backend       backend.Backend[glfwvulkanbackend.GLFWWindowFlags]
	GpuTranslator *translation.GpuTranslator
	FrameSource   *FrameSource
	Overlay       *ImguiOverlay

	SwapchainDimensions *backend.SwapchainDimensions
	Depth               *Depth
	RenderPass          vk.RenderPass
	PipelineCache       vk.PipelineCache

	QueueMutex sync.Mutex
	FrameReady chan struct{}

	DisplayTextureId imgui.TextureRef
}

func NewRenderer(context *vulkan.VulkanContext, dimensions *backend.SwapchainDimensions) *Renderer {
	r := &Renderer{
		SwapchainDimensions: dimensions,
		FrameSource:         NewFrameSource(),
		QueueMutex:          sync.Mutex{},
	}
	r.Handles = vulkan.NewVulkanHandles(context, &r.QueueMutex)

	var err error
	if r.Backend, err = backend.CreateBackend(glfwvulkanbackend.NewGLFWBackend()); err != nil {
		panic(err)
	}
	if r.GpuTranslator, err = translation.NewGpuTranslator(r.Handles, r.Backend); err != nil {
		panic(err)
	}
	r.FrameSource.OnSubmit = r.GpuTranslator.ResetFence

	r.Depth = NewDepth(r)
	r.prepareRenderPass()
	r.preparePipelineCache()
	r.prepareFramebuffers()
	for _, res := range r.Handles.Context.SwapchainImageResources {
		vk.BeginCommandBuffer(res.Cmd, &vk.CommandBufferBeginInfo{
			SType: vk.StructureTypeCommandBufferBeginInfo,
			Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageSimultaneousUseBit),
		})
		vk.EndCommandBuffer(res.Cmd)
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

	for frame := range r.FrameSource.Channel {
		logger.Printf("[%s] retrieved from channel.\n",
			color.Blue.Sprintf("Frame %d", frame.Number),
		)
		r.UpdateCounters()

		// Fetch PM4 command stream.
		gpu.GlobalLiverpool.FrameNumber = frame.Number
		gpu.GlobalLiverpool.Walk()
		stream := gpu.GlobalLiverpool.FlushStream()

		// Translate PM4 command stream into VkCommandBuffer.
		r.GpuTranslator.ResetFrameState(frame.Number)
		commandBuffer := r.GpuTranslator.Translate(frame.Number, &stream)

		// Submit command buffer.
		r.GpuTranslator.ResetWorkerFence()
		r.QueueMutex.Lock()
		result := vk.QueueSubmit(r.Handles.GraphicsQueue, 1, []vk.SubmitInfo{{
			SType:              vk.StructureTypeSubmitInfo,
			CommandBufferCount: 1,
			PCommandBuffers:    []vk.CommandBuffer{*commandBuffer},
		}}, r.GpuTranslator.GetWorkerFence())
		r.QueueMutex.Unlock()
		if err := vulkan.NewError(result); err != nil {
			panic(err)
		}

		// Wait for it to finish.
		r.GpuTranslator.WaitOnWorkerFence()
		r.GpuTranslator.FlushDeferredDestruction()
		r.Handles.FreeCommandBuffer(*commandBuffer)

		// Transition surface and update texture ID for display.
		surface := r.GpuTranslator.GetSurfaceByAddress(frame.GpuAddress)
		if surface != nil {
			err := vulkan.RunWithCommandBuffer(&r.Handles, r.GpuTranslator.GetWorkerFence(), func(commandBuffer vk.CommandBuffer) {
				surface.ImageView.Image.BarrierGeneralShaderAccess(commandBuffer)
			})
			if err != nil {
				panic(err)
			}
			r.QueueMutex.Lock()
			r.DisplayTextureId = r.GpuTranslator.GetSurfaceTexture(surface)
			r.QueueMutex.Unlock()
		}
		r.GpuTranslator.SignalFence()

		// Wait on next frame.
		select {
		case r.FrameReady <- struct{}{}:
		default:
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

func (r *Renderer) WaitOnFence() {
	r.GpuTranslator.WaitOnFence()
}

func (r *Renderer) RegisterFramebuffer(address uintptr, attribute *VideoOutBufferAttribute) {
	// Video-out registration only records the guest display buffers.
	// Surfaces are created on the first BindPipeline with CB_COLOR0_SLICE height
	// (1088 for 1080p), not the 1080 reported by sceVideoOutSetBufferAttribute.
	_ = address
	_ = attribute
}

func (r *Renderer) prepareFramebuffers() {
	swapchainImageResources := r.Handles.Context.SwapchainImageResources
	for _, res := range swapchainImageResources {
		var framebuffer vk.Framebuffer
		result := vk.CreateFramebuffer(r.Handles.Device, &vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      r.RenderPass,
			AttachmentCount: 2,
			PAttachments:    []vk.ImageView{res.View, r.Depth.view},
			Width:           r.SwapchainDimensions.Width,
			Height:          r.SwapchainDimensions.Height,
			Layers:          1,
		}, nil, &framebuffer)
		if err := vulkan.NewError(result); err != nil {
			panic(err)
		}
		res.Framebuffer = framebuffer
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
	if err := vulkan.NewError(result); err != nil {
		panic(err)
	}
	r.RenderPass = renderPass
}
