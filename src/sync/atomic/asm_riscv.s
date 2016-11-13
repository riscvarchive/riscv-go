// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

// RISC-V's atomic operations have two bits, aq ("acquire") and rl ("release"),
// which may be toggled on and off. Their precise semantics are defined in
// section 6.3 of the specification, but the basic idea is as follows:
//
//   - If neither aq nor rl is set, the CPU may reorder the atomic arbitrarily.
//     It guarantees only that it will execute atomically.
//
//   - If aq is set, the CPU may move the instruction backward, but not forward.
//
//   - If rl is set, the CPU may move the instruction forward, but not backward.
//
//   - If both are set, the CPU may not reorder the instruction at all.
//
// These four modes correspond to other well-known memory models on other CPUs.
// On ARM, aq corresponds to a dmb ishst, aq+rl corresponds to a dmb ish. On
// Intel, aq corresponds to an lfence, rl to an sfence, and aq+rl to an mfence
// (or a lock prefix).
//
// Go's memory model requires that
//   - if a read happens after a write, the read must observe the write, and
//     that
//   - if a read happens concurrently with a write, the read may observe the
//     write.
// aq is sufficient to guarantee this, so that's what we use here. (This jibes
// with ARM, which uses dmb ishst.)

#include "textflag.h"

TEXT ·SwapInt32(SB),NOSPLIT,$0-20
	JMP	·SwapUint32(SB)

TEXT ·SwapInt64(SB),NOSPLIT,$0-24
	JMP	·SwapUint64(SB)

TEXT ·SwapUint32(SB),NOSPLIT,$0-20
	MOV	addr+0(FP), A0
	MOVW	new+8(FP), A1
	WORD	$0x0cb525af	// amoswap.w.aq a1,a1,(a0)
	MOVW	A1, old+16(FP)
	RET

TEXT ·SwapUint64(SB),NOSPLIT,$0-24
	MOV	addr+0(FP), A0
	MOV	new+8(FP), A1
	WORD	$0x0cb535af	// amoswap.d.aq a1,a1,(a0)
	MOV	A1, old+16(FP)
	RET

TEXT ·SwapUintptr(SB),NOSPLIT,$0-24
	JMP	·SwapUint64(SB)

TEXT ·SwapPointer(SB),NOSPLIT,$0-24
	JMP	·SwapUint64(SB)

TEXT ·CompareAndSwapInt32(SB),NOSPLIT,$0-17
	JMP	·CompareAndSwapUint32(SB)

TEXT ·CompareAndSwapInt64(SB),NOSPLIT,$0-25
	JMP	·CompareAndSwapUint64(SB)

TEXT ·CompareAndSwapUint32(SB),NOSPLIT,$0-17
	MOV	addr+0(FP), A0
	MOVW	old+8(FP), A1
	MOVW	new+12(FP), A2
cas:
	WORD	$0x140522af	// lr.w.aq t0,(a0)
	BNE	T0, A1, fail
	WORD	$0x1cc5252f	// sc.w.aq a0,a2,(a0)
	// a0 = 0 iff the sc succeeded. Convert that to a boolean.
	SLTIU	$1, A0, A0
	MOV	A0, swapped+16(FP)
	RET
fail:
	MOV	$0, A0
	MOV	A0, swapped+16(FP)
	RET

TEXT ·CompareAndSwapUint64(SB),NOSPLIT,$0-25
	MOV	addr+0(FP), A0
	MOV	old+8(FP), A1
	MOV	new+16(FP), A2
cas:
	WORD	$0x140532af	// lr.d.aq t0,(a0)
	BNE	T0, A1, fail
	WORD	$0x1cc5352f	// sc.d.aq a0,a2,(a0)
	// a0 = 0 iff the sc succeeded, a0 = 0. Convert that to a boolean.
	SLTIU	$1, A0, A0
	MOV	A0, swapped+24(FP)
	RET
fail:
	MOV	$0, A0
	MOV	A0, swapped+24(FP)
	RET

TEXT ·CompareAndSwapUintptr(SB),NOSPLIT,$0-25
	JMP	·CompareAndSwapUint64(SB)

TEXT ·CompareAndSwapPointer(SB),NOSPLIT,$0-25
	JMP	·CompareAndSwapUint64(SB)

TEXT ·AddInt32(SB),NOSPLIT,$0-20
	JMP	·AddUint32(SB)

TEXT ·AddUint32(SB),NOSPLIT,$0-20
	MOV	addr+0(FP), A0
	MOVW	delta+8(FP), A1
	WORD	$0x04b5252f	// amoadd.w.aq a0,a1,(a0)
	ADD	A0, A1
	MOVW	A1, new+16(FP)
	RET

TEXT ·AddInt64(SB),NOSPLIT,$0-24
	JMP	·AddUint64(SB)

TEXT ·AddUint64(SB),NOSPLIT,$0-24
	MOV	addr+0(FP), A0
	MOV	delta+8(FP), A1
	WORD	$0x04b5352f	// amoadd.d.aq a0,a1,(a0)
	ADD	A0, A1
	MOVW	A1, new+16(FP)
	RET

TEXT ·AddUintptr(SB),NOSPLIT,$0-24
	JMP	·AddUint64(SB)

TEXT ·LoadInt32(SB),NOSPLIT,$0-12
	JMP	·LoadUint32(SB)

TEXT ·LoadInt64(SB),NOSPLIT,$0-16
	JMP	·LoadUint64(SB)

TEXT ·LoadUint32(SB),NOSPLIT,$0-12
	MOV	addr+0(FP), A0
	// Since addr is aligned (see comments in doc.go), this load is atomic
	// automatically.
	MOVW	(A0), A0
	MOVW	A0, val+8(FP)
	RET

TEXT ·LoadUint64(SB),NOSPLIT,$0-16
	MOV	addr+0(FP), A0
	// Since addr is aligned (see comments in doc.go), this load is atomic
	// automatically.
	MOV	(A0), A0
	MOV	A0, val+8(FP)
	RET

TEXT ·LoadUintptr(SB),NOSPLIT,$0-16
	JMP	·LoadUint64(SB)

TEXT ·LoadPointer(SB),NOSPLIT,$0-16
	JMP	·LoadUint64(SB)

TEXT ·StoreInt32(SB),NOSPLIT,$0-12
	JMP	·StoreUint32(SB)

TEXT ·StoreInt64(SB),NOSPLIT,$0-16
	JMP	·StoreUint64(SB)

TEXT ·StoreUint32(SB),NOSPLIT,$0-12
	MOV	addr+0(FP), A0
	MOVW	val+8(FP), A1
	// Since addr is aligned (see comments in doc.go), this store is atomic
	// automatically.
	MOVW	A1, (A0)
	RET

TEXT ·StoreUint64(SB),NOSPLIT,$0-16
	MOV	addr+0(FP), A0
	MOV	val+8(FP), A1
	// Since addr is aligned (see comments in doc.go), this store is atomic
	// automatically.
	MOV	A1, (A0)
	RET

TEXT ·StoreUintptr(SB),NOSPLIT,$0-16
	JMP	·StoreUint64(SB)

TEXT ·StorePointer(SB),NOSPLIT,$0-16
	JMP	·StoreUint64(SB)
