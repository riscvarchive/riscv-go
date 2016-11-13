// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

#include "textflag.h"

TEXT ·Cas(SB), NOSPLIT, $0-17
	MOV	ptr+0(FP), A0
	MOVW	old+8(FP), A1
	MOVW	new+12(FP), A2
cas:
	WORD $0x140522af	// lr.w.aq t0,(a0)
	BNE	T0, A1, fail
	WORD $0x1cc5252f	// sc.w.aq a0,a2,(a0)
	// a0 = 0 iff the sc succeeded. Convert that to a boolean.
	SLTIU	$1, A0, A0
	MOV	A0, ret+16(FP)
	RET
fail:
	MOV	$0, A0
	MOV	A0, ret+16(FP)
	RET

TEXT ·Casp1(SB), NOSPLIT, $0-25
	MOV	ptr+0(FP), A0
	MOV	old+8(FP), A1
	MOV	new+16(FP), A2
cas:
	WORD $0x140532af	// lr.d.aq t0,(a0)
	BNE	T0, A1, fail
	WORD $0x1cc5352f	// sc.d.aq a0,a2,(a0)
	// a0 = 0 iff the sc succeeded. Convert that to a boolean.
	SLTIU	$1, A0, A0
	MOV	A0, ret+24(FP)
	RET
fail:
	MOV	$0, A0
	MOV	A0, ret+24(FP)
	RET

TEXT ·Casuintptr(SB),NOSPLIT,$0-25
	JMP ·Casp1(SB)

TEXT ·Storeuintptr(SB),NOSPLIT,$0-16
	MOV	ptr+0(FP), A0
	MOV	new+8(FP), A1
	// Since ptr is aligned, this store is atomic automatically.
	MOV	A1, (A0)
	RET

TEXT ·Loaduintptr(SB),NOSPLIT,$0-16
	MOV	ptr+0(FP), A0
	// Since ptr is aligned, this load is atomic automatically.
	MOV	(A0), A0
	MOV	A0, ret+8(FP)
	RET

TEXT ·Loaduint(SB),NOSPLIT,$0-16
	JMP ·Loaduintptr(SB)

TEXT ·Loadint64(SB),NOSPLIT,$0-16
	JMP ·Loaduintptr(SB)

TEXT ·Xaddint64(SB),NOSPLIT,$0-24
	MOV	ptr+0(FP), A0
	MOV	delta+8(FP), A1
	WORD $0x04b5352f	// amoadd.d.aq a0,a1,(a0)
	ADD	A0, A1, A0
	MOVW	A0, ret+16(FP)
	RET
