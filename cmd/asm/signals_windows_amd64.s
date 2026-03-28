//go:build windows && amd64

#include "reg_amd64.s"
#include "funcdata.h"

// InitSignalsAddr is called from Go's init function.
// It gets the address of our assembly handler and stores it in a Go variable.
TEXT ·InitSignalsAddr(SB), NOSPLIT, $0
    LEAQ ·exceptionHandlerAsm(SB), AX
    MOVQ AX, ·ExceptionHandlerAddr(SB)
    RET

// This function is the assembly exception handler.
// It switches to the Go stack, calls the Go exception handler with exception info and returns.
// LONG NTAPI VectoredHandler(PEXCEPTION_POINTERS ExceptionInfo);
//                                                +0(RCX)
TEXT ·exceptionHandlerAsm(SB), NOSPLIT, $64-0
    NO_LOCAL_POINTERS

    // Save Windows non-volatile registers.
    MOVQ BP, 0(SP)
    MOVQ BX, 8(SP)
    MOVQ DI, 16(SP)
    MOVQ SI, 24(SP)
    MOVQ R12, 32(SP)
    MOVQ R13, 40(SP)
    MOVQ R14, 48(SP)
    MOVQ R15, 56(SP)

    // Save exception info pointer.
    MOVQ CX, R12

    // Save Thread Context Pointer into DX.
    CALL ·GetTLSContext(SB)
    MOVQ AX, DX

    // Save Windows stack.
    MOVQ SP, CTX_SYSTEM_SP(DX)

    // Restore Go stack pointer into scratch register.
    MOVQ CTX_SAVED_G(DX), R14
    MOVQ CTX_GO_SP(DX), BX
    MOVQ CTX_GO_BP(DX), BP

    // Pass the exception info pointer.
    MOVQ R12, CTX_EXC_INFO(DX)

    // For real restore Go stack pointer.
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Call the Go exception handler.
    CALL ·exceptionHandlerGo(SB)

    // Save Thread Context Pointer into DX.
    CALL ·GetTLSContext(SB)
    MOVQ AX, DX

    // Save Go stack pointer.
    MOVQ SP, CTX_GO_SP(DX)
    MOVQ BP, CTX_GO_BP(DX)

    // Switch to Windows stack.
    MOVQ CTX_SYSTEM_SP(DX), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Restore Windows non-volatile registers.
    MOVQ 56(SP), R15
    MOVQ 48(SP), R14
    MOVQ 40(SP), R13
    MOVQ 32(SP), R12
    MOVQ 24(SP), SI
    MOVQ 16(SP), DI
    MOVQ 8(SP), BX
    MOVQ 0(SP), BP

    // Return to Windows.
    RET
