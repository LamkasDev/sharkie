package vulkan

import (
	"sync"

	vk "github.com/goki/vulkan"
)

type VulkanQueue struct {
	Queue       vk.Queue
	FamilyIndex uint32
	Lock        sync.Mutex
}

func (q *VulkanQueue) Submit(submitInfos []vk.SubmitInfo, fence vk.Fence) vk.Result {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	return vk.QueueSubmit(q.Queue, uint32(len(submitInfos)), submitInfos, fence)
}

func (q *VulkanQueue) Present(presentInfo *vk.PresentInfo) vk.Result {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	return vk.QueuePresent(q.Queue, presentInfo)
}

func (q *VulkanQueue) WaitIdle() vk.Result {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	return vk.QueueWaitIdle(q.Queue)
}
