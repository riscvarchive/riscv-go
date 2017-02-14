#include "textflag.h"

// int64 cputicks(void)
TEXT ·cputicks(SB),NOSPLIT,$0-8
	WORD	$0xc0102573	// rdtime a0
	MOV	A0, ret+0(FP)
	RET
