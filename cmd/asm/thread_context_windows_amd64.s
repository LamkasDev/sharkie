//go:build windows && amd64

#include "textflag.h"
#include "funcdata.h"

// GetTLSContext returns the TLS context pointer for the current thread.
TEXT ·GetTLSContext(SB), NOSPLIT, $0-0
    NO_LOCAL_POINTERS

    MOVQ 0x30(GS), AX
    ADDQ ·GoTlsOffset(SB), AX
    MOVQ (AX), AX
    RET

// GetCurrentThreadContext returns ThreadContext for the current thread.
TEXT ·GetCurrentThreadContext(SB), NOSPLIT, $0-8
    NO_LOCAL_POINTERS

    MOVQ 0x30(GS), AX
    ADDQ ·GoTlsOffset(SB), AX
    MOVQ (AX), AX
    MOVQ AX, ret+0(FP)
    RET
