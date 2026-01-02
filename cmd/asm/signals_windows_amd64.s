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
TEXT ·exceptionHandlerAsm(SB), NOSPLIT, $0-0
    NO_LOCAL_POINTERS

    // Save Windows non-volatile registers.
    PUSHQ BP
    PUSHQ BX
    PUSHQ DI
    PUSHQ SI
    PUSHQ R12
    PUSHQ R13
    PUSHQ R14
    PUSHQ R15

    // Save Thread Context Pointer into R13.
    GET_TLS_CONTEXT(R13)

    // Save Windows stack.
    MOVQ SP, CTX_WIN_SP(R13)

    // Restore Go stack pointer into scratch register.
    MOVQ CTX_SAVED_G(R13), R14
    MOVQ CTX_GO_SP(R13), BX
    MOVQ CTX_GO_BP(R13), BP

    // Pass the exception info pointer.
    MOVQ CX, CTX_EXC_INFO(R13)

    // For real restore Go stack pointer.
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Setup fake call frame.
    PUSHQ $0
    MOVQ CTX_RET_ANCHOR(R15), AX
    PUSHQ AX

    // Call the Go exception handler.
    CALL ·exceptionHandlerGo(SB)

    // Clean up fake call frame.
    POPQ AX
    POPQ AX

    // Save Thread Context Pointer into R13.
    GET_TLS_CONTEXT(R13)

    // Save Go stack pointer.
    MOVQ SP, CTX_GO_SP(R13)
    MOVQ BP, CTX_GO_BP(R13)

    // Switch to Windows stack.
    MOVQ CTX_WIN_SP(R13), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Restore Windows non-volatile registers.
    POPQ R15
    POPQ R14
    POPQ R13
    POPQ R12
    POPQ SI
    POPQ DI
    POPQ BX
    POPQ BP

    // Return to Windows.
    RET
