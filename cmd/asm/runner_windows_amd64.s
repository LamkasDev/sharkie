//go:build windows && amd64

#include "reg_amd64.s"
#include "funcdata.h"

// This function switches to the game's stack and jumps to its entry point.
// It does not return.
// func Run(entry,   stackPtr, argsPtr, arg2 uintptr)
//          +0(FP)   +8(FP)    +16(FP)  +24(FP)
TEXT ·Run(SB), NOSPLIT, $0-32
    NO_LOCAL_POINTERS

    // Save Thread Context Pointer into R12.
    GET_TLS_CONTEXT(R12)

    // We fake call site in case we run into an exception handler.
    BYTE $0xEB; BYTE $0x05  // JMP +5 bytes
    CALL ·exceptionHandlerAsm(SB)

    // Save fake caller address in case we run into an exception handler.
    LEAQ ·Run(SB), R15
    ADDQ $7, R15
    MOVQ R15, CTX_RET_ANCHOR(R12)

    // Save the current Go stack so we can restore it later.
    MOVQ SP, CTX_GO_SP(R12)
    MOVQ BP, CTX_GO_BP(R12)
    MOVQ R14, CTX_SAVED_G(R12)

    MOVQ entry+0(FP), AX        // entry = AX
    MOVQ stackPtr+8(FP), BX     // stackPtr = BX
    MOVQ argsPtr+16(FP), DI      // argsPtr = DI

    // Switch to the playstation stack.
    ANDQ $-16, BX
    SUBQ $8, BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Clear our registers.
    XORQ CX, CX
    XORQ DX, DX
    XORQ SI, SI
    XORQ R8, R8
    XORQ R9, R9
    XORQ R10, R10
    XORQ R11, R11
    XORQ R12, R12
    XORQ BP, BP

    // Far jump to the entry point to set CS.
    PUSHQ $0x33 // CS segment selector
    PUSHQ AX    // Entry point address
    RETFQ

// This function switches to the game's stack, calls a function and returns.
// func Call(entry,   stackPtr, arg1,    arg2 uintptr)
//          +0(FP)   +8(FP)    +16(FP)  +24(FP)
TEXT ·Call(SB), NOSPLIT, $48-32
    NO_LOCAL_POINTERS

    // Save Thread Context Pointer into DX.
    GET_TLS_CONTEXT(DX)

    // We fake call site in case we run into an exception handler.
    BYTE $0xEB; BYTE $0x05  // JMP +5 bytes
    CALL ·exceptionHandlerAsm(SB)

    // Save fake caller address in case we run into an exception handler.
    LEAQ ·Call(SB), R15
    ADDQ $7, R15
    MOVQ R15, CTX_RET_ANCHOR(DX)

    // Save callee-saved registers.
    MOVQ BP, 0(SP)
    MOVQ BX, 8(SP)
    MOVQ R12, 16(SP)
    MOVQ R13, 24(SP)
    MOVQ R14, 32(SP)
    MOVQ R15, 40(SP)

    // Save the current Go stack so we can restore it later.
    MOVQ SP, CTX_GO_SP(DX)
    MOVQ BP, CTX_GO_BP(DX)
    MOVQ R14, CTX_SAVED_G(DX)

    MOVQ entry+0(FP), AX     // entry = AX
    MOVQ stackPtr+8(FP), BX  // stackPtr = BX

    // Prepare a return address for guest.
    BYTE $0xE8; BYTE $0x05; BYTE $0x00; BYTE $0x00; BYTE $0x00
    JMP CallRestoreRegisters
    BYTE $0x90; BYTE $0x90; BYTE $0x90

    // Save return address.
    BYTE $0x5E
    MOVQ SI, CTX_CALL_RET(DX)

    // Switch to the playstation stack.
    SUBQ $32, BX
    SUBQ $24, BX
    MOVQ $0, 16(BX)
    MOVQ $0, 8(BX)
    MOVQ $0, 0(BX)
    ANDQ $-16, BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Call the entry function with empty registers.
    BYTE $0x48; BYTE $0x83; BYTE $0xEC; BYTE $0x08  // SUBQ $8, SP
    MOVQ SI, 0(SP)
    XORQ CX, CX
    XORQ DX, DX
    XORQ DI, DI
    XORQ SI, SI
    XORQ R8, R8
    XORQ R9, R9
    XORQ R10, R10
    XORQ R11, R11
    JMP AX

CallRestoreRegisters:
    // Save Thread Context Pointer into R12.
    GET_TLS_CONTEXT(R12)

    // Switch to the Go stack.
    MOVQ CTX_GO_SP(R12), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Restore Go stack.
    MOVQ CTX_GO_BP(R12), BP
    MOVQ CTX_SAVED_G(R12), R14

    // Restore callee-saved registers.
    MOVQ 40(SP), R15
    MOVQ 32(SP), R14
    MOVQ 24(SP), R13
    MOVQ 16(SP), R12
    MOVQ 8(SP), BX
    MOVQ 0(SP), BP

    RET
