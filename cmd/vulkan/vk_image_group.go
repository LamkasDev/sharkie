package vulkan

import (
	"sync"

	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
)

type VulkanImageGroup struct {
	Address      uintptr
	GuestSize    uintptr
	Images       map[uint64]*VulkanImage
	LeadingImage *VulkanImage

	SyncLock sync.Mutex
}

func NewVulkanImageGroup(address uintptr, guestSize uintptr) *VulkanImageGroup {
	return &VulkanImageGroup{
		Address:   address,
		GuestSize: guestSize,
		Images:    make(map[uint64]*VulkanImage),
	}
}

func (g *VulkanImageGroup) GetImage(handles *VulkanHandles, descriptor spirvStructs.ImageDescriptor, format vk.Format, isSurface bool, commandBuffer *VulkanCommandBuffer, frame uint64, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error)) (*VulkanImage, error, bool) {
	g.SyncLock.Lock()
	defer g.SyncLock.Unlock()

	// Check if we have existing fitting image.
	hash := descriptor.ViewHash()
	image, ok := g.Images[hash]
	if ok {
		if isSurface {
			image.IsSurface = true
		}
		g.syncImageFromLeading(handles, image, commandBuffer, frame, getLinearBuffer)
		g.LeadingImage = image

		return image, nil, false
	}

	// Create new image.
	var err error
	image, err = CreateImage(handles, VulkanImageRequest{
		Descriptor: descriptor,
		Format:     format,
		IsSurface:  isSurface,
	}, commandBuffer, frame)
	if err != nil {
		return nil, err, false
	}
	g.Images[hash] = image
	g.syncImageFromLeading(handles, image, commandBuffer, frame, getLinearBuffer)
	g.LeadingImage = image

	// Expand guest size (we update memory tracking in gpu_translator_image.go).
	if image.GuestSize > g.GuestSize {
		g.GuestSize = image.GuestSize
	}

	return image, nil, true
}

func (g *VulkanImageGroup) syncImageFromLeading(handles *VulkanHandles, image *VulkanImage, commandBuffer *VulkanCommandBuffer, frame uint64, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error)) {
	// Check if we need to copy.
	if g.LeadingImage == nil || g.LeadingImage == image {
		// First time upload.
		if g.LeadingImage == nil {
			image.MarkCpuModified(frame)
			if image.ShouldUploadToVkImage(frame) {
				if err := image.UploadToVkImage(handles, commandBuffer, getLinearBuffer, frame); err != nil {
					logger.Printf("failed to upload image: %v\n", err)
				}
			}
		}
		return
	}

	// Fast GPU copy if compatible.
	// Vulkan cmd copy rules (same block size, same extent, compatible features, etc.).
	_, srcBpp := gcn.TranslateGcnFormat(g.LeadingImage.FirstDescriptor.DataFormat, g.LeadingImage.FirstDescriptor.NumFormat, 0)
	_, dstBpp := gcn.TranslateGcnFormat(image.FirstDescriptor.DataFormat, image.FirstDescriptor.NumFormat, 0)
	isCompatible := srcBpp == dstBpp
	if isCompatible {
		err := g.LeadingImage.CopyToImage(handles, commandBuffer, image, frame)
		if err != nil {
			logger.Printf("failed to copy image on GPU: %v\n", err)
		}

		// Assume we successfully synced all state from leading image.
		image.SyncFlags = g.LeadingImage.SyncFlags
	} else {
		// Incompatible formats/extents; do CPU roundtrip (evict basically).
		if g.LeadingImage.ShouldDownloadFromVkImage() {
			if err := g.LeadingImage.DownloadFromVkImage(handles, commandBuffer, getLinearBuffer, frame); err != nil {
				logger.Printf("failed to download leading image: %v\n", err)
			}
		}

		image.MarkCpuModified(frame)
		if image.ShouldUploadToVkImage(frame) {
			if err := image.UploadToVkImage(handles, commandBuffer, getLinearBuffer, frame); err != nil {
				logger.Printf("failed to upload image: %v\n", err)
			}
		}
	}
}

func (g *VulkanImageGroup) Destroy(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) {
	g.SyncLock.Lock()
	defer g.SyncLock.Unlock()

	if g.LeadingImage.ShouldDownloadFromVkImage() {
		if err := g.LeadingImage.DownloadFromVkImage(handles, commandBuffer, getLinearBuffer, frame); err != nil {
			logger.Printf("failed to download leading image: %v\n", err)
		}
	}

	for _, image := range g.Images {
		handles.DeferDestroyImage(image)
	}
	g.Images = nil
	g.LeadingImage = nil
}

func (g *VulkanImageGroup) ShouldUploadToVkImage(frame uint64) bool {
	g.SyncLock.Lock()
	defer g.SyncLock.Unlock()
	return g.LeadingImage.ShouldUploadToVkImage(frame)
}

func (g *VulkanImageGroup) MarkCpuModified(frame uint64) {
	g.SyncLock.Lock()
	defer g.SyncLock.Unlock()
	g.LeadingImage.MarkCpuModified(frame)
}

func (g *VulkanImageGroup) UploadToVkImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) error {
	g.SyncLock.Lock()
	defer g.SyncLock.Unlock()
	return g.LeadingImage.UploadToVkImage(handles, commandBuffer, getLinearBuffer, frame)
}

func (g *VulkanImageGroup) ShouldDownloadFromVkImage() bool {
	g.SyncLock.Lock()
	defer g.SyncLock.Unlock()
	return g.LeadingImage.ShouldDownloadFromVkImage()
}

func (g *VulkanImageGroup) DownloadFromVkImage(handles *VulkanHandles, commandBuffer *VulkanCommandBuffer, getLinearBuffer func(uintptr) (vk.Buffer, uintptr, error), frame uint64) error {
	g.SyncLock.Lock()
	defer g.SyncLock.Unlock()
	return g.LeadingImage.DownloadFromVkImage(handles, commandBuffer, getLinearBuffer, frame)
}
