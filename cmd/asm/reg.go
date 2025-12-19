package asm

type RegContext struct {
	AX, BX, CX, DX, SI, DI, R8, R9, R10, R11, R12, R13, R14, R15 uintptr
	XMM                                                          [16][2]uintptr // 128-bit XMMs (16 * 16 bytes)
	BP                                                           uintptr
}
