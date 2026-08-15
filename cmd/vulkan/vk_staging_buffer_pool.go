package vulkan

import (
	"sync"

	"github.com/LamkasDev/sharkie/cmd/logger"
	vk "github.com/goki/vulkan"
)

type VulkanStagingBuffer struct {
	Buffer vk.Buffer
	Memory vk.DeviceMemory
	Size   vk.DeviceSize
}

type VulkanStagingBufferPool struct {
	Mutex sync.Mutex
	Idle  []*VulkanStagingBuffer
}

// Get finds an idle buffer large enough, or creates a new one if the pool is empty/too small.
func (p *VulkanStagingBufferPool) Get(handles *VulkanHandles, minSize vk.DeviceSize) (*VulkanStagingBuffer, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	// Try to find a suitably sized idle buffer.
	for i, buf := range p.Idle {
		if buf.Size >= minSize {
			p.Idle[i] = p.Idle[len(p.Idle)-1]
			p.Idle = p.Idle[:len(p.Idle)-1]
			return buf, nil
		}
	}

	// No idle buffer large enough found, allocate a new one.
	defaultSize := vk.DeviceSize(2048 * 2048 * 16)
	largeSize := vk.DeviceSize(4096 * 4096 * 16)
	allocSize := minSize
	if minSize <= defaultSize {
		allocSize = defaultSize
	} else if minSize <= largeSize {
		allocSize = largeSize
	} else {
		panic("trying to allocate suspiciously large buffer")
	}

	logger.Printf("allocating staging buffer of %d bytes.\n", allocSize)
	buffer, memory, err := AllocateBuffer(handles, allocSize,
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit|vk.BufferUsageStorageBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, err
	}

	return &VulkanStagingBuffer{Buffer: buffer, Memory: memory, Size: allocSize}, nil
}

func (p *VulkanStagingBufferPool) Put(buffer *VulkanStagingBuffer) {
	p.Mutex.Lock()
	p.Idle = append(p.Idle, buffer)
	p.Mutex.Unlock()
}

func (p *VulkanStagingBufferPool) Destroy(device vk.Device) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	for _, buffer := range p.Idle {
		vk.DestroyBuffer(device, buffer.Buffer, nil)
		vk.FreeMemory(device, buffer.Memory, nil)
	}
	p.Idle = nil
}
