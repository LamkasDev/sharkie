#include "textflag.h"

#define REG_AX 0
#define REG_BX 8
#define REG_CX 16
#define REG_DX 24
#define REG_SI 32
#define REG_DI 40
#define REG_R8 48
#define REG_R9 56
#define REG_R10 64
#define REG_R11 72
#define REG_R12 80
#define REG_R13 88
#define REG_R14 96
#define REG_R15 104
#define REG_XMM 112
#define REG_BP 368          // XMM area is 16 * 16 = 256 bytes. 112 + 256 = 368
#define CONTEXT_SIZE 384    // Aligned to 16 bytes

#define CTX_TID 0
#define CTX_SYSTEM_SP 8
#define CTX_PS_SP 16
#define CTX_GO_SP 24
#define CTX_LAST_GO_SP 32
#define CTX_GO_BP 40
#define CTX_SAVED_G 48
#define CTX_RET_ANCHOR 56
#define CTX_STUB_CTX 64
#define CTX_EXC_INFO 72

// SAVE_REGS saves all general-purpose registers.
#define SAVE_REGS \
    MOVQ AX, REG_AX(SP) \
    MOVQ BX, REG_BX(SP) \
    MOVQ CX, REG_CX(SP) \
    MOVQ DX, REG_DX(SP) \
    MOVQ SI, REG_SI(SP) \
    MOVQ DI, REG_DI(SP) \
    MOVQ R8, REG_R8(SP) \
    MOVQ R9, REG_R9(SP) \
    MOVQ R10, REG_R10(SP) \
    MOVQ R11, REG_R11(SP) \
    MOVQ R12, REG_R12(SP) \
    MOVQ R13, REG_R13(SP) \
    MOVQ R14, REG_R14(SP) \
    MOVQ R15, REG_R15(SP) \
    MOVQ BP, REG_BP(SP)

// RESTORE_REGS restores all general-purpose registers.
#define RESTORE_REGS \
    MOVQ REG_AX(SP), AX \
    MOVQ REG_BX(SP), BX \
    MOVQ REG_CX(SP), CX \
    MOVQ REG_DX(SP), DX \
    MOVQ REG_SI(SP), SI \
    MOVQ REG_DI(SP), DI \
    MOVQ REG_R8(SP), R8 \
    MOVQ REG_R9(SP), R9 \
    MOVQ REG_R10(SP), R10 \
    MOVQ REG_R11(SP), R11 \
    MOVQ REG_R12(SP), R12 \
    MOVQ REG_R13(SP), R13 \
    MOVQ REG_R14(SP), R14 \
    MOVQ REG_R15(SP), R15 \
    MOVQ REG_BP(SP), BP
