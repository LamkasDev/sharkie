package patcher

import (
	"fmt"
	"unsafe"
)

// DecodeTcbAccess disassembles the instruction at RIP and returns the destination register, displacement, and length.
func DecodeTcbAccess(rip uint64) (dstReg uint, displacement int32, instructionLen uint64, err error) {
	instructionData := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(rip))), 15)
	instructions, disasmErr := GlobalPatcher.DetailedDisassembler.Disasm(instructionData, rip, 1)
	if disasmErr != nil || len(instructions) == 0 {
		return 0, 0, 0, fmt.Errorf("failed to disassemble instruction at 0x%X", rip)
	}

	instruction := instructions[0]
	if instruction.Mnemonic != "mov" || len(instruction.X86.Operands) != 2 {
		return 0, 0, 0, fmt.Errorf("invalid TCB access instruction at 0x%X", rip)
	}

	dstReg = instruction.X86.Operands[0].Reg
	displacement = int32(instruction.X86.Operands[1].Mem.Disp)
	instructionLen = uint64(len(instruction.Bytes))
	return dstReg, displacement, instructionLen, nil
}
