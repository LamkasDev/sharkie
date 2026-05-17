package vulkan

import (
	"fmt"

	as "github.com/LamkasDev/asche"
	vk "github.com/goki/vulkan"
)

type VulkanFramebuffer struct {
	Framebuffer       vk.Framebuffer
	RenderPass        vk.RenderPass
	RenderPassNoClear vk.RenderPass
}

type FramebufferRequest struct {
	ImageView vk.ImageView

	FramebufferKey
}

type FramebufferKey struct {
	GpuAddress uintptr
	Format     vk.Format
	Width      uint32
	Height     uint32
}

func (t *GpuTranslator) createFramebuffer(request FramebufferRequest) (*VulkanFramebuffer, error) {
	fb := &VulkanFramebuffer{}

	// Create render pass with clear.
	var renderPass vk.RenderPass
	result := vk.CreateRenderPass(t.handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments: []vk.AttachmentDescription{{
			Format:         request.Format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutGeneral,
		}},
		SubpassCount: 1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
		}},
	}, nil, &renderPass)
	if err := as.NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateRenderPass: %w", err)
	}
	fb.RenderPass = renderPass

	// Create render pass without clear.
	var renderPassNoClear vk.RenderPass
	result = vk.CreateRenderPass(t.handles.Device, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments: []vk.AttachmentDescription{{
			Format:         request.Format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpLoad,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutGeneral,
			FinalLayout:    vk.ImageLayoutGeneral,
		}},
		SubpassCount: 1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
		}},
	}, nil, &renderPassNoClear)
	if err := as.NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateRenderPass (no clear): %w", err)
	}
	fb.RenderPassNoClear = renderPassNoClear

	// Create framebuffer.
	var framebuffer vk.Framebuffer
	result = vk.CreateFramebuffer(t.handles.Device, &vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		RenderPass:      fb.RenderPass,
		AttachmentCount: 1,
		PAttachments:    []vk.ImageView{request.ImageView},
		Width:           request.Width,
		Height:          request.Height,
		Layers:          1,
	}, nil, &framebuffer)
	if err := as.NewError(result); err != nil {
		return nil, fmt.Errorf("vkCreateFramebuffer: %w", err)
	}
	fb.Framebuffer = framebuffer

	return fb, nil
}
