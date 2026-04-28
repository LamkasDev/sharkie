package renderer

import (
	"fmt"
	"runtime"
	"unsafe"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/structs/spirv"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) UpdateUserDataBuffers(draws []LiverpoolDrawCall) {
	t.userDataBuffersMutex.Lock()
	defer t.userDataBuffersMutex.Unlock()

	// Find unique hashes in current draw calls.
	activeHashes := make(map[uint32]bool)
	for i := range draws {
		activeHashes[draws[i].UserDataHash] = true
	}

	// Delete buffers that are no longer active.
	// TODO: we need to delete it only once it's out of use.
	/* for hash, buffer := range t.userDataBuffers {
	if !activeHashes[hash] {
		vk.DestroyBuffer(t.handles.Device, buffer, nil)
		vk.FreeMemory(t.handles.Device, t.userDataBufferMems[hash], nil)
		delete(t.userDataBuffers, hash)
		delete(t.userDataBuffersDebug, hash)
		delete(t.userDataBufferMems, hash)
		/* logger.Printf("[%s] Deleted user data with hash %s.\n",
			color.Blue.Sprint("GPU"),
			color.Yellow.Sprintf("0x%X", hash),
		) */ /*
		}
	} */

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
			vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageUniformBufferBit|vk.BufferUsageUniformTexelBufferBit),
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
		logger.Printf("[%s] Created user data with hash %s (%x).\n",
			color.Blue.Sprint("GPU"),
			color.Yellow.Sprintf("0x%X", hash),
			contents[:16],
		)
	}
}

func (t *GpuTranslator) BindTexelBuffers(commandBuffer vk.CommandBuffer, draw *LiverpoolDrawCall, userData []uint32) ([4]uint32, [4]uint32) {
	var bufferViews [4]vk.BufferView
	var viewCount uint32
	var formatSizes, formatStrides [4]uint32
	for i := range 4 {
		sgprBase := i * 4
		descriptor := NewBufferDescriptor(
			userData[sgprBase],
			userData[sgprBase+1],
			userData[sgprBase+2],
			userData[sgprBase+3],
		)
		if descriptor.BaseAddress == 0 && descriptor.Records == 0 {
			continue
		}

		// Route to the correct buffer based on address range.
		var targetBuffer vk.Buffer
		var relativeOffset uintptr
		if descriptor.BaseAddress >= GlobalGpuAllocator.Base && descriptor.BaseAddress < GlobalGpuAllocator.Base+uintptr(GlobalGpuAllocator.Size) {
			targetBuffer = GlobalGpuAllocator.Buffer
			relativeOffset = descriptor.BaseAddress - GlobalGpuAllocator.Base
		} else if descriptor.BaseAddress >= GlobalAllocator.Base && descriptor.BaseAddress < GlobalAllocator.Base+uintptr(GlobalAllocator.Size) {
			targetBuffer = GlobalAllocator.Buffer
			relativeOffset = descriptor.BaseAddress - GlobalAllocator.Base
		} else {
			logger.Printf("Warning: Base address 0x%X is out of known memory bounds", descriptor.BaseAddress)
			continue
		}

		// A resource set to all zeros acts as an unbound texture or buffer (return 0,0,0,0). Buffer Size (in bytes) =
		// (stride==0) ? num_elements : stride * num_elements.
		var rangeBytes uint32
		if descriptor.Stride == 0 {
			rangeBytes = descriptor.Records
		} else {
			rangeBytes = descriptor.Records * uint32(descriptor.Stride)
		}

		// Create the BufferView scoped exactly to the draw call's needs.
		format, formatSize := translateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
		formatStride := uint32(descriptor.Stride)
		if formatStride == 0 {
			formatStride = formatSize
		}

		viewInfo := vk.BufferViewCreateInfo{
			SType:  vk.StructureTypeBufferViewCreateInfo,
			Buffer: targetBuffer,
			Format: format,
			Offset: vk.DeviceSize(relativeOffset),
			Range:  vk.DeviceSize(rangeBytes),
		}

		var view vk.BufferView
		vk.CreateBufferView(t.handles.Device, &viewInfo, nil, &view)

		if descriptor.Records > 0 {
			count := descriptor.Records * (uint32(descriptor.Stride) / 4)
			if count > 64 {
				count = 64
			}
			data := unsafe.Slice((*uint32)(unsafe.Pointer(descriptor.BaseAddress)), count)
			logger.Printf("[%s] Buffer %d (format=%d,%d records=%d stride=%d base=%x) content: %x\n",
				color.Blue.Sprint("GPU"), i, descriptor.DataFormat, descriptor.NumFormat, descriptor.Records, descriptor.Stride, descriptor.BaseAddress, data)
		}

		bufferViews[i] = view
		formatSizes[i] = formatSize
		formatStrides[i] = formatStride
		viewCount++
	}
	if viewCount == 0 {
		return formatSizes, formatStrides
	}

	// Use next pre-allocated descriptor set.
	if t.texelDescriptorSetIndex >= uint32(len(t.texelDescriptorSets)) {
		logger.Printf("Warning: out of pre-allocated descriptor sets (%d) ", t.texelDescriptorSetIndex)
		return formatSizes, formatStrides
	}
	descriptorSet := t.texelDescriptorSets[t.texelDescriptorSetIndex]
	t.texelDescriptorSetIndex++

	// Update descriptor set.
	var writes []vk.WriteDescriptorSet
	for i := range 4 {
		if bufferViews[i] == vk.NullBufferView {
			continue
		}
		writes = append(writes, vk.WriteDescriptorSet{
			SType:            vk.StructureTypeWriteDescriptorSet,
			DstSet:           descriptorSet,
			DstBinding:       uint32(i),
			DescriptorCount:  1,
			DescriptorType:   vk.DescriptorTypeUniformTexelBuffer,
			PTexelBufferView: []vk.BufferView{bufferViews[i]},
		})
	}
	vk.UpdateDescriptorSets(t.handles.Device, uint32(len(writes)), writes, 0, nil)

	// Bind descriptor set.
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, t.stubPipelineLayout, 1, 1, []vk.DescriptorSet{descriptorSet}, 0, nil)

	return formatSizes, formatStrides
}

