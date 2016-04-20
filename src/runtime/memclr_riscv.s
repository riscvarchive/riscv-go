// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

#include "textflag.h"

// void runtime·memclr(void*, uintptr)
TEXT runtime·memclr(SB),NOSPLIT,$0-16
	MOV	0(FP), T0
	MOV	8(FP), T1
	ADD	T0, T1, T3

out:
	BEQ	T0, T3, done
	MOVB	ZERO, (T0)
	ADD	$1, T0
	JMP	out
done:
	RET
