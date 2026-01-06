//go:build windows && amd64

#include "reg_amd64.s"
#include "funcdata.h"

// This function switches to the game's stack and jumps to its entry point.
// It does not return.
// func Run(entry,   stackPtr, argsPtr, arg2 uintptr)
//          +0(FP)   +8(FP)    +16(FP)  +24(FP)
TEXT ·Run(SB), NOSPLIT, $0-32
    NO_LOCAL_POINTERS

    // We fake call site for stub.
    BYTE $0xEB; BYTE $0x05  // JMP +5 bytes
    CALL ·stubAsm(SB)

    // Setup fake call frame.
    LEAQ ·Run(SB), CX
    ADDQ $7, CX
    PUSHQ CX

    // Save Thread Context Pointer into DX.
    GET_TLS_CONTEXT(DX)

    // Save the current Go stack so we can restore it later.
    MOVQ SP, CTX_GO_SP(DX)
    MOVQ BP, CTX_GO_BP(DX)
    MOVQ R14, CTX_SAVED_G(DX)

    // Load arguments.
    MOVQ entry+0(FP), AX    // entry = AX
    MOVQ stackPtr+8(FP), BX // stackPtr = BX
    MOVQ argsPtr+16(FP), DI // argsPtr = DI

    // Switch to the playstation stack.
    ANDQ $-16, BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Clear our registers.
    XORQ CX, CX
    XORQ DX, DX
    XORQ SI, SI
    XORQ R8, R8
    XORQ R9, R9
    XORQ R10, R10
    XORQ R11, R11

    // Call entry function.
    CALL AX

    // Clean up fake call frame (not really, just for balance).
    POPQ CX

    RET

// This function switches to the game's stack, calls a function and returns.
// func Call(entry,   stackPtr, arg1,    arg2 uintptr)
//          +0(FP)   +8(FP)    +16(FP)  +24(FP)
TEXT ·Call(SB), NOSPLIT, $48-32
    NO_LOCAL_POINTERS

    // We fake call site for stub.
    BYTE $0xEB; BYTE $0x05  // JMP +5 bytes
    CALL ·stubAsm(SB)

    // Save callee-saved registers.
    MOVQ BP, 0(SP)
    MOVQ BX, 8(SP)
    MOVQ R12, 16(SP)
    MOVQ R13, 24(SP)
    MOVQ R14, 32(SP)
    MOVQ R15, 40(SP)

    // Setup fake call frame.
    LEAQ ·Call(SB), CX
    ADDQ $7, CX
    PUSHQ CX

    // Save Thread Context Pointer into DX.
    GET_TLS_CONTEXT(DX)

    // Save the current Go stack so we can restore it later.
    MOVQ SP, CTX_GO_SP(DX)
    MOVQ BP, CTX_GO_BP(DX)
    MOVQ R14, CTX_SAVED_G(DX)

    // Load arguments.
    MOVQ entry+0(FP), AX    // entry = AX
    MOVQ stackPtr+8(FP), BX // stackPtr = BX
    MOVQ arg1+16(FP), DI    // arg1 = DI
    MOVQ arg2+24(FP), SI    // arg2 = SI

    // Switch to the playstation stack.
    ANDQ $-16, BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Clear registers.
    XORQ CX, CX
    XORQ DX, DX
    XORQ R8, R8
    XORQ R9, R9
    XORQ R10, R10
    XORQ R11, R11

    // Call function.
    CALL AX

    // Save Thread Context Pointer into R12.
    GET_TLS_CONTEXT(DX)

    // Switch to the Go stack.
    MOVQ CTX_GO_SP(DX), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Clean up fake call frame.
    POPQ CX

    // Restore Go stack.
    MOVQ CTX_GO_BP(DX), BP
    MOVQ CTX_SAVED_G(DX), R14

    // Restore callee-saved registers.
    MOVQ 40(SP), R15
    MOVQ 32(SP), R14
    MOVQ 24(SP), R13
    MOVQ 16(SP), R12
    MOVQ 8(SP), BX
    MOVQ 0(SP), BP

    RET
