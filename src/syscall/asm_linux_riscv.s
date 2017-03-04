// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

//
// System calls for riscv64, Linux
//

// FIXME call entersyscall/exitsyscall once the runtime is working
// func Syscall(trap int64, a1, a2, a3 int64) (r1, r2, err int64)
TEXT ·Syscall(SB),NOSPLIT,$0-56
	//CALL	runtime·entersyscall(SB)
	MOV	a1+8(FP), A0
	MOV	a2+16(FP), A1
	MOV	a3+24(FP), A2
	MOV	$0, A3
	MOV	$0, A4
	MOV	$0, A5
	MOV	$0, A6
	MOV	trap+0(FP), A7	// syscall entry
	ECALL
	MOV	$-4096, T0
	BLTU	T0, A0, err
	MOV	A0, r1+32(FP)	// r1
	MOV	A1, r2+40(FP)	// r2
	MOV	ZERO, err+48(FP)	// errno
	//CALL	runtime·exitsyscall(SB)
	RET
err:
	MOV	$-1, T0
	MOV	T0, r1+32(FP)	// r1
	MOV	ZERO, r2+40(FP)	// r2
	SUB	A0, ZERO, A0
	MOV	A0, err+48(FP)	// errno
	//CALL	runtime·exitsyscall(SB)
	RET

// func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr)
TEXT ·Syscall6(SB),NOSPLIT,$0-80
	WORD $0

// func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2, err uintptr)
TEXT ·RawSyscall(SB),NOSPLIT,$0-56
	WORD $0

// func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr)
TEXT ·RawSyscall6(SB),NOSPLIT,$0-80
	WORD $0

// func socketcall(call int, a0, a1, a2, a3, a4, a5 uintptr) (n int, err int)
// Kernel interface gets call sub-number and pointer to a0.
TEXT ·socketcall(SB),NOSPLIT,$0-72
	WORD $0

// func rawsocketcall(call int, a0, a1, a2, a3, a4, a5 uintptr) (n int, err int)
// Kernel interface gets call sub-number and pointer to a0.
TEXT ·rawsocketcall(SB),NOSPLIT,$0-72
	WORD $0
