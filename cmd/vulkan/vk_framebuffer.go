package vulkan

import (
	"fmt"

	vk "github.com/goki/vulkan"
)

type VulkanFramebuffer struct {
	Framebuffer                   vk.Framebuffer
	RenderPass                    vk.RenderPass
	RenderPassNoClear             vk.RenderPass
	RenderPassLoadColorClearDepth vk.RenderPass
	RenderPassClearColorLoadDepth vk.RenderPass
}

type FramebufferRequest struct {
	ImageView      *VulkanImageView
	DepthImageView *VulkanImageView

	FramebufferKey
}

type FramebufferKey struct {
	GpuAddress      uintptr
	DepthGpuAddress uintptr

	Format      vk.Format
	DepthFormat vk.Format
	Width       uint32
	Height      uint32
}

func CreateFramebuffer(handles *VulkanHandles, request FramebufferRequest) (*VulkanFramebuffer, error) {
	fb := &VulkanFramebuffer{}

	// Attachments.
	attachments := []vk.AttachmentDescription{{
		Format:         request.Format,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutGeneral,
	}}
	attachmentsNoClear := []vk.AttachmentDescription{{
		Format:         request.Format,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpLoad,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutGeneral,
		FinalLayout:    vk.ImageLayoutGeneral,
	}}

	var depthAttachmentRef *vk.AttachmentReference
	if request.DepthFormat != vk.FormatUndefined {
		depthAttachment := vk.AttachmentDescription{
			Format:         request.DepthFormat,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpClear,
			StencilStoreOp: vk.AttachmentStoreOpStore,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
		}
		depthAttachmentNoClear := vk.AttachmentDescription{
			Format:         request.DepthFormat,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpLoad,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpLoad,
			StencilStoreOp: vk.AttachmentStoreOpStore,
			InitialLayout:  vk.ImageLayoutDepthStencilAttachmentOptimal,
			FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
		}
		depthAttachmentRef = &vk.AttachmentReference{
			Attachment: uint32(len(attachments)),
			Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
		}
		attachments = append(attachments, depthAttachment)
		attachmentsNoClear = append(attachmentsNoClear, depthAttachmentNoClear)
	}

	colorAttachments := make([]vk.AttachmentReference, 8)
	colorAttachments[0] = vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}
	for i := 1; i < 8; i++ {
		colorAttachments[i] = vk.AttachmentReference{
			Attachment: vk.AttachmentUnused,
			Layout:     vk.ImageLayoutUndefined,
		}
	}

	// Create render pass with clear.
	var renderPass vk.RenderPass
	result := vk.CreateRenderPass(handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: uint32(len(attachments)),
		PAttachments:    attachments,
		SubpassCount:    1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:       vk.PipelineBindPointGraphics,
			ColorAttachmentCount:    8,
			PColorAttachments:       colorAttachments,
			PDepthStencilAttachment: depthAttachmentRef,
		}},
	}, nil, &renderPass)
	if err := NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateRenderPass: %w", err)
	}
	fb.RenderPass = renderPass

	// Create render pass without clear.
	var renderPassNoClear vk.RenderPass
	result = vk.CreateRenderPass(handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: uint32(len(attachmentsNoClear)),
		PAttachments:    attachmentsNoClear,
		SubpassCount:    1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:       vk.PipelineBindPointGraphics,
			ColorAttachmentCount:    8,
			PColorAttachments:       colorAttachments,
			PDepthStencilAttachment: depthAttachmentRef,
		}},
	}, nil, &renderPassNoClear)
	if err := NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateRenderPass (no clear): %w", err)
	}
	fb.RenderPassNoClear = renderPassNoClear

	if request.DepthFormat != vk.FormatUndefined {
		attachmentsLoadColorClearDepth := []vk.AttachmentDescription{
			attachmentsNoClear[0],
			attachments[1],
		}
		var renderPassLoadColorClearDepth vk.RenderPass
		result = vk.CreateRenderPass(handles.Device, &vk.RenderPassCreateInfo{
			SType:           vk.StructureTypeRenderPassCreateInfo,
			AttachmentCount: uint32(len(attachmentsLoadColorClearDepth)),
			PAttachments:    attachmentsLoadColorClearDepth,
			SubpassCount:    1,
			PSubpasses: []vk.SubpassDescription{{
				PipelineBindPoint:       vk.PipelineBindPointGraphics,
				ColorAttachmentCount:    8,
				PColorAttachments:       colorAttachments,
				PDepthStencilAttachment: depthAttachmentRef,
			}},
		}, nil, &renderPassLoadColorClearDepth)
		if err := NewError(result); err != nil {
			return nil, fmt.Errorf("vkCreateRenderPass (load color clear depth): %w", err)
		}
		fb.RenderPassLoadColorClearDepth = renderPassLoadColorClearDepth

		attachmentsClearColorLoadDepth := []vk.AttachmentDescription{
			attachments[0],
			attachmentsNoClear[1],
		}
		var renderPassClearColorLoadDepth vk.RenderPass
		result = vk.CreateRenderPass(handles.Device, &vk.RenderPassCreateInfo{
			SType:           vk.StructureTypeRenderPassCreateInfo,
			AttachmentCount: uint32(len(attachmentsClearColorLoadDepth)),
			PAttachments:    attachmentsClearColorLoadDepth,
			SubpassCount:    1,
			PSubpasses: []vk.SubpassDescription{{
				PipelineBindPoint:       vk.PipelineBindPointGraphics,
				ColorAttachmentCount:    8,
				PColorAttachments:       colorAttachments,
				PDepthStencilAttachment: depthAttachmentRef,
			}},
		}, nil, &renderPassClearColorLoadDepth)
		if err := NewError(result); err != nil {
			return nil, fmt.Errorf("vkCreateRenderPass (clear color load depth): %w", err)
		}
		fb.RenderPassClearColorLoadDepth = renderPassClearColorLoadDepth
	}

	// Create framebuffer.
	views := []vk.ImageView{request.ImageView.ImageView}
	if request.DepthImageView != nil {
		views = append(views, request.DepthImageView.ImageView)
	}
	var framebuffer vk.Framebuffer
	result = vk.CreateFramebuffer(handles.Device, &vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		RenderPass:      fb.RenderPass,
		AttachmentCount: uint32(len(views)),
		PAttachments:    views,
		Width:           request.Width,
		Height:          request.Height,
		Layers:          1,
	}, nil, &framebuffer)
	if err := NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateFramebuffer: %w", err)
	}
	fb.Framebuffer = framebuffer

	return fb, nil
}

func (fb *VulkanFramebuffer) Destroy(device vk.Device) {
	if fb == nil {
		return
	}
	if fb.Framebuffer != vk.NullFramebuffer {
		vk.DestroyFramebuffer(device, fb.Framebuffer, nil)
		fb.Framebuffer = vk.NullFramebuffer
	}
	destroyRenderPass := func(rp vk.RenderPass) {
		if rp != vk.NullRenderPass {
			vk.DestroyRenderPass(device, rp, nil)
		}
	}
	destroyRenderPass(fb.RenderPass)
	destroyRenderPass(fb.RenderPassNoClear)
	destroyRenderPass(fb.RenderPassLoadColorClearDepth)
	destroyRenderPass(fb.RenderPassClearColorLoadDepth)
}