// Map GCN DataFormat and NumFormat to Vulkan VkFormat and byte size.
func translateGcnFormat(dataFormat, numFormat uint8) (vk.Format, uint32) {
	switch dataFormat {
	case 1: // 8-bit (1 byte)
		switch numFormat {
		case 0:
			return vk.FormatR8Unorm, 1
		case 1:
			return vk.FormatR8Snorm, 1
		case 2:
			return vk.FormatR8Uscaled, 1
		case 3:
			return vk.FormatR8Sscaled, 1
		case 4:
			return vk.FormatR8Uint, 1
		case 5:
			return vk.FormatR8Sint, 1
		}
	case 2: // 16-bit (2 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16Unorm, 2
		case 1:
			return vk.FormatR16Snorm, 2
		case 2:
			return vk.FormatR16Uscaled, 2
		case 3:
			return vk.FormatR16Sscaled, 2
		case 4:
			return vk.FormatR16Uint, 2
		case 5:
			return vk.FormatR16Sint, 2
		case 7:
			return vk.FormatR16Sfloat, 2
		}
	case 3: // 8_8 (2 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR8g8Unorm, 2
		case 1:
			return vk.FormatR8g8Snorm, 2
		case 2:
			return vk.FormatR8g8Uscaled, 2
		case 3:
			return vk.FormatR8g8Sscaled, 2
		case 4:
			return vk.FormatR8g8Uint, 2
		case 5:
			return vk.FormatR8g8Sint, 2
		}
	case 4: // 32-bit (4 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32Uint, 4
		case 5:
			return vk.FormatR32Sint, 4
		case 7:
			return vk.FormatR32Sfloat, 4
		}
	case 5: // 16_16 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16g16Unorm, 4
		case 1:
			return vk.FormatR16g16Snorm, 4
		case 2:
			return vk.FormatR16g16Uscaled, 4
		case 3:
			return vk.FormatR16g16Sscaled, 4
		case 4:
			return vk.FormatR16g16Uint, 4
		case 5:
			return vk.FormatR16g16Sint, 4
		case 7:
			return vk.FormatR16g16Sfloat, 4
		}
	case 6: // 10_11_11 (4 bytes)
		if numFormat == 7 {
			return vk.FormatB10g11r11UfloatPack32, 4
		}
	case 8: // 10_10_10_2 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatA2b10g10r10UnormPack32, 4
		case 4:
			return vk.FormatA2b10g10r10UintPack32, 4
		}
	case 10: // 8_8_8_8 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR8g8b8a8Unorm, 4
		case 1:
			return vk.FormatR8g8b8a8Snorm, 4
		case 2:
			return vk.FormatR8g8b8a8Uscaled, 4
		case 3:
			return vk.FormatR8g8b8a8Sscaled, 4
		case 4:
			return vk.FormatR8g8b8a8Uint, 4
		case 5:
			return vk.FormatR8g8b8a8Sint, 4
		}
	case 11: // 32_32 (8 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32Uint, 8
		case 5:
			return vk.FormatR32g32Sint, 8
		case 7:
			return vk.FormatR32g32Sfloat, 8
		}
	case 12: // 16_16_16_16 (8 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16g16b16a16Unorm, 8
		case 1:
			return vk.FormatR16g16b16a16Snorm, 8
		case 2:
			return vk.FormatR16g16b16a16Uscaled, 8
		case 3:
			return vk.FormatR16g16b16a16Sscaled, 8
		case 4:
			return vk.FormatR16g16b16a16Uint, 8
		case 5:
			return vk.FormatR16g16b16a16Sint, 8
		case 7:
			return vk.FormatR16g16b16a16Sfloat, 8
		}
	case 13: // 32_32_32 (12 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32b32Uint, 12
		case 5:
			return vk.FormatR32g32b32Sint, 12
		case 7:
			return vk.FormatR32g32b32Sfloat, 12
		}
	case 14: // 32_32_32_32 (16 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32b32a32Uint, 16
		case 5:
			return vk.FormatR32g32b32a32Sint, 16
		case 7:
			return vk.FormatR32g32b32a32Sfloat, 16
		}
	}

	panic(fmt.Sprintf("Unhandled GCN Format: data=%d, num=%d", dataFormat, numFormat))
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
