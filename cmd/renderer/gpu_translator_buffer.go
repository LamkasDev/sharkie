package renderer

import "C"
import (
	"fmt"
	"runtime"
	"unsafe"

	as "github.com/LamkasDev/asche"
	vk "github.com/goki/vulkan"
)

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

func (t *GpuTranslator) allocExternalBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, props vk.MemoryPropertyFlags) (vk.Buffer, vk.DeviceMemory, error) {
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
	allocateFlags := vk.MemoryAllocateFlagsInfo{
		SType: vk.StructureTypeMemoryAllocateFlagsInfo,
	}
	if usage&vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit) != 0 {
		allocateFlags.Flags = vk.MemoryAllocateFlags(vk.MemoryAllocateDeviceAddressBit)
	}
	priorityInfo := as.VkMemoryPriorityAllocateInfoEXT{
		SType:    as.StructureTypeMemoryPriorityAllocateInfoExt,
		PNext:    unsafe.Pointer(&allocateFlags),
		Priority: 1.0,
	}
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType: vk.StructureTypeMemoryAllocateInfo,
		PNext: unsafe.Pointer(&vk.ExportMemoryAllocateInfo{
			SType:       vk.StructureTypeExportMemoryAllocateInfo,
			PNext:       unsafe.Pointer(&priorityInfo),
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
	SetDeviceMemoryPriority(t.handles.Instance, t.handles.Device, mem, 1.0)

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
	allocateFlags := vk.MemoryAllocateFlagsInfo{
		SType: vk.StructureTypeMemoryAllocateFlagsInfo,
	}
	if usage&vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit) != 0 {
		allocateFlags.Flags = vk.MemoryAllocateFlags(vk.MemoryAllocateDeviceAddressBit)
	}
	priorityInfo := as.VkMemoryPriorityAllocateInfoEXT{
		SType:    as.StructureTypeMemoryPriorityAllocateInfoExt,
		PNext:    unsafe.Pointer(&allocateFlags),
		Priority: 1.0,
	}
	result = vk.AllocateMemory(t.handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		PNext:           unsafe.Pointer(&priorityInfo),
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: t.handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyFlagBits(props)),
	}, nil, &mem)
	if err := as.NewError(result); err != nil {
		vk.DestroyBuffer(t.handles.Device, buffer, nil)
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkAllocateMemory: %w", err)
	}

	vk.BindBufferMemory(t.handles.Device, buffer, mem, 0)
	SetDeviceMemoryPriority(t.handles.Instance, t.handles.Device, mem, 1.0)

	return buffer, mem, nil
}

func (t *GpuTranslator) createDiscoveryBuffers() error {
	// 64k entries in GlobalDescriptorMap (each uint32)
	var err error
	t.discoveryMapBuffer, t.discoveryMapMem, err = t.allocBuffer(65536*4,
		vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	// MissingResourceBuffer: count (uint32) + 1024 * 12 * uint32
	t.discoveryReportBuffer, t.discoveryReportMem, err = t.allocBuffer(4+1024*48,
		vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	return nil
}
