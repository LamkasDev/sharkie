//go:build windows && amd64

#include "reg_amd64.s"
#include "funcdata.h"

GLOBL ·GlobalStubContext(SB), NOPTR, $8
GLOBL ·PlaystationStackSP(SB), NOPTR, $8
GLOBL ·StubAddr(SB), NOPTR, $8
GLOBL ·GoStackSP(SB), NOPTR, $8
GLOBL ·GoStackBP(SB), NOPTR, $8
GLOBL ·SavedG(SB), NOPTR, $8
GLOBL ·ReturnAddressAnchor(SB), NOPTR, $8
GLOBL ·CallReturnAddress(SB), NOPTR, $8

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
TEXT ·stubAsm(SB), NOSPLIT, $0-0
    NO_LOCAL_POINTERS

    // We fake call site in case we run into an exception handler.
    BYTE $0xEB; BYTE $0x05  // JMP +5 bytes
    CALL ·exceptionHandlerAsm(SB)

    // Save all general-purpose registers.
    SAVE_REGS

    // Save fake caller address in case we run into an exception handler.
    LEAQ ·stubAsm(SB), SI
    ADDQ $7, SI
    MOVQ SI, ·ReturnAddressAnchor(SB)

    // Save playstation stack.
    MOVQ SP, R13
    MOVQ R13, ·PlaystationStackSP(SB)

    // Pass context pointer.
    MOVQ R13, ·GlobalStubContext(SB)

    // Restore Go stack into scratch registers.
    MOVQ ·SavedG(SB), R14
    MOVQ ·GoStackSP(SB), BX
    MOVQ ·GoStackBP(SB), BP

    // Construct fake call frame.
    SUBQ $16, BX
    MOVQ ·CallReturnAddress(SB), AX
    MOVQ AX, 8(BX)
    MOVQ BP, 0(BX)
    LEAQ 0(BX), BP

    // For real switch to Go stack.
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Call the Go trampoline function.
    CALL ·stubGo(SB)

    // Clean up fake call frame.
    MOVQ 0(SP), BP
    MOVQ SP, BX
    ADDQ $16, BX

    // Save Go stack.
    MOVQ BX, ·GoStackSP(SB)
    MOVQ BP, ·GoStackBP(SB)

    // Switch to playstation stack.
    MOVQ ·PlaystationStackSP(SB), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Restore all general-purpose registers.
    RESTORE_REGS

    // Return to the game code.
    RET
