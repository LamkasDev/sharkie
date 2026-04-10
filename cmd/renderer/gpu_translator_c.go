package renderer

/*
#include <stdlib.h>
#include <stdint.h>

typedef void* VkInstance;
typedef void* VkDevice;
typedef uint64_t VkDeviceAddress;

typedef struct {
    uint32_t SType;
    const void* PNext;
    void* Buffer;
} VkBufferDeviceAddressInfo;

typedef void* (*vgo_vkGetInstanceProcAddr)(VkInstance instance, const char* pName);

typedef VkDeviceAddress (*vgo_vkGetBufferDeviceAddress)(VkDevice device, const VkBufferDeviceAddressInfo* pInfo);

VkDeviceAddress callVkGetBufferDeviceAddress(void* address, VkInstance instance, VkDevice device, const VkBufferDeviceAddressInfo* info) {
    vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
    vgo_vkGetBufferDeviceAddress fn = (vgo_vkGetBufferDeviceAddress)getProc(instance, "vkGetBufferDeviceAddress");
    if (!fn) {
        fn = (vgo_vkGetBufferDeviceAddress)getProc(instance, "vkGetBufferDeviceAddressKHR");
    }
    if (!fn) { return 0; }

    return fn(device, info);
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
