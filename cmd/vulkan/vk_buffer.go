package vulkan

import "C"
import (
	"fmt"
	"runtime"
	"unsafe"

	vk "github.com/goki/vulkan"
)

func AllocateExternalBuffer(handles *VulkanHandles, size vk.DeviceSize, usage vk.BufferUsageFlags, props vk.MemoryPropertyFlags) (vk.Buffer, vk.DeviceMemory, error) {
	handleType := vk.ExternalMemoryHandleTypeDmaBufBit
	if runtime.GOOS == "windows" {
		handleType = vk.ExternalMemoryHandleTypeOpaqueWin32Bit
	}

	var buffer vk.Buffer
	result := vk.CreateBuffer(handles.Device, &vk.BufferCreateInfo{
		SType: vk.StructureTypeBufferCreateInfo,
		PNext: unsafe.Pointer(&vk.ExternalMemoryBufferCreateInfo{
			SType:       vk.StructureTypeExternalMemoryBufferCreateInfo,
			HandleTypes: vk.ExternalMemoryHandleTypeFlags(handleType),
		}),
		Size:  size,
		Usage: usage,
	}, nil, &buffer)
	if err := NewError(result); err != nil {
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkCreateBuffer: %w", err)
	}

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(handles.Device, buffer, &memReqs)
	memReqs.Deref()

	var mem vk.DeviceMemory
	priorityInfo := NewPriorityInfo(usage, 1.0)
	result = vk.AllocateMemory(handles.Device, &vk.MemoryAllocateInfo{
		SType: vk.StructureTypeMemoryAllocateInfo,
		PNext: unsafe.Pointer(&vk.ExportMemoryAllocateInfo{
			SType:       vk.StructureTypeExportMemoryAllocateInfo,
			PNext:       unsafe.Pointer(&priorityInfo),
			HandleTypes: vk.ExternalMemoryHandleTypeFlags(handleType),
		}),
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyFlagBits(props)),
	}, nil, &mem)
	if err := NewError(result); err != nil {
		vk.DestroyBuffer(handles.Device, buffer, nil)
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkAllocateMemory: %w", err)
	}

	SetDeviceMemoryPriority(handles.Instance, handles.Device, mem, 1.0)
	vk.BindBufferMemory(handles.Device, buffer, mem, 0)

	return buffer, mem, nil
}

func AllocateBuffer(handles *VulkanHandles, size vk.DeviceSize, usage vk.BufferUsageFlags, props vk.MemoryPropertyFlags) (vk.Buffer, vk.DeviceMemory, error) {
	var buffer vk.Buffer
	result := vk.CreateBuffer(handles.Device, &vk.BufferCreateInfo{
		SType: vk.StructureTypeBufferCreateInfo,
		Size:  size,
		Usage: usage,
	}, nil, &buffer)
	if err := NewError(result); err != nil {
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkCreateBuffer: %w", err)
	}

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(handles.Device, buffer, &memReqs)
	memReqs.Deref()

	var mem vk.DeviceMemory
	priorityInfo := NewPriorityInfo(usage, 1.0)
	result = vk.AllocateMemory(handles.Device, &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		PNext:           unsafe.Pointer(&priorityInfo),
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: handles.FindMemoryType(memReqs.MemoryTypeBits, vk.MemoryPropertyFlagBits(props)),
	}, nil, &mem)
	if err := NewError(result); err != nil {
		vk.DestroyBuffer(handles.Device, buffer, nil)
		return vk.NullBuffer, vk.NullDeviceMemory, fmt.Errorf("vkAllocateMemory: %w", err)
	}

	SetDeviceMemoryPriority(handles.Instance, handles.Device, mem, 1.0)
	vk.BindBufferMemory(handles.Device, buffer, mem, 0)

	return buffer, mem, nil
}

func NewPriorityInfo(usage vk.BufferUsageFlags, priority float32) VkMemoryPriorityAllocateInfoEXT {
	allocateFlags := vk.MemoryAllocateFlagsInfo{
		SType: vk.StructureTypeMemoryAllocateFlagsInfo,
	}
	if usage&vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit) != 0 {
		allocateFlags.Flags = vk.MemoryAllocateFlags(vk.MemoryAllocateDeviceAddressBit)
	}

	return VkMemoryPriorityAllocateInfoEXT{
		SType:    StructureTypeMemoryPriorityAllocateInfoExt,
		PNext:    unsafe.Pointer(&allocateFlags),
		Priority: priority,
	}
}
