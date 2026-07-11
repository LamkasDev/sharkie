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
