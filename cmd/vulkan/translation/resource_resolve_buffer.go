package translation

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

type ResolvedBufferAccess struct {
	Instruction *gcnSpec.Instruction
	Kind        BufferAccessKind
	Descriptor  spirvStructs.BufferDescriptor
}

func resolveMUBUF(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) ResolvedBufferAccess {
	details := instr.Details.(*gcnSpec.MubufDetails)
	start := int(details.Srsrc * 4)

	var bufferDwords [4]uint32
	for j := range 4 {
		bufferDwords[j] = registers[start+j]
	}
	descriptor := spirvStructs.NewBufferDescriptor(bufferDwords[:])

	return ResolvedBufferAccess{
		Instruction: instr,
		Kind:        spirv.MubufAccessKind(details.Op),
		Descriptor:  descriptor,
	}
}

func resolveMTBUF(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) ResolvedBufferAccess {
	details := instr.Details.(*gcnSpec.MtbufDetails)
	start := int(details.Srsrc * 4)

	var bufferDwords [4]uint32
	for j := range 4 {
		bufferDwords[j] = registers[start+j]
	}
	descriptor := spirvStructs.NewBufferDescriptor(bufferDwords[:])

	return ResolvedBufferAccess{
		Instruction: instr,
		Kind:        spirv.MtbufAccessKind(details.Op),
		Descriptor:  descriptor,
	}
}

// ResolveBufferTargets returns images that are read/written to.
func (t *GpuTranslator) ResolveBufferTargets(accesses []ResolvedBufferAccess, kind BufferAccessKind) ([]*vulkan.VulkanImage, error) {
	seen := map[uintptr]struct{}{}
	var images []*vulkan.VulkanImage
	for _, access := range accesses {
		if access.Kind != kind {
			continue
		}
		if _, ok := seen[access.Descriptor.BaseAddress]; ok {
			continue
		}
		t.imageGroupsMutex.Lock()
		group, ok := t.imageGroups[access.Descriptor.BaseAddress]
		t.imageGroupsMutex.Unlock()
		if !ok {
			continue
		}
		seen[access.Descriptor.BaseAddress] = struct{}{}
		images = append(images, group.LeadingImage)
	}

	return images, nil
}
