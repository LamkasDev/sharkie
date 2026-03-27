//go:build amd64

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
    ADDQ $7, CX // 2 (JMP) + 5 (CALL)
    PUSHQ CX

    // Save Thread Context Pointer into DX.
    CALL ·GetTLSContext(SB)
    MOVQ AX, DX

    // Save the current Go stack so we can restore it later.
    MOVQ SP, CTX_GO_SP(DX)
    MOVQ BP, CTX_GO_BP(DX)
    MOVQ R14, CTX_SAVED_G(DX)

    // Load arguments.
    MOVQ entry+0(FP), AX    // entry = AX
    MOVQ stackPtr+8(FP), BX // stackPtr = BX
    MOVQ argsPtr+16(FP), DI // argsPtr = DI

    // Switch to the playstation stack.
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
// We can't expand the caller's stack afterwards or there will be trouble (split stack overflow).
// func Call(entry,   stackPtr, arg1,    arg2 uintptr)
//          +0(FP)   +8(FP)    +16(FP)  +24(FP)
TEXT ·Call(SB), NOSPLIT|NOFRAME, $0-32
    NO_LOCAL_POINTERS

    // We fake call site for stub.
    BYTE $0xEB; BYTE $0x05  // JMP +5 bytes
    CALL ·stubAsm(SB)

    // Setup fake call frame.
    LEAQ ·Call(SB), CX
    ADDQ $7, CX // 2 (JMP) + 5 (CALL)
    PUSHQ CX

    // Save Thread Context Pointer into DX.
    CALL ·GetTLSContext(SB)
    MOVQ AX, DX

    // Save callee-saved registers.
    MOVQ BP, CTX_CALL_SAVED_BP(DX)
    MOVQ BX, CTX_CALL_SAVED_BX(DX)
    MOVQ R12, CTX_CALL_SAVED_R12(DX)
    MOVQ R13, CTX_CALL_SAVED_R13(DX)
    MOVQ R14, CTX_CALL_SAVED_R14(DX)
    MOVQ R15, CTX_CALL_SAVED_R15(DX)
    MOVQ SP, CTX_CALL_SAVED_SP(DX)

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

    // Save Thread Context Pointer into DX.
    CALL ·GetTLSContext(SB)
    MOVQ AX, DX

    // Restore Go stack.
    MOVQ CTX_CALL_SAVED_SP(DX), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Clean up fake call frame.
    POPQ CX

    // Restore callee-saved registers.
    MOVQ CTX_CALL_SAVED_R15(DX), R15
    MOVQ CTX_CALL_SAVED_R14(DX), R14
    MOVQ CTX_CALL_SAVED_R13(DX), R13
    MOVQ CTX_CALL_SAVED_R12(DX), R12
    MOVQ CTX_CALL_SAVED_BX(DX), BX
    MOVQ CTX_CALL_SAVED_BP(DX), BP

    RET
