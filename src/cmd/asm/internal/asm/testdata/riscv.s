// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This input was created by hand.
// Modified to fit what plan 9 assemblers tend to look
// like, with 'destination' last, be it register target,
// or location.

TEXT memcpy(SB),0,$-24
	LD	dst+0(FP), T0
	LD	src+8(FP), T1
	LD	size+16(FP), T2

loop:
	BGE	T2, ZERO, done

	LBU	(T1), T4
	SB	(T0), T4

	// Plan 9 assemblers tend not to have data type attributes
	// in the instruction if they are well defined in the operand.
	// Skip the "I" since it's clearly immediate.
	ADD	$1, T0, T0
	ADD	$1, T1, T1
	ADD	$-1, T2, T2

	JMP	loop

done:
	RET
