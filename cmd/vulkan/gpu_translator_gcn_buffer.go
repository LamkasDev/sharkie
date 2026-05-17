package vulkan

import (
	"fmt"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
)

// TODO: resource tracking.
func (t *GpuTranslator) BindTexelBuffers(commandBuffer vk.CommandBuffer, userData []uint32, stage gcn.GcnShaderStage, setIndex uint32, bindPoint vk.PipelineBindPoint) ([4]uint32, [4]uint32) {
	var bufferViews [4]vk.BufferView
	var viewCount uint32
	var formatSizes, formatStrides [4]uint32
	offset := gpu.GcnStageToUserDataOffset[stage]
	for i := range 4 {
		sgprBase := offset + uint32(i*4)
		descriptor := spirvStructs.NewBufferDescriptor(
			userData[sgprBase],
			userData[sgprBase+1],
			userData[sgprBase+2],
			userData[sgprBase+3],
		)

		// Route to the correct buffer based on address range.
		targetBuffer, relativeOffset, err := t.GetBufferFromAddress(descriptor.BaseAddress)
		if err != nil {
			continue
		}

		// A resource set to all zeros acts as an unbound texture or buffer (return 0,0,0,0). Buffer Size (in bytes) =
		// (stride==0) ? num_elements : stride * num_elements.
		var rangeBytes uint64
		if descriptor.Stride == 0 {
			rangeBytes = uint64(descriptor.Records)
		} else {
			rangeBytes = uint64(descriptor.Records) * uint64(descriptor.Stride)
		}

		// Create the BufferView scoped exactly to the draw call's needs.
		format, formatSize := TranslateGcnFormat(descriptor.DataFormat, descriptor.NumFormat)
		formatStride := uint32(descriptor.Stride)
		if formatStride == 0 {
			formatStride = formatSize
		}

		// TODO: physical device format support, alignment.

		viewInfo := vk.BufferViewCreateInfo{
			SType:  vk.StructureTypeBufferViewCreateInfo,
			Buffer: targetBuffer,
			Format: format,
			Offset: vk.DeviceSize(relativeOffset),
			Range:  vk.DeviceSize(rangeBytes),
		}

		var view vk.BufferView
		result := vk.CreateBufferView(t.handles.Device, &viewInfo, nil, &view)
		if err = as.NewError(result); err != nil {
			// logger.Printf("  failed to create buffer view: %v (format=%d offset=0x%X range=0x%X)\n", err, format, alignedOffset, rangeBytes)
			view = vk.NullBufferView
		}

		if descriptor.Records > 0 {
			count := descriptor.Records * (uint32(descriptor.Stride) / 4)
			if count > 64 {
				count = 64
			}
			/* data := unsafe.Slice((*uint32)(unsafe.Pointer(descriptor.BaseAddress)), count)
			logger.Printf("[%s] Buffer %d for %s (format=%d,%d records=%d stride=%d base=%x) content: %x\n",
				color.Blue.Sprint("GPU"), i,
				stage,
				descriptor.DataFormat, descriptor.NumFormat, descriptor.Records, descriptor.Stride, descriptor.BaseAddress,
				data,
			) */
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
	vk.CmdBindDescriptorSets(commandBuffer, bindPoint, t.pipelineLayout, setIndex, 1, []vk.DescriptorSet{descriptorSet}, 0, nil)

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
