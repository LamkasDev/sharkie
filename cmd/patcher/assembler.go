package patcher

import (
	"encoding/binary"
	"math"

	"github.com/bpfsnoop/gapstone"
)

type regEnc struct {
	code uint8
	rex  byte
}

func encodeGpr64(reg uint) (regEnc, bool) {
	switch reg {
	case gapstone.X86_REG_RAX:
		return regEnc{0, 0}, true
	case gapstone.X86_REG_RCX:
		return regEnc{1, 0}, true
	case gapstone.X86_REG_RDX:
		return regEnc{2, 0}, true
	case gapstone.X86_REG_RBX:
		return regEnc{3, 0}, true
	case gapstone.X86_REG_RSP:
		return regEnc{4, 0}, true
	case gapstone.X86_REG_RBP:
		return regEnc{5, 0}, true
	case gapstone.X86_REG_RSI:
		return regEnc{6, 0}, true
	case gapstone.X86_REG_RDI:
		return regEnc{7, 0}, true
	case gapstone.X86_REG_R8:
		return regEnc{0, 0x01}, true
	case gapstone.X86_REG_R9:
		return regEnc{1, 0x01}, true
	case gapstone.X86_REG_R10:
		return regEnc{2, 0x01}, true
	case gapstone.X86_REG_R11:
		return regEnc{3, 0x01}, true
	case gapstone.X86_REG_R12:
		return regEnc{4, 0x01}, true
	case gapstone.X86_REG_R13:
		return regEnc{5, 0x01}, true
	case gapstone.X86_REG_R14:
		return regEnc{6, 0x01}, true
	case gapstone.X86_REG_R15:
		return regEnc{7, 0x01}, true
	default:
		return regEnc{}, false
	}
}

// asmHelper helps build small assembly snippets for trampolines.
type asmHelper struct {
	buf []byte
}

func newAsmHelper() *asmHelper {
	return &asmHelper{buf: []byte{}}
}

// mov_r64_from_gs_mem adds `mov dst, gs:[disp32]`
func (a *asmHelper) mov_r64_from_gs_mem(dst uint, disp int32) {
	enc, ok := encodeGpr64(dst)
	if !ok {
		panic("unsupported register")
	}

	a.buf = append(a.buf, 0x65) // GS

	rex := byte(0x48) // REX.W
	if enc.rex != 0 {
		rex |= 0x04 // REX.R
	}
	a.buf = append(a.buf, rex)
	a.buf = append(a.buf, 0x8B)

	modRM := 0x00<<6 | enc.code<<3 | 0x04
	a.buf = append(a.buf, modRM)
	a.buf = append(a.buf, 0x25) // no index, no base, disp32

	var dispBytes [4]byte
	binary.LittleEndian.PutUint32(dispBytes[:], uint32(disp))
	a.buf = append(a.buf, dispBytes[:]...)
}

// mov_r64_from_gs_mem adds `mov dst, fs:[disp32]`
func (a *asmHelper) mov_r64_from_fs_mem(dst uint, disp int32) {
	enc, ok := encodeGpr64(dst)
	if !ok {
		panic("unsupported register")
	}

	a.buf = append(a.buf, 0x64) // FS

	rex := byte(0x48) // REX.W
	if enc.rex != 0 {
		rex |= 0x04 // REX.R
	}
	a.buf = append(a.buf, rex)
	a.buf = append(a.buf, 0x8B)

	modRM := 0x00<<6 | enc.code<<3 | 0x04
	a.buf = append(a.buf, modRM)
	a.buf = append(a.buf, 0x25) // no index, no base, disp32

	var dispBytes [4]byte
	binary.LittleEndian.PutUint32(dispBytes[:], uint32(disp))
	a.buf = append(a.buf, dispBytes[:]...)
}

// mov_r64_from_mem adds `mov dst, [base + disp32]`
func (a *asmHelper) mov_r64_from_mem(dst, base uint, disp int32) {
	dstEnc, ok := encodeGpr64(dst)
	if !ok {
		panic("unsupported dst register")
	}
	baseEnc, ok := encodeGpr64(base)
	if !ok {
		panic("unsupported base register")
	}

	// REX.W
	rex := byte(0x48)
	if dstEnc.rex != 0 {
		rex |= 0x04 // REX.R
	}
	if baseEnc.rex != 0 {
		rex |= 0x01 // REX.B
	}
	a.buf = append(a.buf, rex)

	// MOV r64, r/m64
	a.buf = append(a.buf, 0x8B)

	// MODRM: MOD=10 (disp32)
	modRM := 0x80 | (dstEnc.code << 3) | baseEnc.code
	a.buf = append(a.buf, modRM)

	// SIB if base == RSP/R12
	if baseEnc.code == 4 {
		a.buf = append(a.buf, 0x24)
	}

	var dispBytes [4]byte
	binary.LittleEndian.PutUint32(dispBytes[:], uint32(disp))
	a.buf = append(a.buf, dispBytes[:]...)
}

// add_r64_imm32 adds `add reg, imm32`
func (a *asmHelper) add_r64_imm32(reg uint, imm int32) {
	enc, ok := encodeGpr64(reg)
	if !ok {
		panic("unsupported register")
	}

	rex := byte(0x48)
	if enc.rex != 0 {
		rex |= 0x01
	}
	a.buf = append(a.buf, rex)

	a.buf = append(a.buf, 0x81)
	a.buf = append(a.buf, 0xC0|enc.code)

	var immBytes [4]byte
	binary.LittleEndian.PutUint32(immBytes[:], uint32(imm))
	a.buf = append(a.buf, immBytes[:]...)
}

// jmp_rel32 adds `jmp rel32`
func (a *asmHelper) jmp_rel32(targetAddr, sourceAddr uint64) {
	a.buf = append(a.buf, 0xE9)
	rel := int64(targetAddr) - int64(sourceAddr+5)
	if rel < math.MinInt32 || rel > math.MaxInt32 {
		panic("jmp_rel32 out of range")
	}

	rel32 := int32(rel)
	var relBytes [4]byte
	binary.LittleEndian.PutUint32(relBytes[:], uint32(rel32))
	a.buf = append(a.buf, relBytes[:]...)
}

func (a *asmHelper) bytes() []byte {
	return a.buf
}
