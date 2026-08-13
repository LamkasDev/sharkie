package vulkan

import (
	"fmt"

	vk "github.com/goki/vulkan"
)

type VulkanBuffer struct {
	Buffer vk.Buffer
	Memory vk.DeviceMemory
	Size   vk.DeviceSize
	Mapped []byte
}

type VulkanBufferRing struct {
	Buffers []*VulkanBuffer
}

func CreateBufferRing(handles *VulkanHandles, count int, size vk.DeviceSize, usage vk.BufferUsageFlags, properties vk.MemoryPropertyFlags) (*VulkanBufferRing, error) {
	ring := &VulkanBufferRing{
		Buffers: make([]*VulkanBuffer, count),
	}

	for i := 0; i < count; i++ {
		buffer, memory, err := AllocateBuffer(handles, size, usage, properties)
		if err != nil {
			ring.Destroy(handles.Device)
			return nil, fmt.Errorf("failed to allocate buffer ring element %d: %w", i, err)
		}

		mapped := handles.MapMemory(memory, size)
		ring.Buffers[i] = &VulkanBuffer{
			Buffer: buffer,
			Memory: memory,
			Size:   size,
			Mapped: mapped,
		}
	}

	return ring, nil
}

func (r *VulkanBufferRing) Get(frame uint64) *VulkanBuffer {
	return r.Buffers[frame%uint64(len(r.Buffers))]
}

func (r *VulkanBufferRing) Destroy(device vk.Device) {
	for _, b := range r.Buffers {
		if b != nil {
			if b.Mapped != nil {
				vk.UnmapMemory(device, b.Memory)
			}
			if b.Buffer != vk.NullBuffer {
				vk.DestroyBuffer(device, b.Buffer, nil)
			}
			if b.Memory != vk.NullDeviceMemory {
				vk.FreeMemory(device, b.Memory, nil)
			}
		}
	}
}
