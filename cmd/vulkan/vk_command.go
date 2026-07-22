package vulkan

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

type VulkanCommandBuffer struct {
	CommandBuffer vk.CommandBuffer
	Dependencies  []*gpu.LiverpoolWaitRegMemory
	Writes        []*gpu.LiverpoolWriteData
	Submitted     bool
}

func CreateCommandBuffer(handles *VulkanHandles) (*VulkanCommandBuffer, error) {
	buffers := make([]vk.CommandBuffer, 1)
	handles.UploadPoolMutex.Lock()
	result := vk.AllocateCommandBuffers(handles.Device, &vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        handles.UploadPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}, buffers)
	if err := NewError(result); err != nil {
		return nil, err
	}
	handles.UploadPoolMutex.Unlock()

	return &VulkanCommandBuffer{
		CommandBuffer: buffers[0],
		Dependencies:  []*gpu.LiverpoolWaitRegMemory{},
		Writes:        []*gpu.LiverpoolWriteData{},
	}, nil
}

func (commandBuffer *VulkanCommandBuffer) Destroy(handles *VulkanHandles) {
	handles.UploadPoolMutex.Lock()
	vk.FreeCommandBuffers(handles.Device, handles.UploadPool, 1, []vk.CommandBuffer{commandBuffer.CommandBuffer})
	handles.UploadPoolMutex.Unlock()
}

func (commandBuffer *VulkanCommandBuffer) CanSubmit(frame uint64) bool {
	for _, dependency := range commandBuffer.Dependencies {
		if logger.LogRendererInternal {
			logger.Printf("[%s] waiting on reg memory (address=%s, function=%s, reference=%s).\n",
				color.Blue.Sprintf("Frame %d", frame),
				color.Yellow.Sprintf("0x%X", dependency.Address),
				color.Yellow.Sprintf("0x%X", dependency.Function),
				color.Yellow.Sprintf("0x%X", dependency.Reference),
			)
		}
		if !dependency.Satisfied() {
			return false
		}
		if logger.LogRendererInternal {
			logger.Printf("[%s] finished waiting on reg memory.\n",
				color.Blue.Sprintf("Frame %d", frame),
			)
		}
	}

	return true
}

// RunWithCommandBuffer records GPU work into a one-off command buffer and waits (image creation / readback).
func RunWithCommandBuffer(handles *VulkanHandles, fence vk.Fence, fn func(buffer *VulkanCommandBuffer)) error {
	commandBuffer, err := CreateCommandBuffer(handles)
	if err != nil {
		return err
	}
	vk.BeginCommandBuffer(commandBuffer.CommandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	fn(commandBuffer)
	vk.EndCommandBuffer(commandBuffer.CommandBuffer)

	status := vk.GetFenceStatus(handles.Device, fence)
	if status == vk.Success {
		vk.ResetFences(handles.Device, 1, []vk.Fence{fence})
	}
	handles.QueueMutex.Lock()
	result := vk.QueueSubmit(handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{commandBuffer.CommandBuffer},
	}}, fence)
	handles.QueueMutex.Unlock()
	if err = NewError(result); err != nil {
		return err
	}
	vk.WaitForFences(handles.Device, 1, []vk.Fence{fence}, vk.True, vk.MaxUint64)
	commandBuffer.Destroy(handles)

	return nil
}
