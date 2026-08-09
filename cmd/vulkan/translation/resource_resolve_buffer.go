package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

// ResolvedBufferAccess is a T# resolved at a specific MUBUF/MTBUF instruction for a user-data snapshot.
type ResolvedBufferAccess struct {
	Instruction *gcnSpec.Instruction
	Kind        BufferAccessKind
	Descriptor  spirvStructs.BufferDescriptor
}

// ResolveBufferResources simulates scalar SGPR updates and resolves T# descriptors.
func (t *GpuTranslator) ResolveBufferResources(shader *spirv.SpirvShader, userData []uint32) []ResolvedBufferAccess {
	stageBase := spirvStructs.GcnStageToUserDataOffset[shader.GcnShader.Stage]
	registers := gcnSpec.GcnRegisters{}
	for i := range 16 {
		offset := int(stageBase) + i
		registers[i] = userData[offset]
	}

	var accesses []ResolvedBufferAccess
	rpo := shader.GcnShader.Cfg.ReversePostOrder()
	for _, blockId := range rpo {
		block := &shader.GcnShader.Cfg.Blocks[blockId]
		for i := range block.Instructions {
			instr := &block.Instructions[i]
			accesses = t.resolveBufferResourcesIns(shader.GcnShader, instr, &registers, accesses)
		}
	}

	return accesses
}

func (t *GpuTranslator) resolveBufferResourcesIns(shader *gcn.GcnShader, instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters, accesses []ResolvedBufferAccess) []ResolvedBufferAccess {
	switch instr.Encoding {
	case gcnSpec.EncSOP1:
		details := instr.Details.(*gcnSpec.ScalarDetails)
		if details.Op == gcnSpec.Sop1OpSwappcB64 {
			fetchPCLo := registers[details.Src0]
			fetchPCHi := registers[details.Src0+1]
			fetchShaderAddress := uintptr(fetchPCLo) | (uintptr(fetchPCHi) << 32)
			if fetchShaderAddress != 0 {
				fetchShader := gpu.GlobalLiverpool.GetShader(gcn.GcnShaderStageFetch, fetchShaderAddress)
				if fetchShader != nil {
					fetchInstructions := ParseFetchShaderInstructions(fetchShader)
					for _, fetchInstr := range fetchInstructions {
						accesses = t.resolveBufferResourcesIns(shader, fetchInstr, registers, accesses)
					}
				}
			}
		}
		applySOP1(shader, instr, registers)
	case gcnSpec.EncSOP2:
		applySOP2(instr, registers)
	case gcnSpec.EncSOPC:
		applySOPC(instr, registers)
	case gcnSpec.EncSMRD:
		t.applySMRD(instr, registers)
	case gcnSpec.EncMUBUF:
		accesses = append(accesses, resolveMUBUF(instr, registers))
	case gcnSpec.EncMTBUF:
		accesses = append(accesses, resolveMTBUF(instr, registers))
	}

	return accesses
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
		t.imagesMutex.Lock()
		image, ok := t.images[access.Descriptor.BaseAddress]
		t.imagesMutex.Unlock()
		if !ok {
			continue
		}
		seen[access.Descriptor.BaseAddress] = struct{}{}
		images = append(images, image)
	}

	return images, nil
}
