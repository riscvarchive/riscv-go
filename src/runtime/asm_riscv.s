// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

#include "textflag.h"

// func rt0_go()
TEXT runtime·rt0_go(SB),NOSPLIT,$0
	CALL	runtime·schedinit(SB)
	CALL	runtime·mstart(SB)

// func cputicks() int64
TEXT runtime·cputicks(SB),NOSPLIT,$0-8
	WORD	$0xc0102573	// rdtime a0
	MOV	A0, ret+0(FP)
	RET

// func systemstack(fn func())
TEXT runtime·systemstack(SB), NOSPLIT, $0-8
	WORD $0
// func getcallerpc(argp unsafe.Pointer) uintptr
TEXT runtime·getcallerpc(SB),NOSPLIT,$8-16
	WORD $0
// func fastrand() uint32
TEXT runtime·fastrand(SB),NOSPLIT,$-8-4
	WORD $0
// func eqstring(s1, s2 string) bool
TEXT runtime·eqstring(SB),NOSPLIT,$0-33
	WORD $0
// func morestack()
TEXT runtime·morestack(SB),NOSPLIT,$-8-0
	WORD $0
// func return0()
TEXT runtime·return0(SB), NOSPLIT, $0
	WORD $0
// func memequal(a, b unsafe.Pointer, size uintptr) bool
TEXT runtime·memequal(SB),NOSPLIT,$-8-25
	WORD $0
// func gogo(buf *gobuf)
TEXT runtime·gogo(SB), NOSPLIT, $24-8
	WORD $0
// func jmpdefer(fv *funcval, argp uintptr)
TEXT runtime·jmpdefer(SB), NOSPLIT, $-8-16
	WORD $0
// func procyield(cycles uint32)
TEXT runtime·procyield(SB),NOSPLIT,$0-0
	WORD $0
// func mcall(fn func(*g))
TEXT runtime·mcall(SB), NOSPLIT, $-8-8
	WORD $0
// func gosave(buf *gobuf)
TEXT runtime·gosave(SB), NOSPLIT, $-8-8
	WORD $0
// func asmcgocall(fn, arg unsafe.Pointer) int32
TEXT ·asmcgocall(SB),NOSPLIT,$0-20
	WORD $0
// func memhash_varlen(p unsafe.Pointer, h uintptr) uintptr
TEXT runtime·memhash_varlen(SB),NOSPLIT,$40-24
	WORD $0
// func memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen(SB),NOSPLIT,$40-17
	WORD $0
// func asminit()
TEXT runtime·asminit(SB),NOSPLIT,$-8-0
	WORD $0
// func publicationBarrier()
TEXT ·publicationBarrier(SB),NOSPLIT|NOFRAME,$0-0
	WORD $0
// func reflectcall(argtype *_type, fn, arg unsafe.Pointer, argsize uint32, retoffset uint32)
TEXT ·reflectcall(SB), NOSPLIT, $-8-32
	WORD $0
// func goexit(neverCallThisFunction)
TEXT runtime·goexit(SB),NOSPLIT,$-8-0
	WORD $0
TEXT reflect·call(SB), NOSPLIT, $0-0
	WORD $0
// func cmpstring(s1, s2 string) int
TEXT runtime·cmpstring(SB),NOSPLIT,$-4-40
	WORD $0
// func setcallerpc(argp unsafe.Pointer, pc uintptr)
TEXT runtime·setcallerpc(SB),NOSPLIT,$8-16
	WORD $0
// func IndexByte(s []byte, c byte) int
TEXT bytes·IndexByte(SB),NOSPLIT,$0-40
	WORD $0
// func IndexByte(s string, c byte) int
TEXT strings·IndexByte(SB),NOSPLIT,$0-32
	WORD $0
// func Equal(a, b []byte) bool
TEXT bytes·Equal(SB),NOSPLIT,$0-49
	WORD $0
// func stackBarrier()
TEXT runtime·stackBarrier(SB),NOSPLIT,$0
	WORD $0
// func systemstack_switch()
TEXT runtime·systemstack_switch(SB),NOSPLIT,$0-0
	WORD $0
// func cgocallback_gofunc(fv uintptr, frame uintptr, framesize, ctxt uintptr)
TEXT ·cgocallback_gofunc(SB),NOSPLIT,$24-32
	WORD $0
