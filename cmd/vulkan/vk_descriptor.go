package vulkan

import (
	"fmt"

	vk "github.com/goki/vulkan"
)

type VulkanDescriptorPool2 struct {
	Pools []*VulkanDescriptorPool
}

func CreateDescriptorPool2(handles *VulkanHandles, layout vk.DescriptorSetLayout, poolSizes []vk.DescriptorPoolSize, maxSets uint32) (*VulkanDescriptorPool2, error) {
	var err error
	pool := &VulkanDescriptorPool2{
		Pools: make([]*VulkanDescriptorPool, 6),
	}
	for i := range pool.Pools {
		pool.Pools[i], err = CreateDescriptorPool(handles, layout, poolSizes, maxSets)
		if err != nil {
			return nil, err
		}
	}

	return pool, nil
}

func (p *VulkanDescriptorPool2) Get(handles *VulkanHandles, frame uint64) (vk.DescriptorSet, error) {
	return p.Pools[frame%uint64(len(p.Pools))].Get(handles)
}

func (p *VulkanDescriptorPool2) DefaultSet(frame uint64) vk.DescriptorSet {
	return p.Pools[frame%uint64(len(p.Pools))].DefaultSet
}

func (p *VulkanDescriptorPool2) SetCopyTemplate(template []vk.CopyDescriptorSet) {
	for _, pool := range p.Pools {
		pool.CopyTemplate = template
	}
}

func (p *VulkanDescriptorPool2) Reset(frame uint64) {
	p.Pools[frame%uint64(len(p.Pools))].Reset()
}

func (p *VulkanDescriptorPool2) Destroy(handles *VulkanHandles) {
	for _, pool := range p.Pools {
		pool.Destroy(handles)
	}
}

type VulkanDescriptorPool struct {
	Layout       vk.DescriptorSetLayout
	Pool         vk.DescriptorPool
	DefaultSet   vk.DescriptorSet
	Sets         []vk.DescriptorSet
	CopyTemplate []vk.CopyDescriptorSet
	CurrentIndex int
	MaxSets      uint32
}

func CreateDescriptorPool(handles *VulkanHandles, layout vk.DescriptorSetLayout, poolSizes []vk.DescriptorPoolSize, maxSets uint32) (*VulkanDescriptorPool, error) {
	var vkDescriptorPool vk.DescriptorPool
	result := vk.CreateDescriptorPool(handles.Device, &vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		PPoolSizes:    poolSizes,
		PoolSizeCount: uint32(len(poolSizes)),
		MaxSets:       maxSets,
		Flags:         vk.DescriptorPoolCreateFlags(vk.DescriptorPoolCreateFreeDescriptorSetBit | vk.DescriptorPoolCreateUpdateAfterBindBit),
	}, nil, &vkDescriptorPool)
	if err := NewError(result); err != nil {
		return nil, fmt.Errorf("failed to create descriptor pool: %w", err)
	}

	var err error
	pool := &VulkanDescriptorPool{
		Layout:  layout,
		Pool:    vkDescriptorPool,
		Sets:    []vk.DescriptorSet{},
		MaxSets: maxSets,
	}
	pool.DefaultSet, err = pool.Get(handles)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// Get returns the next available descriptor set. It allocates a new one if necessary.
func (p *VulkanDescriptorPool) Get(handles *VulkanHandles) (vk.DescriptorSet, error) {
	if p.CurrentIndex >= int(p.MaxSets) {
		return vk.NullDescriptorSet, fmt.Errorf("descriptor pool exhausted (max %d sets)", p.MaxSets)
	}

	// Allocate a new set if we've run out of pre-allocated ones.
	if p.CurrentIndex >= len(p.Sets) {
		descSets := make([]vk.DescriptorSet, 1)
		result := vk.AllocateDescriptorSets(handles.Device, &vk.DescriptorSetAllocateInfo{
			SType:              vk.StructureTypeDescriptorSetAllocateInfo,
			DescriptorPool:     p.Pool,
			DescriptorSetCount: 1,
			PSetLayouts:        []vk.DescriptorSetLayout{p.Layout},
		}, &descSets[0])

		if err := NewError(result); err != nil {
			return vk.NullDescriptorSet, fmt.Errorf("failed to allocate descriptor set: %w", err)
		}

		// Update the copies to use the new destination set.
		if p.DefaultSet != vk.NullDescriptorSet && p.CopyTemplate != nil {
			copies := make([]vk.CopyDescriptorSet, len(p.CopyTemplate))
			for i, tpl := range p.CopyTemplate {
				copies[i] = tpl
				copies[i].SrcSet = p.DefaultSet
				copies[i].DstSet = descSets[0]
			}
			vk.UpdateDescriptorSets(handles.Device, 0, nil, uint32(len(copies)), copies)
		}

		p.Sets = append(p.Sets, descSets[0])
	}

	// Fetch the set and advance the cursor.
	set := p.Sets[p.CurrentIndex]
	p.CurrentIndex++
	return set, nil
}

// Reset rewinds the pool cursor. Call this at the start of a new frame.
func (p *VulkanDescriptorPool) Reset() {
	p.CurrentIndex = 1
}

// Destroy cleans up the Vulkan resources.
func (p *VulkanDescriptorPool) Destroy(handles *VulkanHandles) {
	if p.Layout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(handles.Device, p.Layout, nil)
		p.Layout = vk.NullDescriptorSetLayout
	}
	if p.Pool != vk.NullDescriptorPool {
		vk.DestroyDescriptorPool(handles.Device, p.Pool, nil)
		p.Pool = vk.NullDescriptorPool
	}
}
