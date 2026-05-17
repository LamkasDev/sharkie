package vulkan

import (
	as "github.com/LamkasDev/asche"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) createCommandPoolAndFence() error {
	var pool vk.CommandPool
	result := vk.CreateCommandPool(t.handles.Device, &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: t.handles.GraphicsQueueFamilyIndex,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
	}, nil, &pool)
	if err := as.NewError(result); err != nil {
		return err
	}
	t.pool = pool

	var fence vk.Fence
	vk.CreateFence(t.handles.Device, &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}, nil, &fence)
	t.fence = fence

	var workerFence vk.Fence
	vk.CreateFence(t.handles.Device, &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}, nil, &workerFence)
	t.workerFence = workerFence

	return nil
}
