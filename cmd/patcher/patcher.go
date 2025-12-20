package patcher

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/bpfsnoop/gapstone"
	"github.com/gookit/color"
)

var GlobalPatcher = NewPatcher()

// Patcher keeps track of patching state.
type Patcher struct {
	FastDisassembler           gapstone.Engine
	DetailedDisassembler       gapstone.Engine
	NeededTcbAccessTrampolines []gapstone.Instruction
}

// NewPatcher creates a new instance of Patcher.
func NewPatcher() *Patcher {
	var err error
	p := &Patcher{}
	p.FastDisassembler, err = gapstone.New(gapstone.CS_ARCH_X86, gapstone.CS_MODE_64)
	if err != nil {
		panic(err)
	}
	p.DetailedDisassembler, err = gapstone.New(gapstone.CS_ARCH_X86, gapstone.CS_MODE_64)
	if err != nil {
		panic(err)
	}
	if err = p.DetailedDisassembler.SetOption(gapstone.CS_OPT_DETAIL, gapstone.CS_OPT_ON); err != nil {
		panic(err)
	}

	return p
}

// Patch patches the ELF file.
func (p *Patcher) Patch(e *elf.Elf) {
	sys_struct.TlsOnce.Do(sys_struct.AllocTlsSlot)
	if sys_struct.TlsSlot >= 64 {
		panic("TLS slot is too high, cannot patch TCB access")
	}
	if e.Name != "libkernel.sprx" {
		color.Gray.Printf("Skipping %s patching...\n", e.Name)
		return
	}

	// Do a quick pass and patch instructions that we can, otherwise store them for later.
	patchCount := 0
	p.NeededTcbAccessTrampolines = []gapstone.Instruction{}
	for _, s := range e.LoadSections {
		if (s.PFlags & elf.PF_X) == 0 {
			continue
		}

		sectionStart := s.PVaddr
		sectionEnd := s.PVaddr + s.PFilesz
		if sectionEnd > uint64(len(e.Memory)) {
			sectionEnd = uint64(len(e.Memory))
		}
		sectionData := e.Memory[int(sectionStart):int(sectionEnd)]

		instructions, err := p.FastDisassembler.Disasm(sectionData, sectionStart, 0)
		if err != nil {
			fmt.Printf(
				"Failed to disassemble %s: %v\n",
				color.Red.Sprint(e.Name),
				color.Red.Sprint(err.Error()),
			)
			continue
		}
		for _, instruction := range instructions {
			if instruction.Mnemonic != "mov" {
				continue
			}
			instructionData := e.Memory[int(instruction.Address) : int(instruction.Address)+len(instruction.Bytes)]
			detailedInstruction, err := p.DetailedDisassembler.Disasm(instructionData, uint64(instruction.Address), 1)
			if err != nil {
				fmt.Printf(
					"Failed to disassemble %s: %v\n",
					color.Red.Sprint(e.Name),
					color.Red.Sprint(err.Error()),
				)
			}
			if p.PatchTcbAccess(detailedInstruction[0], instructionData) {
				patchCount++
			}
		}
	}

	// Process trampoline candidates.
	for _, inst := range p.NeededTcbAccessTrampolines {
		if p.CreateTcbAccessTrampoline(e, inst) {
			patchCount++
		}
	}

	fmt.Printf("Patched %s instructions.\n", color.Yellow.Sprintf("%d", patchCount))
}

