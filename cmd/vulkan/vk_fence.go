package vulkan

import (
	"fmt"
	"sync"

	vk "github.com/goki/vulkan"
)

type VulkanFencePool2 struct {
	Pools []*VulkanFencePool
}

func CreateFencePool2(handles *VulkanHandles, initialCapacity int) (*VulkanFencePool2, error) {
	var err error
	pool := &VulkanFencePool2{
		Pools: make([]*VulkanFencePool, 4),
	}
	for i := range pool.Pools {
		pool.Pools[i], err = CreateFencePool(handles, initialCapacity)
		if err != nil {
			return nil, err
		}
	}

	return pool, nil
}

func (p *VulkanFencePool2) Get(handles *VulkanHandles, frame uint64) (vk.Fence, error) {
	return p.Pools[frame%uint64(len(p.Pools))].Get(handles)
}

func (p *VulkanFencePool2) Put(handles *VulkanHandles, fence vk.Fence, frame uint64) {
	p.Pools[frame%uint64(len(p.Pools))].Put(handles, fence)
}

func (p *VulkanFencePool2) Destroy(handles *VulkanHandles) {
	for _, pool := range p.Pools {
		pool.Destroy(handles)
	}
}

type VulkanFencePool struct {
	Mutex     sync.Mutex
	Available []vk.Fence
	AllFences []vk.Fence
}

func CreateFencePool(handles *VulkanHandles, initialCapacity int) (*VulkanFencePool, error) {
	pool := &VulkanFencePool{
		Available: make([]vk.Fence, 0, initialCapacity),
		AllFences: make([]vk.Fence, 0, initialCapacity),
	}

	for i := 0; i < initialCapacity; i++ {
		fence, err := pool.allocateFence(handles)
		if err != nil {
			pool.Destroy(handles)
			return nil, err
		}
		pool.Available = append(pool.Available, fence)
		pool.AllFences = append(pool.AllFences, fence)
	}

	return pool, nil
}

func (p *VulkanFencePool) allocateFence(handles *VulkanHandles) (vk.Fence, error) {
	var fence vk.Fence
	result := vk.CreateFence(handles.Device, &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: 0,
	}, nil, &fence)
	if err := NewError(result); err != nil {
		return vk.NullFence, fmt.Errorf("failed to create fence: %w", err)
	}

	return fence, nil
}

// Get pops an available fence from the pool. If empty, it allocates a new one.
func (p *VulkanFencePool) Get(handles *VulkanHandles) (vk.Fence, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	// Pop the last available fence.
	if len(p.Available) > 0 {
		fence := p.Available[len(p.Available)-1]
		p.Available = p.Available[:len(p.Available)-1]
		return fence, nil
	}

	// Expand pool if exhausted.
	fence, err := p.allocateFence(handles)
	if err != nil {
		return vk.NullFence, err
	}
	p.AllFences = append(p.AllFences, fence)

	return fence, nil
}

// Put resets the fence to an unsignaled state and returns it to the available pool.
func (p *VulkanFencePool) Put(handles *VulkanHandles, fence vk.Fence) {
	status := vk.GetFenceStatus(handles.Device, fence)
	if status == vk.Success {
		vk.ResetFences(handles.Device, 1, []vk.Fence{fence})
	}
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.Available = append(p.Available, fence)
}

// Destroy cleans up all Vulkan resources managed by the pool.
func (p *VulkanFencePool) Destroy(handles *VulkanHandles) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	for _, fence := range p.AllFences {
		if fence != vk.NullFence {
			vk.DestroyFence(handles.Device, fence, nil)
		}
	}
	p.AllFences = nil
	p.Available = nil
}
