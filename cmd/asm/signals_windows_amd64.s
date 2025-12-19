//go:build windows && amd64

#include "textflag.h"
#include "funcdata.h"

GLOBL ·ExceptionHandlerAddr(SB), NOPTR, $8
GLOBL ·GlobalExceptionInfo(SB), NOPTR, $8
GLOBL ·WindowsStackSP(SB), NOPTR, $8
GLOBL ·GoStackSP(SB), NOPTR, $8
GLOBL ·GoStackBP(SB), NOPTR, $8
GLOBL ·SavedG(SB), NOPTR, $8
GLOBL ·ReturnAddressAnchor(SB), NOPTR, $8

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
TEXT ·exceptionHandlerAsm(SB), NOSPLIT, $8-0
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

    // Save Windows stack.
    MOVQ SP, ·WindowsStackSP(SB)

    // Restore Go stack pointer into scratch register.
    MOVQ ·SavedG(SB), R14
    MOVQ ·GoStackSP(SB), BX

    // Create fake call frame.
    SUBQ $16, BX
    MOVQ ·ReturnAddressAnchor(SB), SI
    MOVQ SI, 8(BX)

    // Pass the exception info pointer.
    MOVQ CX, ·GlobalExceptionInfo(SB)
    MOVQ ·GoStackBP(SB), BP

    // For real restore Go stack pointer.
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Call the Go exception handler.
    CALL ·exceptionHandlerGo(SB)

    // Save Go stack pointer.
    MOVQ SP, BX
    ADDQ $16, BX
    MOVQ BX, ·GoStackSP(SB)
    MOVQ BP, ·GoStackBP(SB)

    // Switch to Windows stack.
    MOVQ ·WindowsStackSP(SB), BX
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
