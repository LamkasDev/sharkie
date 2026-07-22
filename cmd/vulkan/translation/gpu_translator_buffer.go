package translation

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetBufferView(descriptor spirvStructs.BufferDescriptor) (vk.BufferView, error) {
	// If the stride is 0 or elements are 0, return null view.
	if descriptor.Stride == 0 || descriptor.NumRecords == 0 {
		return vk.NullBufferView, nil
	}

	format, _ := vulkan.TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
	if format == vk.FormatUndefined {
		format = vk.FormatR32Uint // Fallback if unknown
	}

	// We decode EVERYTHING into R32G32B32A32_UINT
	texelSize := 16
	format = vk.FormatR32g32b32a32Uint

	// Allocate a new tight buffer on the device.
	tightSize := vk.DeviceSize(descriptor.NumRecords * uint32(texelSize))
	buffer, mem, err := vulkan.AllocateBuffer(&t.handles, tightSize,
		vk.BufferUsageFlags(vk.BufferUsageUniformTexelBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return vk.NullBufferView, err
	}

	// Read from the guest buffer linearly.
	var pData unsafe.Pointer
	if result := vk.MapMemory(t.handles.Device, mem, 0, tightSize, 0, &pData); result == vk.Success {
		dst := unsafe.Slice((*byte)(pData), tightSize)
		src := unsafe.Slice((*byte)(unsafe.Pointer(descriptor.BaseAddress)), uint32(descriptor.Stride)*descriptor.NumRecords)

		for i := uint32(0); i < descriptor.NumRecords; i++ {
			srcOffset := i * uint32(descriptor.Stride)
			dstOffset := i * uint32(texelSize)
			// Ensure we don't read out of bounds.
			if int(srcOffset+uint32(descriptor.Stride)) <= len(src) && int(dstOffset+uint32(texelSize)) <= len(dst) {
				decoded := DecodeGcnBufferFormat(src[srcOffset:srcOffset+uint32(descriptor.Stride)], descriptor.DataFormat, descriptor.NumFormat)
				// Write 4 uint32s to dst
				for j := 0; j < 4; j++ {
					dst[dstOffset+uint32(j*4)] = byte(decoded[j])
					dst[dstOffset+uint32(j*4)+1] = byte(decoded[j] >> 8)
					dst[dstOffset+uint32(j*4)+2] = byte(decoded[j] >> 16)
					dst[dstOffset+uint32(j*4)+3] = byte(decoded[j] >> 24)
				}
			}
		}
		logger.Printf("%+v\n", src)

		vk.UnmapMemory(t.handles.Device, mem)
	}

	// Create BufferView
	var bufferView vk.BufferView
	result := vk.CreateBufferView(t.handles.Device, &vk.BufferViewCreateInfo{
		SType:  vk.StructureTypeBufferViewCreateInfo,
		Buffer: buffer,
		Format: format,
		Offset: 0,
		Range:  tightSize,
	}, nil, &bufferView)
	if err := vulkan.NewError(result); err != nil {
		vk.DestroyBuffer(t.handles.Device, buffer, nil)
		vk.FreeMemory(t.handles.Device, mem, nil)
		return vk.NullBufferView, err
	}

	t.deferDestroyBuffer(buffer, mem)
	t.deferDestroyBufferView(bufferView)

	return bufferView, nil
}
