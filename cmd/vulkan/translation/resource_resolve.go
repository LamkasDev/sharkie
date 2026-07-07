package translation

import (
	"unsafe"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	spirvGcn "github.com/LamkasDev/sharkie/cmd/spirv/gcn"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	"go101.org/nstd"
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

func applySOP1(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	if details.Dst >= uint32(len(registers)) {
		return
	}
	dst := int(details.Dst - gcnSpec.OpSgpr0)

	switch details.Op {
	case gcnSpec.Sop1OpMovB32:
		registers[dst] = readScalarOperand(details.Src0, instr, registers)
	case gcnSpec.Sop1OpAndSaveexecB64:
		registers[dst] = registers[gcnSpec.OpExecLo]
		registers[dst+1] = registers[gcnSpec.OpExecHi]

		src0Lo := readScalarOperand(details.Src0, instr, registers)
		src0Hi := readScalarOperand(details.Src0+1, instr, registers)
		resLo := src0Lo & registers[gcnSpec.OpExecLo]
		resHi := src0Hi & registers[gcnSpec.OpExecHi]
		registers[gcnSpec.OpExecLo] = resLo
		registers[gcnSpec.OpExecHi] = resHi
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi((resLo | resHi) != 0))
	}
}

func applySOP2(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	if details.Dst >= uint32(len(registers)) {
		return
	}
	dst := int(details.Dst - gcnSpec.OpSgpr0)
	src0 := readScalarOperand(details.Src0, instr, registers)
	src1 := readScalarOperand(details.Src1, instr, registers)

	switch details.Op {
	case gcnSpec.Sop2OpCselectB32:
		registers[dst] = src0
		if registers[gcnSpec.OpScc] != 0 {
			registers[dst] = src1
		}
	case gcnSpec.Sop2OpLshlB32:
		res := src0 << (src1 & 31)
		registers[dst] = res
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi(res != 0))
	case gcnSpec.Sop2OpAndB32:
		res := src0 & src1
		registers[dst] = res
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi(res != 0))
	case gcnSpec.Sop2OpAndB64:
		src0Hi := readScalarOperand(details.Src0+1, instr, registers)
		src1Hi := readScalarOperand(details.Src1+1, instr, registers)
		resLo := src0 & src1
		resHi := src0Hi & src1Hi
		registers[dst] = resLo
		registers[dst+1] = resHi
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi((resLo | resHi) != 0))
	case gcnSpec.Sop2OpBfeU32:
		offset := src1 & 0x7F
		width := (src1 >> 16) & 0x7F
		var res uint32
		if width == 0 {
			res = 0
		} else if width+offset <= 32 {
			// Extract bits [offset : offset+width-1].
			res = (src0 >> offset) & ((1 << width) - 1)
		} else {
			// Shift right by offset.
			res = src0 >> offset
		}
		registers[dst] = res
		registers[gcnSpec.OpScc] = uint32(nstd.Btoi(res != 0))
	}
}

func readScalarOperand(op uint32, instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) uint32 {
	switch {
	case op <= gcnSpec.OpExecHi:
		return registers[op]
	case op >= gcnSpec.OpInt0 && op <= gcnSpec.OpPosInt64:
		return op - gcnSpec.OpInt0
	case op >= gcnSpec.OpNegInt1 && op <= gcnSpec.OpNegInt16:
		panic("unhandled")
	case op >= gcnSpec.OpFloat05 && op <= gcnSpec.OpFloatNeg40:
		panic("unhandled")
	case op >= gcnSpec.OpVccz && op <= gcnSpec.OpScc:
		return registers[op]
	case op == gcnSpec.OpLiteral && instr.HasLiteral:
		return instr.Literal
	case op >= gcnSpec.OpVgpr0:
		panic("unhandled")
	}

	return 0
}

func applySMRD(instr *gcnSpec.Instruction, registers *gcnSpec.GcnRegisters) {
	details := instr.Details.(*gcnSpec.SmrdDetails)
	if details.Dst >= uint32(len(registers)) {
		return
	}
	count := spirvGcn.SmrdLoadDwordCount(details.Op)

	var dwords []uint32
	switch {
	case details.Op >= gcnSpec.SmrdOpBufferLoadDword && details.Op <= gcnSpec.SmrdOpBufferLoadDwordx16:
		base := details.Base * 2
		address := uintptr(registers[base]) | (uintptr(registers[base+1]&0xFFFF) << 32)
		dwords = unsafe.Slice((*uint32)(unsafe.Pointer(address)), count)
	default:
		base := details.Base * 2
		address := uintptr(registers[base]) | (uintptr(registers[base+1]) << 32)
		offset := uintptr(0)
		if details.ImmOff {
			if instr.HasLiteral {
				offset = uintptr(instr.Literal)
			} else {
				offset = uintptr(details.Offset * 4)
			}
		}
		dwords = unsafe.Slice((*uint32)(unsafe.Pointer(address+offset)), count)
	}

	dst := details.Dst
	for i := uint32(0); i < count && int(dst+i) < len(registers); i++ {
		if int(i) < len(dwords) {
			registers[dst+i] = dwords[i]
		} else {
			registers[dst+i] = 0
		}
	}
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
