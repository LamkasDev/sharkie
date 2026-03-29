//go:build linux && amd64

#include "reg_amd64.s"
#include "thread_context_amd64.s"
#include "funcdata.h"

// InitSignalsAddr is called from Go's init function.
// It gets the address of our assembly handler and stores it in a Go variable.
TEXT ·InitSignalsAddr(SB), NOSPLIT, $0
    LEAQ ·exceptionHandlerAsm(SB), AX
    MOVQ AX, ·ExceptionHandlerAddr(SB)
    RET

// This function is the assembly exception handler.
// It switches to the Go stack, calls the Go exception handler with exception info and returns.
// void signalHandlerAsm(int sig, siginfo_t *info, void *ucontext);
//                       (DI)     (SI)             (DX)
TEXT ·exceptionHandlerAsm(SB), NOSPLIT, $80-0
    NO_LOCAL_POINTERS

    // Save Linux non-volatile registers.
    MOVQ BP, 0(SP)
    MOVQ BX, 8(SP)
    MOVQ R12, 16(SP)
    MOVQ R13, 24(SP)
    MOVQ R14, 32(SP)
    MOVQ R15, 40(SP)

    // Save signal info.
    MOVQ DI, 48(SP) // sig
    MOVQ SI, 56(SP) // info
    MOVQ DX, 64(SP) // ucontext

    // Save Thread Context Pointer into DX.
    CALL ·GetTLSContext(SB)
    MOVQ AX, DX

    // Save Linux stack.
    MOVQ SP, CTX_SYSTEM_SP(DX)

    // Restore Go stack pointer into scratch register.
    MOVQ CTX_SAVED_G(DX), R14
    MOVQ CTX_GO_SP(DX), BX
    MOVQ CTX_GO_BP(DX), BP

    // Create an exception info struct and pass the pointer.
    MOVQ 56(SP), AX // info
    MOVQ AX, CTX_EXC_INFO_BUF(DX)
    MOVQ 64(SP), AX // ucontext
    MOVQ AX, CTX_EXC_INFO_BUF+8(DX)
    LEAQ CTX_EXC_INFO_BUF(DX), AX
    MOVQ AX, CTX_EXC_INFO(DX)

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

    // Switch to Linux stack.
    MOVQ CTX_SYSTEM_SP(DX), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Restore Linux non-volatile registers.
    MOVQ 40(SP), R15
    MOVQ 32(SP), R14
    MOVQ 24(SP), R13
    MOVQ 16(SP), R12
    MOVQ 8(SP), BX
    MOVQ 0(SP), BP

    // Return to Linux.
    RET
