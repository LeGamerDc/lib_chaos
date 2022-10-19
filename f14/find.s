#include "textflag.h"

// func find(a, b [16]byte) (c uint32)
TEXT ·find(SB),NOSPLIT,$0-36
MOVOU a+0(FP), X0
MOVOU b+16(FP), X1
PCMPEQB X0, X1
PMOVMSKB X1, AX
MOVL AX, c+32(FP)
RET