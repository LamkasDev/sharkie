//go:build linux && amd64

#include "textflag.h"
#include "funcdata.h"

// GetTLSContext returns the TLS context pointer for the current thread.
TEXT ·GetTLSContext(SB), NOSPLIT, $0-0
    NO_LOCAL_POINTERS

    MOVQ ·GoTlsOffset(SB), CX
    BYTE $0x64; BYTE $0x48; BYTE $0x8B; BYTE $0x01  // MOVQ FS:(CX), AX
    RET

// GetCurrentThreadContext returns ThreadContext for the current thread.
TEXT ·GetCurrentThreadContext(SB), NOSPLIT, $0-8
    NO_LOCAL_POINTERS

    MOVQ ·GoTlsOffset(SB), CX
    BYTE $0x64; BYTE $0x48; BYTE $0x8B; BYTE $0x01  // MOVQ FS:(CX), AX
    MOVQ AX, ret+0(FP)
    RET
