package renderer

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/irq"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	"github.com/LamkasDev/sharkie/cmd/vulkan/translation"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

type Renderer struct {
	Handles        *vulkan.VulkanHandles
	Backend        backend.Backend[glfwvulkanbackend.GLFWWindowFlags]
	GpuTranslator  *translation.GpuTranslator
	RingWorkSource *RingWorkSource
	FrameSource    *FlipSource
	Overlay        *ImguiOverlay

	SwapchainDimensions *backend.SwapchainDimensions
	Depth               *Depth
	RenderPass          vk.RenderPass
	PipelineCache       vk.PipelineCache

	FrameReady chan struct{}

	DisplayTextureId       imgui.TextureRef
	DisplayTextureModified bool
	DrawUI                 func()
	FramebufferOffsetY     float32
}

func NewRenderer(context *vulkan.VulkanContext, dimensions *backend.SwapchainDimensions) *Renderer {
	r := &Renderer{
		SwapchainDimensions: dimensions,
		RingWorkSource:      NewRingWorkSource(),
		FrameSource:         NewFlipSource(),
		FrameReady:          make(chan struct{}, 1),
	}
	r.Handles = vulkan.NewVulkanHandles(context)

	var err error
	if r.Backend, err = backend.CreateBackend(glfwvulkanbackend.NewGLFWBackend()); err != nil {
		panic(err)
	}
	if r.GpuTranslator, err = translation.NewGpuTranslator(r.Handles, r.Backend); err != nil {
		panic(err)
	}
	r.RingWorkSource.OnSubmit = r.GpuTranslator.ResetFence

	r.Depth = NewDepth(r)
	r.prepareRenderPass()
	r.preparePipelineCache()
	r.RecreateSwapchain()

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
	if r.DrawUI != nil {
		r.DrawUI()
	}
}

func (r *Renderer) DrawFramebuffer() {
	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: r.FramebufferOffsetY})
	imgui.SetNextWindowSize(imgui.Vec2{X: float32(r.SwapchainDimensions.Width), Y: float32(r.SwapchainDimensions.Height) - r.FramebufferOffsetY})
	imgui.PushStyleColorVec4(imgui.ColWindowBg, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 1.0})
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	if imgui.BeginV("##fb", nil, ImguiOverlayFlags|imgui.WindowFlagsNoBringToFrontOnFocus) {
		texId := r.DisplayTextureId
		if texId.CData != nil {
			imgui.Image(texId, imgui.Vec2{
				X: float32(r.SwapchainDimensions.Width),
				Y: float32(r.SwapchainDimensions.Height) - r.FramebufferOffsetY,
			})
		}
		imgui.End()
	}
	imgui.PopStyleVar()
	imgui.PopStyleColor()
}

func (r *Renderer) ConsumeRingWork(done chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(done)

	for ringWork := range r.RingWorkSource.Channel {
		logger.Printf("[%s] retrieved ring work from channel.\n",
			color.Blue.Sprintf("Frame %d", ringWork.Number),
		)

		// Fetch PM4 command streams.
		gpu.GlobalLiverpool.FrameNumber = ringWork.Number
		streams := gpu.GlobalLiverpool.Walk(ringWork.RingWork)

		// Translate command streams.
		r.GpuTranslator.ResetFrameState(ringWork.Number)
		if IsDumpRequested() {
			r.GpuTranslator.EnableDebugDump(ringWork.Number)
			DumpPending.Store(true)
			DumpRequested.Store(false)
		}
		r.GpuTranslator.StartCommandBuffer()
		r.GpuTranslator.BeforeTranslate()
		for _, stream := range streams {
			r.GpuTranslator.UpdateUserDataBuffers(stream)
			r.GpuTranslator.Translate(ringWork.Number, stream)
		}

		// Submit command buffer.
		r.GpuTranslator.EndCommandBuffer()
		r.GpuTranslator.SubmitCommandBuffers()
		r.Handles.FlushDeferredDestruction()

		// Signal that we're done.
		irq.GlobalInterruptHandler.Signal(irq.InterruptIdGpuIdle)
		r.GpuTranslator.SignalFence()
	}
}

