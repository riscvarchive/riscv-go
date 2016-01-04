// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This input was created by hand.

TEXT memcpy(SB),0,$-24
	// TODO(prattmic): RISCV gas assembly does register first, memory second
	// for load/store. The same is done here for now, for consistency with gas,
	// but we should probably switch to source first, destination second.
	// N.B. Even ARM, which typically uses destination first, source second
	// does the opposite in Go assembly.
	LD	T0, dst+0(FB)
	LD	T1, src+8(FB)
	LD	T2, size+16(FB)

loop:
	BGE	ZERO, T2, done

	LBU	T4, (T1)
	SB	T4, (T0)

	ADDI	T0, T0, $1
	ADDI	T1, T1, $1
	ADDI	T2, T2, $-1

	// Unconditional jump, discard link.
	// Should be pseudo-op 'J'.
	JAL	loop, ZERO

done:
	RET
