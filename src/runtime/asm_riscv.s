// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

#include "textflag.h"

// int64 runtime·cputicks(void)
TEXT runtime·cputicks(SB),NOSPLIT,$0-8
	WORD	$0xc0102573	// rdtime a0
	MOV	A0, ret+0(FP)
	RET
