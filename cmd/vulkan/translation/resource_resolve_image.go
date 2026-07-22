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

// ResolveImageResources simulates scalar SGPR updates, then resolves T# descriptors at
// precomputed MIMG sites from AnalyzeResources.
func ResolveImageResources(sites []SpirvShaderResource, shader *spirv.SpirvShader, userData []uint32) []ResolvedImageAccess {
	siteByOffset := make(map[uintptr]SpirvShaderResource, len(sites))
	for _, site := range sites {
		siteByOffset[site.InstructionOffset] = site
	}

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
			switch instr.Encoding {
			case gcnSpec.EncSOP1:
				applySOP1(instr, &registers)
			case gcnSpec.EncSOP2:
				applySOP2(instr, &registers)
			case gcnSpec.EncSOPC:
				applySOPC(instr, &registers)
			case gcnSpec.EncSMRD:
				applySMRD(instr, &registers)
			case gcnSpec.EncMIMG:
				site := siteByOffset[instr.DwordOffset]
				accesses = append(accesses, resolveMIMG(instr, &registers, site.Kind))
			}
		}
	}

	return accesses
}

func resolveMIMG(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters, kind ImageAccessKind) ResolvedImageAccess {
	details := instr.Details.(*gcnSpec.MimgDetails)
	start := int(details.Srsrc * 4)

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

// StoreTargets returns base addresses of images written by store ops.
func (t *GpuTranslator) StoreTargets(accesses []ResolvedImageAccess) ([]*vulkan.VulkanImage, error) {
	seen := map[uintptr]struct{}{}
	var images []*vulkan.VulkanImage
	for _, access := range accesses {
		if access.Kind != ImageAccessStore && access.Kind != ImageAccessStoreMip {
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