func (r *Renderer) ConsumeFlips(done chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(done)

	for frame := range r.FrameSource.Channel {
		logger.Printf("[%s] retrieved flip from channel.\n",
			color.Blue.Sprintf("Frame %d", frame.Number),
		)
		r.UpdateCounters()

		// Get surface.
		var err error
		surface := r.GpuTranslator.GetSurfaceByAddress(frame.Flip.GpuAddress)
		if surface == nil {
			group := r.GpuTranslator.GetImageByAddress(frame.Flip.GpuAddress)
			if group == nil {
				logger.Printf("[%s] failed to find surface image.\n",
					color.Blue.Sprintf("Frame %d", frame.Number),
				)
			} else {
				surface, err = r.GpuTranslator.GetSurface(group.LeadingImage.FirstDescriptor, 0)
				if err != nil {
					panic(err)
				}
			}
		}

		// Transition surface and update texture ID for display.
		if surface != nil {
			var dumpedStaging *vulkan.VulkanStagingBuffer
			var dumpSize vk.DeviceSize
			var dumpWidth, dumpHeight int
			var dumpFormat vk.Format

			shouldDump := DumpPending.Load()
			if shouldDump && surface.ImageView != nil && surface.ImageView.Image != nil {
				dumpWidth = int(surface.ImageView.Image.FirstDescriptor.Width)
				dumpHeight = int(surface.ImageView.Image.FirstDescriptor.Height)
				dumpFormat = surface.ImageView.Image.ImageFormat
				if dumpWidth > 0 && dumpHeight > 0 {
					dumpSize = vk.DeviceSize(dumpWidth * dumpHeight * 4)
					dumpedStaging, _ = r.Handles.StagingBufferPool.Get(r.Handles, dumpSize)
				}
			}

			err = vulkan.RunWithCommandBuffer(r.Handles.GraphicsQueue, r.Handles, func(commandBuffer *vulkan.VulkanCommandBuffer) {
				if surface.ImageView.Image.ShouldUploadToVkImage(frame.Number) {
					err = surface.ImageView.Image.UploadToVkImage(r.Handles, commandBuffer, r.GpuTranslator.GetLinearBuffer, frame.Number)
					if err != nil {
						panic(err)
					}
				}
				if dumpedStaging != nil {
					surface.ImageView.Image.BarrierTransferSrc(commandBuffer)
					vk.CmdCopyImageToBuffer(commandBuffer.CommandBuffer, surface.ImageView.Image.Image, vk.ImageLayoutTransferSrcOptimal, dumpedStaging.Buffer, 1, []vk.BufferImageCopy{{
						BufferOffset:      0,
						BufferRowLength:   0,
						BufferImageHeight: 0,
						ImageSubresource: vk.ImageSubresourceLayers{
							AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
							MipLevel:       0,
							BaseArrayLayer: 0,
							LayerCount:     1,
						},
						ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
						ImageExtent: vk.Extent3D{Width: uint32(dumpWidth), Height: uint32(dumpHeight), Depth: 1},
					}})
					vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
						vk.PipelineStageFlags(vk.PipelineStageTransferBit),
						vk.PipelineStageFlags(vk.PipelineStageHostBit),
						0, 1, []vk.MemoryBarrier{{
							SType:         vk.StructureTypeMemoryBarrier,
							SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
							DstAccessMask: vk.AccessFlags(vk.AccessHostReadBit),
						}}, 0, nil, 0, nil)
				}
				surface.ImageView.Image.BarrierGeneralShaderAccess(commandBuffer)
			}, frame.Number)
			if err != nil {
				panic(err)
			}
			r.DisplayTextureId = r.GpuTranslator.GetSurfaceTexture(surface)

			if dumpedStaging != nil {
				r.saveDumpPNG(dumpedStaging, dumpSize, dumpWidth, dumpHeight, dumpFormat, frame.Number)
				r.Handles.StagingBufferPool.Put(dumpedStaging)
				report, _ := r.GpuTranslator.GetAndClearDebugDump()
				if report != "" {
					txtPath, err := SaveDebugText(report, frame.Number)
					if err == nil {
						logger.Printf("[F9] Saved frame %d debug info to %s\n", frame.Number, txtPath)
					}
				}
				DumpPending.Store(false)
			}
		}

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

func (r *Renderer) RegisterFramebuffer(address uintptr, attribute *VideoOutBufferAttribute) {
	// Video-out registration only records the guest display buffers.
	// Surfaces are created on the first BindPipeline with CB_COLOR0_SLICE height
	// (1088 for 1080p), not the 1080 reported by sceVideoOutSetBufferAttribute.
	_ = address
	_ = attribute
}

func (r *Renderer) RecreateSwapchain() {
	vk.DeviceWaitIdle(r.Handles.Device)
	r.SwapchainDimensions = r.Handles.Context.SwapchainDimensions

	// Destroy old framebuffers.
	for _, res := range r.Handles.Context.SwapchainImageResources {
		if res.Framebuffer != vk.NullFramebuffer {
			vk.DestroyFramebuffer(r.Handles.Device, res.Framebuffer, nil)
			res.Framebuffer = vk.NullFramebuffer
		}
	}

	// Recreate depth buffer.
	if r.Depth != nil {
		r.Depth.Destroy(r)
	}
	r.Depth = NewDepth(r)

	// Recreate framebuffers.
	r.prepareFramebuffers()
	for _, res := range r.Handles.Context.SwapchainImageResources {
		vk.BeginCommandBuffer(res.Cmd, &vk.CommandBufferBeginInfo{
			SType: vk.StructureTypeCommandBufferBeginInfo,
			Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageSimultaneousUseBit),
		})
		vk.EndCommandBuffer(res.Cmd)
	}
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

func (r *Renderer) saveDumpPNG(staging *vulkan.VulkanStagingBuffer, size vk.DeviceSize, width, height int, format vk.Format, frameNumber uint64) {
	data := r.Handles.MapMemory(staging.Memory, size)
	defer vk.UnmapMemory(r.Handles.Device, staging.Memory)

	rgbaImg := image.NewRGBA(image.Rect(0, 0, width, height))
	isBGRA := (format == vk.FormatB8g8r8a8Unorm || format == vk.FormatB8g8r8a8Srgb)
	for y := 0; y < height; y++ {
		rowSrc := data[y*width*4 : (y+1)*width*4]
		rowDst := rgbaImg.Pix[y*rgbaImg.Stride : y*rgbaImg.Stride+width*4]
		if isBGRA {
			for x := 0; x < width; x++ {
				b := rowSrc[x*4+0]
				g := rowSrc[x*4+1]
				r := rowSrc[x*4+2]
				a := rowSrc[x*4+3]
				rowDst[x*4+0] = r
				rowDst[x*4+1] = g
				rowDst[x*4+2] = b
				rowDst[x*4+3] = a
			}
		} else {
			copy(rowDst, rowSrc)
		}
	}

	dumpDir := "/home/cute-foxgirls/.cache/sharkie/dump"
	os.MkdirAll(dumpDir, 0755)
	filePath := filepath.Join(dumpDir, fmt.Sprintf("frame_%04d.png", frameNumber))
	f, err := os.Create(filePath)
	if err == nil {
		defer f.Close()
		png.Encode(f, rgbaImg)
		logger.Printf("Dumped frame %d to %s\n", frameNumber, filePath)
	}
}
