package patcher

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/bpfsnoop/gapstone"
	"github.com/gookit/color"
)

const TcbAccessNoPatch = 0
const TcbAccessDirect = 1
const TcbAccessTrampoline = 2

// FilterTcbAccess checks if instruction is TCB access and optionally adds it to trampoline list.
func (p *Patcher) FilterTcbAccess(instruction gapstone.Instruction) int {
	// We are on Windows, so we need to patch fs to gs.
	// The instruction is mov ?, fs:[?].
	if instruction.Mnemonic != "mov" || len(instruction.X86.Operands) != 2 {
		return TcbAccessNoPatch
	}
	op := instruction.X86.Operands[1]
	if op.Type != gapstone.X86_OP_MEM || op.Mem.Segment != gapstone.X86_REG_FS {
		return TcbAccessNoPatch
	}
	if len(instruction.Bytes) < 5 {
		logger.Printf(
			"Failed to patch %s-byte TCB access.\n",
			color.Red.Sprintf("%d", len(instruction.Bytes)),
		)
		return TcbAccessNoPatch
	}

	// Only patch if displacement is 0, otherwise use a trampoline.
	if op.Mem.Disp != 0 {
		if op.Mem.Disp != 0x10 {
			logger.Print(color.Gray.Sprintf(
				"Unknown displacement 0x%X for TCB access at 0x%X, skipping...\n",
				op.Mem.Disp,
				instruction.Address,
			))
			return TcbAccessNoPatch
		}
		return TcbAccessTrampoline
	}

	return TcbAccessDirect
}

// PatchTcbAccess patches a TCB access instruction with 0 displacement.
func (p *Patcher) PatchTcbAccess(instruction gapstone.Instruction, instructionBytes []byte) bool {
	// Find the prefix 0x64 (FS)
	prefixOffset := -1
	for j := 0; j < len(instruction.Bytes); j++ {
		if instruction.Bytes[j] == 0x64 {
			prefixOffset = j
			break
		}
	}
	if prefixOffset == -1 {
		logger.Printf(
			"Failed to find %s prefix for TCB access.\n",
			color.Red.Sprint("FS"),
		)
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
	}

	/* logger.Printf(
		"Patched fs TCB access at %s.\n",
		color.Yellow.Sprintf("0x%X", instruction.Address),
	) */
	return true
}

// CreateTcbAccessTrampoline creates trampoline for a TCB access instruction with non-zero displacement.
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

	/* logger.Printf(
		"Patched fs:%s TCB access at %s (trampolined to %s).\n",
		color.Yellow.Sprintf("0x%X", displacement),
		color.Yellow.Sprintf("0x%X", instruction.Address),
		color.Yellow.Sprintf("0x%X", trampolineAddr),
	) */
	return true
}
