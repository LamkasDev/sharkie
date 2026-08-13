package translation

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

type ResolvedImageAccess struct {
	Instruction *gcnSpec.Instruction
	Kind        ImageAccessKind
	Descriptor  spirvStructs.ImageDescriptor
	Sampler     *spirvStructs.SamplerDescriptor
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
		Instruction: instr,
		Kind:        kind,
		Descriptor:  descriptor,
		Sampler:     samplerDescriptor,
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
