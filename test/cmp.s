#include "textflag.h"

TEXT ·add(SB),NOSPLIT,$0-24
MOVQ $0, c+24(SP)
MOVBLZX a+8(SP), AX
MOVBLZX b+16(SP), CX
ADDL CX, AX
MOVB AL, c+24(SP)
RET
