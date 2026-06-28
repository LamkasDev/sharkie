package vulkan

import (
	"fmt"

	vk "github.com/goki/vulkan"
)

type VulkanFramebuffer struct {
	Framebuffer       vk.Framebuffer
	RenderPass        vk.RenderPass
	RenderPassNoClear vk.RenderPass
}

type FramebufferRequest struct {
	ImageView      vk.ImageView
	DepthImageView vk.ImageView

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

func (t *GpuTranslator) createFramebuffer(request FramebufferRequest) (*VulkanFramebuffer, error) {
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
			FinalLayout:    vk.ImageLayoutGeneral,
		}
		depthAttachmentNoClear := vk.AttachmentDescription{
			Format:         request.DepthFormat,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpLoad,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpLoad,
			StencilStoreOp: vk.AttachmentStoreOpStore,
			InitialLayout:  vk.ImageLayoutGeneral,
			FinalLayout:    vk.ImageLayoutGeneral,
		}
		depthAttachmentRef = &vk.AttachmentReference{
			Attachment: uint32(len(attachments)),
			Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
		}
		attachments = append(attachments, depthAttachment)
		attachmentsNoClear = append(attachmentsNoClear, depthAttachmentNoClear)
	}

	// Create render pass with clear.
	var renderPass vk.RenderPass
	result := vk.CreateRenderPass(t.handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: uint32(len(attachments)),
		PAttachments:    attachments,
		SubpassCount:    1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
			PDepthStencilAttachment: depthAttachmentRef,
		}},
	}, nil, &renderPass)
	if err := NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateRenderPass: %w", err)
	}
	fb.RenderPass = renderPass

	// Create render pass without clear.
	var renderPassNoClear vk.RenderPass
	result = vk.CreateRenderPass(t.handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: uint32(len(attachmentsNoClear)),
		PAttachments:    attachmentsNoClear,
		SubpassCount:    1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
			PDepthStencilAttachment: depthAttachmentRef,
		}},
	}, nil, &renderPassNoClear)
	if err := NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateRenderPass (no clear): %w", err)
	}
	fb.RenderPassNoClear = renderPassNoClear

	// Create framebuffer.
	views := []vk.ImageView{request.ImageView}
	if request.DepthImageView != vk.NullImageView {
		views = append(views, request.DepthImageView)
	}
	var framebuffer vk.Framebuffer
	result = vk.CreateFramebuffer(t.handles.Device, &vk.FramebufferCreateInfo{
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
