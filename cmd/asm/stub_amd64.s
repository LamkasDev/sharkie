//go:build amd64

#include "reg_amd64.s"
#include "thread_context_amd64.s"
#include "funcdata.h"

// InitStubAddr is called from Go's init function.
// It gets the address of our assembly handler and stores it in a Go variable.
TEXT ·InitStubAddr(SB), NOSPLIT, $0
    LEAQ ·stubAsm(SB), AX
    MOVQ AX, ·StubAddr(SB)
    RET

// stubAsm is a generic stub that can call any Go function and is entered from guest code.
// The target Go function's address is expected to be in R11.
// Arguments for the Go function are passed in guest registers (DI, SI, DX, CX, R8, R9) and on the guest stack.
// The return address for guest code is expected to be on the top of the guest stack.
TEXT ·stubAsm(SB), NOSPLIT|NOFRAME, $0-0
    NO_LOCAL_POINTERS
    BYTE $0x48; BYTE $0x81; BYTE $0xEC; BYTE $0x80; BYTE $0x01; BYTE $0x00; BYTE $0x00 // SUBQ $384, SP

    // Save all general-purpose registers.
    SAVE_REGS

    // Save Thread Context Pointer into R12.
    CALL ·GetTLSContext(SB)
    MOVQ AX, R12

    // Save playstation stack.
    MOVQ SP, CTX_PS_SP(R12)

    // Restore Go stack into scratch registers.
    MOVQ CTX_SAVED_G(R12), R14
    MOVQ CTX_GO_SP(R12), BX
    MOVQ CTX_GO_BP(R12), BP

    // Pass context pointer.
    MOVQ SP, CTX_STUB_CTX(R12)

    // For real switch to Go stack.
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Call the Go trampoline function.
    CALL ·stubGo(SB)

    // Save Thread Context Pointer into R12.
    CALL ·GetTLSContext(SB)
    MOVQ AX, R12

    // Save Go stack.
    MOVQ SP, CTX_GO_SP(R12)
    MOVQ BP, CTX_GO_BP(R12)

    // Switch to playstation stack.
    MOVQ CTX_PS_SP(R12), BX
    BYTE $0x48; BYTE $0x89; BYTE $0xDC  // MOVQ BX, SP

    // Restore all general-purpose registers.
    RESTORE_REGS

    // Return to the game code.
    BYTE $0x48; BYTE $0x81; BYTE $0xC4; BYTE $0x80; BYTE $0x01; BYTE $0x00; BYTE $0x00 // ADDQ $384, SP
    RET
