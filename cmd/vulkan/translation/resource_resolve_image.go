package translation

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

// ResolvedImageAccess is a T# resolved at a specific MIMG instruction for a user-data snapshot.
type ResolvedImageAccess struct {
	InstructionOffset uintptr
	Kind              ImageAccessKind
	Descriptor        spirvStructs.ImageDescriptor
	Sampler           *spirvStructs.SamplerDescriptor
}

// ResolveImageResources simulates scalar SGPR updates and resolves T# descriptors.
func ResolveImageResources(shader *spirv.SpirvShader, userData []uint32) []ResolvedImageAccess {
	stageBase := spirvStructs.GcnStageToUserDataOffset[shader.GcnShader.Stage]
	registers := gcnSpec.GcnRegisters{}
	for i := range 16 {
		offset := int(stageBase) + i
		registers[i] = userData[offset]
	}

	var accesses []ResolvedImageAccess
	rpo := shader.GcnShader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.GcnShader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instr := &block.Instructions[i]
			accesses = resolveImageResourcesIns(instr, &registers, accesses)
		}
	}

	return accesses
}

func resolveImageResourcesIns(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters, accesses []ResolvedImageAccess) []ResolvedImageAccess {
	switch instr.Encoding {
	case gcnSpec.EncSOP1:
		applySOP1(instr, registers)
	case gcnSpec.EncSOP2:
		applySOP2(instr, registers)
	case gcnSpec.EncSOPC:
		applySOPC(instr, registers)
	case gcnSpec.EncSMRD:
		applySMRD(instr, registers)
	case gcnSpec.EncMIMG:
		accesses = append(accesses, resolveMIMG(instr, registers))
	}

	return accesses
}

func resolveMIMG(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) ResolvedImageAccess {
	details := instr.Details.(*gcnSpec.MimgDetails)
	start := int(details.Srsrc * 4)
	kind := spirv.MimgAccessKind(details.Op)

	var imageDwords [8]uint32
	for i := range 8 {
		imageDwords[i] = registers[start+i]
	}
	descriptor := spirvStructs.NewImageDescriptor(imageDwords[:])

	var samplerDescriptor *spirvStructs.SamplerDescriptor
	var samplerDwords [4]uint32
	if kind == ImageAccessSample {
		samplerStart := int(details.Ssamp * 4)
		for i := range 4 {
			samplerDwords[i] = registers[samplerStart+i]
		}
		samplerDescriptor = spirvStructs.NewSamplerDescriptor(samplerDwords[:])
	}

	return ResolvedImageAccess{
		InstructionOffset: instr.DwordOffset,
		Kind:              kind,
		Descriptor:        descriptor,
		Sampler:           samplerDescriptor,
	}
}

// ResolveImageTargets returns images that are read/written/sampled from.
func (t *GpuTranslator) ResolveImageTargets(accesses []ResolvedImageAccess, kind ImageAccessKind) ([]*vulkan.VulkanImage, error) {
	seen := map[uintptr]struct{}{}
	var images []*vulkan.VulkanImage
	for _, access := range accesses {
		if access.Kind != kind {
			continue
		}
		if _, ok := seen[access.Descriptor.BaseAddress]; ok {
			continue
		}
		view, err, _ := t.GetImageView(access.Descriptor)
		if err != nil {
			return nil, err
		}
		seen[access.Descriptor.BaseAddress] = struct{}{}
		images = append(images, view.Image)
	}

	return images, nil
}
