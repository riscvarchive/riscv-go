// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT _rt0_riscv_linux(SB),NOSPLIT,$-8
	MOV	0(SP), A0	// argc
	ADD	$8, SP, A1	// argv
	CALL	main(SB)

TEXT main(SB),NOSPLIT,$-8
	MOV	$runtime·rt0_go(SB), T0
	JALR	RA, T0
exit:
	MOV	$0, A0
	MOV	$94, A7	// sys_exit
	ECALL
