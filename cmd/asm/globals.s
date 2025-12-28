//go:build windows && amd64

#include "textflag.h"

TEXT ·SetGuestSP(SB), NOSPLIT, $0-8
    MOVQ sp+0(FP), AX
    MOVQ AX, ·PlaystationStackSP(SB)
    RET

TEXT ·GetGuestSP(SB), NOSPLIT, $0-8
    MOVQ ·PlaystationStackSP(SB), AX
    MOVQ AX, ret+0(FP)
    RET
