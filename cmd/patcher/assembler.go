package patcher

import (
	"encoding/binary"

	"github.com/bpfsnoop/gapstone"
)

// asmHelper helps build small assembly snippets for trampolines.
type asmHelper struct {
	buf []byte
}

func newAsmHelper() *asmHelper {
	return &asmHelper{buf: []byte{}}
}

// mov_r64_from_gs_mem adds `mov dst, gs:[disp32]`
func (a *asmHelper) mov_r64_from_gs_mem(dst uint, disp int32) {
	a.buf = append(a.buf, 0x65) // GS prefix

	rex := byte(0x48)
	dstCode := uint8(dst - gapstone.X86_REG_RAX)
	if dstCode >= 8 {
		rex |= 4 // REX.R
	}
	a.buf = append(a.buf, rex)
	a.buf = append(a.buf, 0x8B)

	regField := (dstCode & 7) << 3
	modRM := byte(0b00_000_100) | regField
	a.buf = append(a.buf, modRM)

	sib := byte(0b00_100_101)
	a.buf = append(a.buf, sib)

	dispBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(dispBytes, uint32(disp))
	a.buf = append(a.buf, dispBytes...)
}

// mov_r64_from_mem adds `mov dst, [base + disp32]`
func (a *asmHelper) mov_r64_from_mem(dst, base uint, disp int32) {
	rex := byte(0x48)
	dstCode := uint8(dst - gapstone.X86_REG_RAX)
	baseCode := uint8(base - gapstone.X86_REG_RAX)

	if dstCode >= 8 {
		rex |= 4 // REX.R
	}
	if baseCode >= 8 {
		rex |= 1 // REX.B
	}
	a.buf = append(a.buf, rex)
	a.buf = append(a.buf, 0x8B)

	mod := byte(0b10)
	regField := (dstCode & 7) << 3
	rmField := baseCode & 7
	modRM := (mod << 6) | regField | rmField
	a.buf = append(a.buf, modRM)

	if (baseCode & 7) == 4 {
		sib := byte(0b00_100_100)
		a.buf = append(a.buf, sib)
	}

	dispBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(dispBytes, uint32(disp))
	a.buf = append(a.buf, dispBytes...)
}

// add_r64_imm32 adds `add reg, imm32`
func (a *asmHelper) add_r64_imm32(reg uint, imm int32) {
	rex := byte(0x48)
	regCode := uint8(reg - gapstone.X86_REG_RAX)
	if regCode >= 8 {
		rex |= 1 // REX.B
	}
	a.buf = append(a.buf, rex)
	a.buf = append(a.buf, 0x81)

	modRM := byte(0b11_000_000) | (regCode & 7)
	a.buf = append(a.buf, modRM)

	immBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(immBytes, uint32(imm))
	a.buf = append(a.buf, immBytes...)
}

// jmp_rel32 adds `jmp rel32`
func (a *asmHelper) jmp_rel32(targetAddr, sourceAddr uint64) {
	a.buf = append(a.buf, 0xE9)
	rel32 := int32(targetAddr - (sourceAddr + 5))
	relBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(relBytes, uint32(rel32))
	a.buf = append(a.buf, relBytes...)
}

func (a *asmHelper) bytes() []byte {
	return a.buf
}