// PatchTcbAccess patches a TCB access instruction or adds it to trampoline list if unable to.
func (p *Patcher) PatchTcbAccess(instruction gapstone.Instruction, instructionBytes []byte) bool {
	// We are on Windows, so we need to patch fs to gs.
	// The instruction is mov ?, fs:[?].
	if instruction.Mnemonic != "mov" || len(instruction.X86.Operands) != 2 {
		return false
	}
	op := instruction.X86.Operands[1]
	if op.Type != gapstone.X86_OP_MEM || op.Mem.Segment != gapstone.X86_REG_FS {
		return false
	}
	if len(instruction.Bytes) < 5 {
		fmt.Printf("Failed to patch %s-byte TCB access.\n", color.Red.Sprintf("%d", len(instruction.Bytes)))
		return false
	}

	// Only patch if displacement is 0, otherwise use a trampoline.
	if op.Mem.Disp != 0 {
		p.NeededTcbAccessTrampolines = append(p.NeededTcbAccessTrampolines, instruction)
		return false
	}

	// Find the prefix 0x64 (FS)
	prefixOffset := -1
	for j := 0; j < len(instruction.Bytes); j++ {
		if instruction.Bytes[j] == 0x64 {
			prefixOffset = j
			break
		}
	}
	if prefixOffset == -1 {
		fmt.Printf("Failed to find %s prefix for TCB access.\n", color.Red.Sprint("FS"))
		return false
	}

	// Patch prefix to 0x65 (GS)
	instructionBytes[prefixOffset] = 0x65

	// Calculate new displacement (TlsSlotsOffset = 0x1480)
	newDisplacement := uint32(0x1480 + sys_struct.TlsSlot*8)

	// Patch displacement (last 4 bytes)
	if len(instruction.Bytes) >= 5 {
		displacementOffset := len(instruction.Bytes) - 4
		binary.LittleEndian.PutUint32(instructionBytes[displacementOffset:], newDisplacement)
		return true
	}

	/* fmt.Printf(
		"Patched fs TCB access at %s.\n",
		color.Yellow.Sprintf("0x%X", instruction.Address),
	) */

	return false
}

// CreateTcbAccessTrampoline creates trampoline for a TCB access instruction, when displacement is not 0.
func (p *Patcher) CreateTcbAccessTrampoline(e *elf.Elf, instruction gapstone.Instruction) bool {
	dstReg := instruction.X86.Operands[0].Reg
	displacement := instruction.X86.Operands[1].Mem.Disp
	scratchReg := uint(gapstone.X86_REG_R11)
	if dstReg == scratchReg {
		scratchReg = uint(gapstone.X86_REG_R10)
	}
	realInstructionAddr := uint64(e.BaseAddress) + uint64(instruction.Address)
	returnAddr := realInstructionAddr + uint64(len(instruction.Bytes))

	// Create trampoline for TCB access with displacement.
	trampolineAsm := newAsmHelper()
	trampolineAsm.mov_r64_from_gs_mem(scratchReg, int32(0x1480+sys_struct.TlsSlot*8))
	trampolineAsm.add_r64_imm32(scratchReg, int32(displacement))
	trampolineAsm.mov_r64_from_mem(dstReg, scratchReg, 0)
	trampoline := trampolineAsm.bytes()
	trampolineSize := len(trampoline) + 5 // 5 bytes for jmp rel32
	trampolineAddr, _ := sys_struct.AllocExecututableMemory(uintptr(trampolineSize))

	jumpBackAsm := newAsmHelper()
	jumpBackSourceAddr := uint64(trampolineAddr) + uint64(len(trampoline))
	jumpBackAsm.jmp_rel32(returnAddr, jumpBackSourceAddr)
	jumpBackCode := jumpBackAsm.bytes()

	trampoline = append(trampoline, jumpBackCode...)
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(trampolineAddr)), trampolineSize),
		trampoline,
	)

	// Patch original instruction to make a relative jump to our trampoline.
	patch := make([]byte, len(instruction.Bytes))
	for i := range patch {
		patch[i] = 0x90 // NOP
	}
	patch[0] = 0xE9 // JMP rel32
	rel32Jump := int32(uint64(trampolineAddr) - (realInstructionAddr + 5))
	binary.LittleEndian.PutUint32(patch[1:], uint32(rel32Jump))
	copy(e.Memory[instruction.Address:], patch)

	/* fmt.Printf(
		"Patched fs:%s TCB access at %s (trampolined to %s).\n",
		color.Yellow.Sprintf("0x%X", displacement),
		color.Yellow.Sprintf("0x%X", instruction.Address),
		color.Yellow.Sprintf("0x%X", trampolineAddr),
	) */

	return true
}
