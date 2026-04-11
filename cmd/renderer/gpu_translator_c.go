package renderer

/*
#include <stdlib.h>
#include <stdint.h>

typedef void* VkInstance;
typedef void* VkDevice;
typedef uint64_t VkDeviceAddress;

#define VK_EXTERNAL_MEMORY_HANDLE_TYPE_OPAQUE_FD_BIT 0x00000001
#define VK_EXTERNAL_MEMORY_HANDLE_TYPE_DMA_BUF_BIT_EXT 0x00000200

typedef struct {
    uint32_t SType;
    const void* PNext;
    void* Buffer;
} VkBufferDeviceAddressInfo;

typedef struct {
    uint32_t SType;
    const void* PNext;
    void* Memory;
    uint32_t HandleType;
} VkMemoryGetFdInfoKHR;

typedef struct {
    uint32_t SType;
    const void* PNext;
    void* Memory;
    uint32_t HandleType;
} VkMemoryGetWin32HandleInfoKHR;

typedef void* (*vgo_vkGetInstanceProcAddr)(VkInstance instance, const char* pName);

typedef VkDeviceAddress (*vgo_vkGetBufferDeviceAddress)(VkDevice device, const VkBufferDeviceAddressInfo* pInfo);
typedef int (*vgo_vkGetMemoryFdKHR)(VkDevice device, const VkMemoryGetFdInfoKHR* pGetFdInfo, int* pFd);
typedef int (*vgo_vkGetMemoryWin32HandleKHR)(VkDevice device, const VkMemoryGetWin32HandleInfoKHR* pGetWin32HandleInfo, void** pHandle);

VkDeviceAddress callVkGetBufferDeviceAddress(void* address, VkInstance instance, VkDevice device, const VkBufferDeviceAddressInfo* info) {
    vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
    vgo_vkGetBufferDeviceAddress fn = (vgo_vkGetBufferDeviceAddress)getProc(instance, "vkGetBufferDeviceAddress");
    if (!fn) {
        fn = (vgo_vkGetBufferDeviceAddress)getProc(instance, "vkGetBufferDeviceAddressKHR");
    }
    if (!fn) { return 0; }

    return fn(device, info);
}

int callVkGetMemoryFdKHR(void* address, VkInstance instance, VkDevice device, const VkMemoryGetFdInfoKHR* info, int* fd) {
    vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
    vgo_vkGetMemoryFdKHR fn = (vgo_vkGetMemoryFdKHR)getProc(instance, "vkGetMemoryFdKHR");
    if (!fn) { return -1; }

    return fn(device, info, fd);
}

int callVkGetMemoryWin32HandleKHR(void* address, VkInstance instance, VkDevice device, const VkMemoryGetWin32HandleInfoKHR* info, void** handle) {
    vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
    vgo_vkGetMemoryWin32HandleKHR fn = (vgo_vkGetMemoryWin32HandleKHR)getProc(instance, "vkGetMemoryWin32HandleKHR");
    if (!fn) { return -1; }

    return fn(device, info, handle);
}
*/
import "C"
import (
	"unsafe"

	"github.com/elokore/glfw/v3.4/glfw"
	vk "github.com/goki/vulkan"
)

func GetBufferDeviceAddress(instance vk.Instance, device vk.Device, buffer vk.Buffer) uint64 {
	info := C.VkBufferDeviceAddressInfo{
		SType:  (C.uint32_t)(vk.StructureTypeBufferDeviceAddressInfo),
		Buffer: unsafe.Pointer(buffer),
	}
	addr := C.callVkGetBufferDeviceAddress(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		(C.VkDevice)(unsafe.Pointer(device)),
		&info,
	)

	return uint64(addr)
}

func GetMemoryFd(instance vk.Instance, device vk.Device, memory vk.DeviceMemory) int {
	info := C.VkMemoryGetFdInfoKHR{
		SType:      (C.uint32_t)(vk.StructureTypeMemoryGetFdInfo),
		Memory:     unsafe.Pointer(memory),
		HandleType: (C.uint32_t)(C.VK_EXTERNAL_MEMORY_HANDLE_TYPE_DMA_BUF_BIT_EXT),
	}
	var fd C.int
	res := C.callVkGetMemoryFdKHR(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		(C.VkDevice)(unsafe.Pointer(device)),
		&info,
		&fd,
	)
	if res != 0 {
		return -1
	}

	return int(fd)
}

func GetMemoryWin32Handle(instance vk.Instance, device vk.Device, memory vk.DeviceMemory) uintptr {
	info := C.VkMemoryGetWin32HandleInfoKHR{
		SType:      (C.uint32_t)(vk.StructureTypeMemoryGetWin32HandleInfo),
		Memory:     unsafe.Pointer(memory),
		HandleType: (C.uint32_t)(vk.ExternalMemoryHandleTypeOpaqueWin32Bit),
	}
	var handle unsafe.Pointer
	res := C.callVkGetMemoryWin32HandleKHR(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		(C.VkDevice)(unsafe.Pointer(device)),
		&info,
		&handle,
	)
	if res != 0 {
		return 0
	}

	return uintptr(handle)
}
