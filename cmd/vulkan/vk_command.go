package vulkan

import (
	"runtime"

	vk "github.com/goki/vulkan"
)

func CreateCommandPoolAndFence(handles *VulkanHandles) (vk.CommandPool, vk.Fence, error) {
	var pool vk.CommandPool
	result := vk.CreateCommandPool(handles.Device, &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: handles.GraphicsQueueFamilyIndex,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}, nil, &pool)
	if err := NewError(result); err != nil {
		return vk.NullCommandPool, vk.NullFence, err
	}

	var workerFence vk.Fence
	vk.CreateFence(handles.Device, &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}, nil, &workerFence)

	return pool, workerFence, nil
}

// RunWithCommandBuffer records GPU work into a one-off command buffer and waits (image creation / readback).
func RunWithCommandBuffer(handles *VulkanHandles, workerFence vk.Fence, fn func(buffer vk.CommandBuffer)) error {
	commandBuffer := handles.AllocateCommandBuffer()
	vk.BeginCommandBuffer(commandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	fn(commandBuffer)
	vk.EndCommandBuffer(commandBuffer)

	commandBuffers := []vk.CommandBuffer{commandBuffer}
	submitInfos := []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    commandBuffers,
	}}

	pinner := &runtime.Pinner{}
	pinner.Pin(&commandBuffers)
	pinner.Pin(&submitInfos)
	defer pinner.Unpin()

	handles.QueueMutex.Lock()
	status := vk.GetFenceStatus(handles.Device, workerFence)
	if status == vk.Success {
		vk.ResetFences(handles.Device, 1, []vk.Fence{workerFence})
	}
	result := vk.QueueSubmit(handles.GraphicsQueue, 1, submitInfos, workerFence)
	handles.QueueMutex.Unlock()
	err := NewError(result)
	if err == nil {
		vk.WaitForFences(handles.Device, 1, []vk.Fence{workerFence}, vk.True, vk.MaxUint64)
	}
	vk.FreeCommandBuffers(handles.Device, handles.UploadPool, 1, []vk.CommandBuffer{commandBuffer})

	return err
}
