package vulkan

/*
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

typedef void* VkInstance;
typedef void* VkDevice;
typedef uint64_t VkDeviceAddress;
typedef uint32_t VkFormat;

typedef struct {
    uint32_t sType;
    const void* pNext;
    uint32_t viewMask;
    uint32_t colorAttachmentCount;
    const VkFormat* pColorAttachmentFormats;
    VkFormat depthAttachmentFormat;
    VkFormat stencilAttachmentFormat;
} VkPipelineRenderingCreateInfo;
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

typedef struct {
    uint32_t SType;
    const void* PNext;
    uint32_t ObjectType;
    uint64_t ObjectHandle;
    const char* PObjectName;
} VkDebugUtilsObjectNameInfoEXT;

typedef void* (*vgo_vkGetInstanceProcAddr)(VkInstance instance, const char* pName);

typedef void (*vgo_vkSetDeviceMemoryPriorityEXT)(VkDevice device, void* memory, float priority);
typedef VkDeviceAddress (*vgo_vkGetBufferDeviceAddress)(VkDevice device, const VkBufferDeviceAddressInfo* pInfo);
typedef int (*vgo_vkGetMemoryFdKHR)(VkDevice device, const VkMemoryGetFdInfoKHR* pGetFdInfo, int* pFd);
typedef int (*vgo_vkGetMemoryWin32HandleKHR)(VkDevice device, const VkMemoryGetWin32HandleInfoKHR* pGetWin32HandleInfo, void** pHandle);
typedef void (*vgo_vkGetPhysicalDeviceProperties2KHR)(void* physicalDevice, void* pProperties);
typedef int (*vgo_vkSetDebugUtilsObjectNameEXT)(VkDevice device, const VkDebugUtilsObjectNameInfoEXT* pNameInfo);
typedef void (*vgo_vkCmdBeginRendering)(void* commandBuffer, const void* pRenderingInfo);
typedef void (*vgo_vkCmdEndRendering)(void* commandBuffer);

void callVkSetDeviceMemoryPriorityEXT(void* address, VkInstance instance, VkDevice device, void* memory, float priority) {
	vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
	vgo_vkSetDeviceMemoryPriorityEXT fn = (vgo_vkSetDeviceMemoryPriorityEXT)getProc(instance, "vkSetDeviceMemoryPriorityEXT");
	if (!fn) { return; }

	fn(device, memory, priority);
}

void callVkGetPhysicalDeviceProperties2KHR(void* address, VkInstance instance, void* physicalDevice, void* pProperties) {
	vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
	vgo_vkGetPhysicalDeviceProperties2KHR fn = (vgo_vkGetPhysicalDeviceProperties2KHR)getProc(instance, "vkGetPhysicalDeviceProperties2KHR");
	if (!fn) {
		fn = (vgo_vkGetPhysicalDeviceProperties2KHR)getProc(instance, "vkGetPhysicalDeviceProperties2");
	}
	if (!fn) { return; }

	fn(physicalDevice, pProperties);
}

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

int32_t callVkSetDebugUtilsObjectNameEXT(void* address, VkInstance instance, VkDevice device, const VkDebugUtilsObjectNameInfoEXT* pNameInfo) {
    vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
    vgo_vkSetDebugUtilsObjectNameEXT fn = (vgo_vkSetDebugUtilsObjectNameEXT)getProc(instance, "vkSetDebugUtilsObjectNameEXT");
    if (!fn) { return -1; }

    return fn(device, pNameInfo);
}

void callVkCmdBeginRendering(void* address, VkInstance instance, void* commandBuffer, const void* pRenderingInfo) {
    vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
    vgo_vkCmdBeginRendering fn = (vgo_vkCmdBeginRendering)getProc(instance, "vkCmdBeginRenderingKHR");
    if (!fn) fn = (vgo_vkCmdBeginRendering)getProc(instance, "vkCmdBeginRendering");
    if (!fn) return;
    fn(commandBuffer, pRenderingInfo);
}

void callVkCmdEndRendering(void* address, VkInstance instance, void* commandBuffer) {
    vgo_vkGetInstanceProcAddr getProc = (vgo_vkGetInstanceProcAddr)address;
    vgo_vkCmdEndRendering fn = (vgo_vkCmdEndRendering)getProc(instance, "vkCmdEndRenderingKHR");
    if (!fn) fn = (vgo_vkCmdEndRendering)getProc(instance, "vkCmdEndRendering");
    if (!fn) return;
    fn(commandBuffer);
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

func SetDeviceMemoryPriority(instance vk.Instance, device vk.Device, memory vk.DeviceMemory, priority float32) {
	C.callVkSetDeviceMemoryPriorityEXT(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		(C.VkDevice)(unsafe.Pointer(device)),
		unsafe.Pointer(memory),
		(C.float)(priority),
	)
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

func GetPhysicalDeviceProperties2(instance vk.Instance, physicalDevice vk.PhysicalDevice, props unsafe.Pointer) {
	C.callVkGetPhysicalDeviceProperties2KHR(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		unsafe.Pointer(physicalDevice),
		props,
	)
}

func SetDebugUtilsObjectName(instance vk.Instance, device vk.Device, objectType vk.ObjectType, objectHandle uint64, name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	info := C.VkDebugUtilsObjectNameInfoEXT{
		SType:        (C.uint32_t)(vk.StructureTypeDebugUtilsObjectNameInfo),
		PNext:        nil,
		ObjectType:   (C.uint32_t)(objectType),
		ObjectHandle: (C.uint64_t)(objectHandle),
		PObjectName:  cName,
	}
	C.callVkSetDebugUtilsObjectNameEXT(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		(C.VkDevice)(unsafe.Pointer(device)),
		&info,
	)
}

func CmdBeginRendering(instance vk.Instance, commandBuffer vk.CommandBuffer, renderingInfo *vk.RenderingInfo) {
	cRenderingInfo, _ := renderingInfo.PassRef()
	C.callVkCmdBeginRendering(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		unsafe.Pointer(commandBuffer),
		unsafe.Pointer(cRenderingInfo),
	)
}

func CmdEndRendering(instance vk.Instance, commandBuffer vk.CommandBuffer) {
	C.callVkCmdEndRendering(
		unsafe.Pointer(glfw.GetVulkanGetInstanceProcAddress()),
		(C.VkInstance)(unsafe.Pointer(instance)),
		unsafe.Pointer(commandBuffer),
	)
}

func CreatePipelineRenderingCreateInfoC(info *vk.PipelineRenderingCreateInfo) (unsafe.Pointer, func()) {
	cinfo := (*C.VkPipelineRenderingCreateInfo)(C.malloc(C.sizeof_VkPipelineRenderingCreateInfo))
	C.memset(unsafe.Pointer(cinfo), 0, C.sizeof_VkPipelineRenderingCreateInfo)
	cinfo.sType = C.uint32_t(info.SType)
	cinfo.pNext = nil
	cinfo.viewMask = C.uint32_t(info.ViewMask)
	cinfo.colorAttachmentCount = C.uint32_t(info.ColorAttachmentCount)

	var cFormats *C.VkFormat
	if len(info.PColorAttachmentFormats) > 0 {
		cFormats = (*C.VkFormat)(C.malloc(C.size_t(4 * len(info.PColorAttachmentFormats))))
		formatSlice := unsafe.Slice((*vk.Format)(unsafe.Pointer(cFormats)), len(info.PColorAttachmentFormats))
		for i, v := range info.PColorAttachmentFormats {
			formatSlice[i] = v
		}
	}
	cinfo.pColorAttachmentFormats = cFormats
	cinfo.depthAttachmentFormat = C.VkFormat(info.DepthAttachmentFormat)
	cinfo.stencilAttachmentFormat = C.VkFormat(info.StencilAttachmentFormat)

	return unsafe.Pointer(cinfo), func() {
		if cFormats != nil {
			C.free(unsafe.Pointer(cFormats))
		}
		C.free(unsafe.Pointer(cinfo))
	}
}
