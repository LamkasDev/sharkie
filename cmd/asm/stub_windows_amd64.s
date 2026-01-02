//go:build windows && amd64

#include "reg_amd64.s"
#include "funcdata.h"

// InitStubAddr is called from Go's init function.
// It gets the address of our assembly handler and stores it in a Go variable.
TEXT ·InitStubAddr(SB), NOSPLIT, $0
    LEAQ ·stubAsm(SB), AX
    MOVQ AX, ·StubAddr(SB)
    RET

// stubAsm is a generic stub that can call any Go function.
// The target function's address is passed in R11.
// Arguments are passed in registers (RCX, RDX, R8, R9) and on the stack.
// The return address is expected to be on the top of the stack.
TEXT ·stubAsm(SB), NOSPLIT|NOFRAME, $0-0
    NO_LOCAL_POINTERS

    // We fake call site in case we run into an exception handler.
    BYTE $0xEB; BYTE $0x05  // JMP +5 bytes
    CALL ·exceptionHandlerAsm(SB)

    // Save all general-purpose registers.
    SAVE_REGS

    // Save Thread Context Pointer into R15.
    GET_TLS_CONTEXT(R15)

    // Save fake caller address in case we run into an exception handler.
    LEAQ ·stubAsm(SB), SI
    ADDQ $7, SI
    MOVQ SI, CTX_RET_ANCHOR(R15)

    // Save playstation stack.
    MOVQ SP, CTX_PS_SP(R15)

    // Restore Go stack into scratch registers.
    MOVQ CTX_SAVED_G(R15), R14
    MOVQ CTX_GO_SP(R15), BX
    MOVQ CTX_GO_BP(R15), BP

    // Pass context pointer.
    MOVQ SP, CTX_STUB_CTX(R15)

    // For real switch to Go stack.
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Setup fake call frame.
    PUSHQ $0
    MOVQ CTX_RET_ANCHOR(R15), AX
    PUSHQ AX

    // Call the Go trampoline function.
    CALL ·stubGo(SB)

    // Clean up fake call frame.
    POPQ AX
    POPQ AX

    // Save Thread Context Pointer into R15.
    GET_TLS_CONTEXT(R15)

    // Save Go stack.
    MOVQ SP, CTX_GO_SP(R15)
    MOVQ BP, CTX_GO_BP(R15)

    // Switch to playstation stack.
    MOVQ CTX_PS_SP(R15), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Restore all general-purpose registers.
    RESTORE_REGS

    // Return to the game code.
    RET
