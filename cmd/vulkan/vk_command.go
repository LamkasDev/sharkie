package vulkan

import (
	vk "github.com/goki/vulkan"
)

func CreateCommandPool(handles *VulkanHandles) (vk.CommandPool, error) {
	var pool vk.CommandPool
	result := vk.CreateCommandPool(handles.Device, &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: handles.GraphicsQueueFamilyIndex,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}, nil, &pool)
	if err := NewError(result); err != nil {
		return vk.NullCommandPool, err
	}

	return pool, nil
}

// RunWithCommandBuffer records GPU work into a one-off command buffer and waits (image creation / readback).
func RunWithCommandBuffer(handles *VulkanHandles, fn func(buffer vk.CommandBuffer)) error {
	commandBuffer := handles.AllocateCommandBuffer()
	vk.BeginCommandBuffer(commandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	fn(commandBuffer)
	vk.EndCommandBuffer(commandBuffer)

	status := vk.GetFenceStatus(handles.Device, handles.WorkerFence)
	if status == vk.Success {
		vk.ResetFences(handles.Device, 1, []vk.Fence{handles.WorkerFence})
	}
	handles.QueueMutex.Lock()
	result := vk.QueueSubmit(handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{commandBuffer},
	}}, handles.WorkerFence)
	handles.QueueMutex.Unlock()
	if err := NewError(result); err != nil {
	}
	vk.WaitForFences(handles.Device, 1, []vk.Fence{handles.WorkerFence}, vk.True, vk.MaxUint64)
	vk.FreeCommandBuffers(handles.Device, handles.UploadPool, 1, []vk.CommandBuffer{commandBuffer})

	return nil
}
