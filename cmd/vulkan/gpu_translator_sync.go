package vulkan

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) memorySyncWorker() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	logger.Printf("Memory Sync Worker started.\n")
	for {
		logger.Printf("[SyncWorker] Waiting for sync request...\n")
		addr := structs.WaitForSyncRequest()
		if addr == 0 {
			continue
		}
		logger.Printf("[SyncWorker] Got sync request for 0x%X\n", addr)

		err := t.syncTrackedPage(addr)
		if err != nil {
			logger.Printf("[SyncWorker] Failed to sync memory at 0x%X: %v\n", addr, err)
		} else {
			logger.Printf("[SyncWorker] Successfully synced memory at 0x%X\n", addr)
		}

		structs.CompleteSyncRequest()
	}
}

func (t *GpuTranslator) syncTrackedPage(faultAddr uintptr) error {
	structs.GlobalMemoryManager.Lock.RLock()
	trackedPage, ok := structs.GlobalMemoryManager.TrackedPages[faultAddr]
	if !ok {
		structs.GlobalMemoryManager.Lock.RUnlock()
		return fmt.Errorf("no tracked page at 0x%X", faultAddr)
	}
	addr := trackedPage.BaseAddress
	alignedAddr := addr & ^(structs.SystemPageSize - 1)
	format, ok := structs.GlobalMemoryManager.PageFormats[alignedAddr]
	structs.GlobalMemoryManager.Lock.RUnlock()

	if !ok || format == nil {
		return fmt.Errorf("no format registered for 0x%X", addr)
	}

	var image vk.Image
	var expectedLayout vk.ImageLayout
	t.surfacesMutex.Lock()
	for key, surface := range t.surfaces {
		if key.GpuAddress == addr {
			image = surface.Value.image
			expectedLayout = vk.ImageLayoutGeneral
			break
		}
	}
	t.surfacesMutex.Unlock()

	if image == vk.NullImage {
		t.imagesMutex.Lock()
		img, ok := t.images[addr]
		if ok {
			image = img
			expectedLayout = vk.ImageLayoutGeneral
		}
		t.imagesMutex.Unlock()
	}

	if image == vk.NullImage {
		return fmt.Errorf("no VkImage found for 0x%X", addr)
	}

	// Check actual dimensions from image descriptors if available
	actualHeight := format.Height
	actualWidth := format.Pitch
	t.imagesMutex.Lock()
	if desc, ok := t.imageDescriptors[uint64(addr)]; ok {
		if uint32(desc.Height) < actualHeight {
			actualHeight = uint32(desc.Height)
		}
		if uint32(desc.Width) < actualWidth {
			actualWidth = uint32(desc.Width)
		}
	}
	t.imagesMutex.Unlock()

	bpp := structs.GetBytesPerPixel(format.DataFormat)
	size := vk.DeviceSize(actualWidth * actualHeight * bpp)

	// Create temporary buffer
	buffer, bufferMem, err := t.AllocBuffer(size, vk.BufferUsageFlags(vk.BufferUsageTransferDstBit), vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return fmt.Errorf("failed to allocate staging buffer: %w", err)
	}
	defer vk.DestroyBuffer(t.handles.Device, buffer, nil)
	defer vk.FreeMemory(t.handles.Device, bufferMem, nil)

	// Allocate command buffer
	cmdBuf := t.handles.AllocateCommandBuffer(t.handles.UploadPool)
	defer vk.FreeCommandBuffers(t.handles.Device, t.handles.UploadPool, 1, []vk.CommandBuffer{cmdBuf})

	vk.BeginCommandBuffer(cmdBuf, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})

	// Transition to TransferSrcOptimal
	vk.CmdPipelineBarrier(cmdBuf, vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{{
		SType:               vk.StructureTypeImageMemoryBarrier,
		SrcAccessMask:       vk.AccessFlags(vk.AccessMemoryWriteBit),
		DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
		OldLayout:           expectedLayout,
		NewLayout:           vk.ImageLayoutTransferSrcOptimal,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               image,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}})

	vk.CmdCopyImageToBuffer(cmdBuf, image, vk.ImageLayoutTransferSrcOptimal, buffer, 1, []vk.BufferImageCopy{{
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
		ImageExtent: vk.Extent3D{Width: actualWidth, Height: actualHeight, Depth: 1},
	}})

	// Transition back
	vk.CmdPipelineBarrier(cmdBuf, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit), 0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{{
		SType:               vk.StructureTypeImageMemoryBarrier,
		SrcAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
		DstAccessMask:       vk.AccessFlags(vk.AccessMemoryReadBit),
		OldLayout:           vk.ImageLayoutTransferSrcOptimal,
		NewLayout:           expectedLayout,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               image,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}})

	vk.EndCommandBuffer(cmdBuf)

	// Submit synchronously
	t.QueueMutex.Lock()
	vk.QueueSubmit(t.handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cmdBuf},
	}}, vk.NullFence)
	res := vk.QueueWaitIdle(t.handles.GraphicsQueue)
	if res != vk.Success {
		t.QueueMutex.Unlock()
		return fmt.Errorf("QueueWaitIdle failed: %v", res)
	}
	t.QueueMutex.Unlock()

	// Map memory
	memPtr := t.handles.MapMemory(bufferMem, size)
	defer vk.UnmapMemory(t.handles.Device, bufferMem)

	// Swizzle (Linear Vulkan -> Tiled PS4 CPU)
	info := structs.TextureInfo{
		Width:         actualWidth,
		Height:        actualHeight,
		Pitch:         actualWidth,
		Format:        format.DataFormat,
		TilingIndex:   format.TilingIndex,
		BytesPerPixel: bpp,
	}
	swizzled := structs.SwizzleTexture(memPtr, info)

	// Copy swizzled data to actual CPU memory
	// First, unprotect the memory so the worker doesn't page fault
	pageMask := uintptr(4096 - 1)
	alignedAddr = addr &^ pageMask
	alignedSize := (uintptr(size) + (addr - alignedAddr) + pageMask) &^ pageMask
	mprotectSlice := unsafe.Slice((*byte)(unsafe.Pointer(alignedAddr)), alignedSize)
	err = syscall.Mprotect(mprotectSlice, syscall.PROT_READ|syscall.PROT_WRITE)
	if err != nil {
		logger.Printf("[SyncWorker] Failed to unprotect memory at 0x%X: %v\n", alignedAddr, err)
	}

	cpuSlice := unsafe.Slice((*byte)(unsafe.Pointer(addr)), size)
	copy(cpuSlice, swizzled)

	// Reprotect the entire region as PROT_READ so we can catch future CPU writes!
	err = syscall.Mprotect(mprotectSlice, syscall.PROT_READ)
	if err != nil {
		logger.Printf("[SyncWorker] Failed to reprotect memory at 0x%X to PROT_READ: %v\n", alignedAddr, err)
	}

	// Update C tracking state so segv_handler knows the region is now PROT_READ
	structs.SetRegionProtState(alignedAddr, alignedSize, 1)

	// We DO NOT untrack the region! We leave it tracked so we can catch future CPU writes.
	logger.Printf("[SyncWorker] Untrack finished\n")

	return nil
}
