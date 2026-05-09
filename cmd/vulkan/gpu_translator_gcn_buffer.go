package vulkan

import (
	"fmt"
	"unsafe"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
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
		buffer, mem, err := t.AllocBuffer(size,
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
		logger.Printf("[%s] Created user data with hash %s (vtx=%x, frag=%x).\n",
			color.Blue.Sprint("GPU"),
			color.Yellow.Sprintf("0x%X", hash),
			contents[UserDataOffsetVertex:UserDataOffsetHull],
			contents[UserDataOffsetFragment:UserDataOffsetCompute],
		)
	}
}

func (t *GpuTranslator) BindTexelBuffers(commandBuffer vk.CommandBuffer, draw *LiverpoolDrawCall, userData []uint32) ([4]uint32, [4]uint32) {
	var bufferViews [4]vk.BufferView
	var viewCount uint32
	var formatSizes, formatStrides [4]uint32
	for i := range 4 {
		sgprBase := i * 4
		descriptor := spirv.NewBufferDescriptor(
			userData[sgprBase],
			userData[sgprBase+1],
			userData[sgprBase+2],
			userData[sgprBase+3],
		)
		if descriptor.BaseAddress == 0 && descriptor.Records == 0 {
			continue
		}

		// Route to the correct buffer based on address range.
		targetBuffer, relativeOffset, err := t.GetBufferFromAddress(descriptor.BaseAddress)
		if err != nil {
			panic(err)
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
		format, formatSize := TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
		formatStride := uint32(descriptor.Stride)
		if formatStride == 0 {
			formatStride = formatSize
		}

		// Check if format supports texel buffers.
		var props vk.FormatProperties
		vk.GetPhysicalDeviceFormatProperties(t.handles.PhysicalDevice, format, &props)
		props.Deref()
		if (props.BufferFeatures & vk.FormatFeatureFlags(vk.FormatFeatureUniformTexelBufferBit)) == 0 {
			// logger.Printf("  warning: format %d does not support uniform texel buffers, skipping.\n", format)
			formatSizes[i] = formatSize
			formatStrides[i] = formatStride
			continue
		}

		// Align offset to 16 bytes (common minTexelBufferOffsetAlignment).
		// TODO: query this from physical device limits.
		alignedOffset := relativeOffset & ^uintptr(15)
		padding := uint32(relativeOffset - alignedOffset)
		alignedRange := rangeBytes + padding

		// Align range to format size.
		if formatSize > 0 {
			alignedRange = (alignedRange / formatSize) * formatSize
		}

		viewInfo := vk.BufferViewCreateInfo{
			SType:  vk.StructureTypeBufferViewCreateInfo,
			Buffer: targetBuffer,
			Format: format,
			Offset: vk.DeviceSize(alignedOffset),
			Range:  vk.DeviceSize(alignedRange),
		}

		var view vk.BufferView
		result := vk.CreateBufferView(t.handles.Device, &viewInfo, nil, &view)
		if err := as.NewError(result); err != nil {
			// logger.Printf("  failed to create buffer view: %v (format=%d offset=0x%X range=0x%X)\n", err, format, alignedOffset, rangeBytes)
			view = vk.NullBufferView
		}

		if descriptor.Records > 0 {
			count := descriptor.Records * (uint32(descriptor.Stride) / 4)
			if count > 64 {
				count = 64
			}
			data := unsafe.Slice((*uint32)(unsafe.Pointer(descriptor.BaseAddress)), count)
			logger.Printf("[%s] Buffer %d (format=%d,%d records=%d stride=%d base=%x) content: %x\n",
				color.Blue.Sprint("GPU"), i,
				descriptor.DataFormat, descriptor.NumFormat, descriptor.Records, descriptor.Stride, descriptor.BaseAddress,
				data,
			)
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

func (t *GpuTranslator) GetBufferFromAddress(address uintptr) (vk.Buffer, uintptr, error) {
	if address >= GlobalGpuAllocator.Base && address < GlobalGpuAllocator.Base+uintptr(GlobalGpuAllocator.Size) {
		return GlobalGpuAllocator.Buffer, address - GlobalGpuAllocator.Base, nil
	} else if address >= GlobalAllocator.Base && address < GlobalAllocator.Base+uintptr(GlobalAllocator.Size) {
		return GlobalAllocator.Buffer, address - GlobalAllocator.Base, nil
	}

	return vk.NullBuffer, 0, fmt.Errorf("address 0x%X is out of known memory bounds", address)
}
