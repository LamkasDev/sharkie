package patcher

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
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
	if runtime.GOOS == "windows" {
		instructionBytes[prefixOffset] = 0x65
	}

	// Patch displacement (last 4 bytes)
	if len(instruction.Bytes) >= 5 {
		displacementOffset := len(instruction.Bytes) - 4
		binary.LittleEndian.PutUint32(instructionBytes[displacementOffset:], uint32(asm.PlaystationTlsOffset))
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
	if runtime.GOOS == "windows" {
		trampolineAsm.mov_r64_from_gs_mem(scratchReg, int32(asm.PlaystationTlsOffset))
	} else {
		trampolineAsm.mov_r64_from_fs_mem(scratchReg, int32(asm.PlaystationTlsOffset))
	}
	trampolineAsm.mov_r64_from_mem(dstReg, scratchReg, int32(displacement))
	trampolineCode := trampolineAsm.bytes()
	trampolineSize := len(trampolineCode) + 5 // 5 bytes for jmp rel32
	trampolineAddr, _ := sys_struct.AllocExecutableMemory(uintptr(trampolineSize))

	jumpBackAsm := newAsmHelper()
	jumpBackSourceAddr := uint64(trampolineAddr) + uint64(len(trampolineCode))
	jumpBackAsm.jmp_rel32(returnAddr, jumpBackSourceAddr)
	jumpBackCode := jumpBackAsm.bytes()

	trampoline := append(trampolineCode, jumpBackCode...)
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

	logger.Printf(
		"Patched fs:%s TCB access at %s (trampolined to %s).\n",
		color.Yellow.Sprintf("0x%X", displacement),
		color.Yellow.Sprintf("0x%X", instruction.Address),
		color.Yellow.Sprintf("0x%X", trampolineAddr),
	)
	return true
}
