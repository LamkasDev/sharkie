package renderer

import (
	"fmt"
	"runtime"
	"unsafe"

	as "github.com/LamkasDev/asche"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) UpdateConstRamBuffers(draws []LiverpoolDrawCall) {
	t.constRamBuffersMutex.Lock()
	defer t.constRamBuffersMutex.Unlock()

	// Find unique hashes in current draw calls.
	activeHashes := make(map[uint32]bool)
	for i := range draws {
		activeHashes[draws[i].ConstRamHash] = true
	}

	// Delete buffers that are no longer active.
	for hash, buffer := range t.constRamBuffers {
		if !activeHashes[hash] {
			vk.DestroyBuffer(t.handles.Device, buffer, nil)
			vk.FreeMemory(t.handles.Device, t.constRamBufferMems[hash], nil)
			delete(t.constRamBuffers, hash)
			delete(t.constRamBuffersDebug, hash)
			delete(t.constRamBufferMems, hash)
		}
	}

	// Create buffers for new active hashes.
	for hash := range activeHashes {
		if _, ok := t.constRamBuffers[hash]; ok {
			continue
		}

		// Get contents from global state.
		contents, ok := GlobalConstRamSnapshots[hash]
		if !ok {
			continue
		}

		// Allocate and upload.
		size := vk.DeviceSize(len(contents) * 4)
		buffer, mem, err := t.allocBuffer(size,
			vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageUniformBufferBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
		if err != nil {
			panic(fmt.Errorf("allocConstRamBuffer: %w", err))
		}

		data := t.handles.MapMemory(mem, size)
		copy(data, unsafe.Slice((*byte)(unsafe.Pointer(&contents[0])), size))
		vk.UnmapMemory(t.handles.Device, mem)

		t.constRamBuffers[hash] = buffer
		t.constRamBuffersDebug[hash] = contents[:]
		t.constRamBufferMems[hash] = mem
	}
}

func (t *GpuTranslator) UpdateUserDataBuffers(draws []LiverpoolDrawCall) {
	t.userDataBuffersMutex.Lock()
	defer t.userDataBuffersMutex.Unlock()

	// Find unique hashes in current draw calls.
	activeHashes := make(map[uint32]bool)
	for i := range draws {
		activeHashes[draws[i].UserDataHash] = true
	}

	// Delete buffers that are no longer active.
	for hash, buffer := range t.userDataBuffers {
		if !activeHashes[hash] {
			vk.DestroyBuffer(t.handles.Device, buffer, nil)
			vk.FreeMemory(t.handles.Device, t.userDataBufferMems[hash], nil)
			delete(t.userDataBuffers, hash)
			delete(t.userDataBuffersDebug, hash)
			delete(t.userDataBufferMems, hash)
		}
	}

	// Create buffers for new active hashes.
	for hash := range activeHashes {
		if _, ok := t.userDataBuffers[hash]; ok {
			continue
		}

		// Get contents from global state.
		contents, ok := GlobalUserDataSnapshots[hash]
		if !ok {
			continue
		}

		// Allocate and upload.
		size := vk.DeviceSize(len(contents) * 4)
		buffer, mem, err := t.allocBuffer(size,
			vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageUniformBufferBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
		if err != nil {
			panic(fmt.Errorf("allocUserDataBuffer: %w", err))
		}

		data := t.handles.MapMemory(mem, size)
		copy(data, unsafe.Slice((*byte)(unsafe.Pointer(&contents[0])), size))
		vk.UnmapMemory(t.handles.Device, mem)

		t.userDataBuffers[hash] = buffer
		t.userDataBuffersDebug[hash] = contents[:]
		t.userDataBufferMems[hash] = mem
	}
}

func (t *GpuTranslator) AllocExternalBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, props vk.MemoryPropertyFlags) (vk.Buffer, vk.DeviceMemory, error) {
	handleType := vk.ExternalMemoryHandleTypeDmaBufBit
	if runtime.GOOS == "windows" {
		handleType = vk.ExternalMemoryHandleTypeOpaqueWin32Bit
	}

	var buffer vk.Buffer
	result := vk.CreateBuffer(t.handles.Device, &vk.BufferCreateInfo{
		SType: vk.StructureTypeBufferCreateInfo,
		PNext: unsafe.Pointer(&vk.ExternalMemoryBufferCreateInfo{
			SType:       vk.StructureTypeExternalMemoryBufferCreateInfo,
			HandleTypes: vk.ExternalMemoryHandleTypeFlags(handleType),
		}),
		Size:  size,
		Usage: usage,
	}, nil, &buffer)
	if err := as.NewError(result); err != nil {
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkCreateBuffer: %w", err)
	}

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(t.handles.Device, buffer, &memReqs)
	memReqs.Deref()

	var mem vk.DeviceMemory
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType: vk.StructureTypeMemoryAllocateInfo,
		PNext: unsafe.Pointer(&vk.ExportMemoryAllocateInfo{
			SType:       vk.StructureTypeExportMemoryAllocateInfo,
			HandleTypes: vk.ExternalMemoryHandleTypeFlags(handleType),
		}),
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyFlagBits(props)),
	}, nil, &mem)
	if err := as.NewError(result); err != nil {
		vk.DestroyBuffer(t.handles.Device, buffer, nil)
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkAllocateMemory: %w", err)
	}

	vk.BindBufferMemory(t.handles.Device, buffer, mem, 0)

	return buffer, mem, nil
}

func (t *GpuTranslator) allocBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, props vk.MemoryPropertyFlags) (vk.Buffer, vk.DeviceMemory, error) {
	var buffer vk.Buffer
	result := vk.CreateBuffer(t.handles.Device, &vk.BufferCreateInfo{
		SType: vk.StructureTypeBufferCreateInfo,
		Size:  size,
		Usage: usage,
	}, nil, &buffer)
	if err := as.NewError(result); err != nil {
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkCreateBuffer: %w", err)
	}

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(t.handles.Device, buffer, &memReqs)
	memReqs.Deref()

	var mem vk.DeviceMemory
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyFlagBits(props)),
	}, nil, &mem)
	if err := as.NewError(result); err != nil {
		vk.DestroyBuffer(t.handles.Device, buffer, nil)
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkAllocateMemory: %w", err)
	}

	vk.BindBufferMemory(t.handles.Device, buffer, mem, 0)

	return buffer, mem, nil
}

func (t *GpuTranslator) createCommandPool() error {
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

	return nil
}
