package vulkan

import vk "github.com/goki/vulkan"

func CreateFence(handles *VulkanHandles) (vk.Fence, error) {
	var fence vk.Fence
	vk.CreateFence(handles.Device, &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}, nil, &fence)

	return fence, nil
}
