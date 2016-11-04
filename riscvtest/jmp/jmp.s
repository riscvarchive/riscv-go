#include "textflag.h"

TEXT ·ReturnZero(SB),NOSPLIT,$0-8
	JMP ·returnZero(SB)
