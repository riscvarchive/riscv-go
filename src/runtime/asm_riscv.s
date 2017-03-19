// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

// func rt0_go()
TEXT runtime·rt0_go(SB),NOSPLIT,$0
	// X2 = stack; A0 = argc; A1 = argv

	ADD	$-24, X2
	MOV	A0, 8(X2) // argc
	MOV	A1, 16(X2) // argv

	// create istack out of the given (operating system) stack.
	// _cgo_init may update stackguard.
	MOV	$runtime·g0(SB), g
	MOV	$(-64*1024), T0
	ADD	T0, X2, T1
	MOV	T1, g_stackguard0(g)
	MOV	T1, g_stackguard1(g)
	MOV	T1, (g_stack+stack_lo)(g)
	MOV	X2, (g_stack+stack_hi)(g)

	// if there is a _cgo_init, call it using the gcc ABI.
	MOV	_cgo_init(SB), T0
	BEQ	T0, ZERO, nocgo

	MOV	ZERO, A3	// arg 3: not used
	MOV	ZERO, A2	// arg 2: not used
	MOV	$setg_gcc<>(SB), A1	// arg 1: setg
	MOV	g, A0	// arg 0: G
	JALR	RA, T0

nocgo:
	// update stackguard after _cgo_init
	MOV	(g_stack+stack_lo)(g), T0
	ADD	$const__StackGuard, T0
	MOV	T0, g_stackguard0(g)
	MOV	T0, g_stackguard1(g)

	// set the per-goroutine and per-mach "registers"
	MOV	$runtime·m0(SB), T0

	// save m->g0 = g0
	MOV	g, m_g0(T0)
	// save m0 to g0->m
	MOV	T0, g_m(g)

	CALL	runtime·check(SB)

	// args are already prepared
	CALL	runtime·args(SB)
	CALL	runtime·osinit(SB)
	CALL	runtime·schedinit(SB)

	// create a new goroutine to start program
	MOV	$runtime·mainPC(SB), T0		// entry
	ADD	$-24, X2
	MOV	T0, 16(X2)
	MOV	ZERO, 8(X2)
	MOV	ZERO, 0(X2)
	CALL	runtime·newproc(SB)
	ADD	$24, X2

	// start this M
	CALL	runtime·mstart(SB)

	WORD $0 // crash if reached
	RET

// void setg_gcc(G*); set g called from gcc with g in A0
TEXT setg_gcc<>(SB),NOSPLIT,$0-0
	MOV	A0, g
	CALL	runtime·save_g(SB)
	RET

// func cputicks() int64
TEXT runtime·cputicks(SB),NOSPLIT,$0-8
	WORD	$0xc0102573	// rdtime a0
	MOV	A0, ret+0(FP)
	RET

// systemstack_switch is a dummy routine that systemstack leaves at the bottom
// of the G stack. We need to distinguish the routine that
// lives at the bottom of the G stack from the one that lives
// at the top of the system stack because the one at the top of
// the system stack terminates the stack walk (see topofstack()).
TEXT runtime·systemstack_switch(SB), NOSPLIT, $0-0
	UNDEF
	JALR	RA, ZERO	// make sure this function is not leaf
	RET

// func systemstack(fn func())
TEXT runtime·systemstack(SB), NOSPLIT, $0-8
	MOV	fn+0(FP), CTXT	// CTXT = fn
	MOV	g_m(g), T0	// T0 = m

	MOV	m_gsignal(T0), T1	// T1 = gsignal
	BEQ	g, T1, noswitch

	MOV	m_g0(T0), T1	// T1 = g0
	BEQ	g, T1, noswitch

	MOV	m_curg(T0), T2
	BEQ	g, T2, switch

	// Bad: g is not gsignal, not g0, not curg. What is it?
	// Hide call from linker nosplit analysis.
	MOV	$runtime·badsystemstack(SB), T1
	JALR	RA, T1

switch:
	// save our state in g->sched. Pretend to
	// be systemstack_switch if the G stack is scanned.
	MOV	$runtime·systemstack_switch(SB), T2
	ADD	$8, T2	// get past prologue
	MOV	T2, (g_sched+gobuf_pc)(g)
	MOV	X2, (g_sched+gobuf_sp)(g)
	MOV	ZERO, (g_sched+gobuf_lr)(g)
	MOV	g, (g_sched+gobuf_g)(g)

	// switch to g0
	MOV	T1, g
	CALL	runtime·save_g(SB)
	MOV	(g_sched+gobuf_sp)(g), T0
	// make it look like mstart called systemstack on g0, to stop traceback
	ADD	$-16, T0
	AND	$~15, T0
	MOV	$runtime·mstart(SB), T1
	MOV	T1, 0(T0)
	MOV	T0, X2

	// call target function
	MOV	0(CTXT), T1	// code pointer
	JALR	RA, T1

	// switch back to g
	MOV	g_m(g), T0
	MOV	m_curg(T0), g
	CALL	runtime·save_g(SB)
	MOV	(g_sched+gobuf_sp)(g), X2
	MOV	ZERO, (g_sched+gobuf_sp)(g)
	RET

noswitch:
	// already on m stack, just call directly
	MOV	0(CTXT), T1	// code pointer
	JALR	RA, T1
	RET

// func getcallerpc(argp unsafe.Pointer) uintptr
TEXT runtime·getcallerpc(SB),NOSPLIT,$8-16
	MOV	16(X2), T0		// LR saved by caller
	MOV	runtime·stackBarrierPC(SB), T1
	BNE	T0, T1, nobar
	// Get original return PC.
	CALL	runtime·nextBarrierPC(SB)
	MOV	8(X2), T0
nobar:
	MOV	T0, ret+8(FP)
	RET

// func fastrand() uint32
TEXT runtime·fastrand(SB),NOSPLIT,$0-4
	MOV	g_m(g), A2
	MOVWU	m_fastrand(A2), A1
	ADD	A1, A1
	// TODO(sorear): Just use ADDW once an encoding is added
	SLL	$32, A1
	SRA	$32, A1
	BGE	A1, ZERO, noxor
	MOV	$0x88888eef - 1<<32, A0
	XOR	A0, A1
noxor:
	MOVW	A1, m_fastrand(A2)
	MOVW	A1, ret+0(FP)
	RET

// eqstring tests whether two strings are equal.
// The compiler guarantees that strings passed
// to eqstring have equal length.
// See runtime_test.go:eqstring_generic for
// equivalent Go code.

// func eqstring(s1, s2 string) bool
TEXT runtime·eqstring(SB),NOSPLIT,$0-33
	MOV	s1_base+0(FP), T0
	MOV	s2_base+16(FP), T1
	MOV	$1, T2
	MOVB	T2, ret+32(FP)
	BNE	T0, T1, diff_len
	RET
diff_len:
	MOV	s1_len+8(FP), T2
	ADD	T0, T2, T3
loop:
	BNE	T0, T3, 2(PC)
	RET
	MOVBU	(T0), T5
	ADD	$1, T0
	MOVBU	(T1), T6
	ADD	$1, T1
	BEQ	T5, T6, loop
	MOVB	ZERO, ret+32(FP)
	RET

// func morestack()
TEXT runtime·morestack(SB),NOSPLIT,$-4-0
	WORD $0

// func return0()
TEXT runtime·return0(SB), NOSPLIT, $0
	MOV	$0, A0
	RET

// func memequal(a, b unsafe.Pointer, size uintptr) bool
TEXT runtime·memequal(SB),NOSPLIT,$-8-25
	MOV	a+0(FP), A1
	MOV	b+8(FP), A2
	BEQ	A1, A2, eq
	MOV	size+16(FP), A3
	ADD	A1, A3, A4
loop:
	BNE	A1, A4, test
	MOV	$1, A1
	MOVB	A1, ret+24(FP)
	RET
test:
	MOVBU	(A1), A6
	ADD	$1, A1
	MOVBU	(A2), A7
	ADD	$1, A2
	BEQ	A6, A7, loop

	MOVB	ZERO, ret+24(FP)
	RET
eq:
	MOV	$1, A1
	MOVB	A1, ret+24(FP)
	RET

// func memequal_varlen(a, b unsafe.Pointer) bool
TEXT runtime·memequal_varlen(SB),NOSPLIT,$40-17
	MOV	a+0(FP), A1
	MOV	b+8(FP), A2
	BEQ	A1, A2, eq
	MOV	8(CTXT), A3    // compiler stores size at offset 8 in the closure
	MOV	A1, 8(X2)
	MOV	A2, 16(X2)
	MOV	A3, 24(X2)
	CALL	runtime·memequal(SB)
	MOVBU	32(X2), A1
	MOVB	A1, ret+16(FP)
	RET
eq:
	MOV	$1, A1
	MOVB	A1, ret+16(FP)
	RET

// restore state from Gobuf; longjmp

// func gogo(buf *gobuf)
TEXT runtime·gogo(SB), NOSPLIT, $16-8
	MOV	buf+0(FP), T0

	// If ctxt is not nil, invoke deletion barrier before overwriting.
	MOV	gobuf_ctxt(T0), T1
	BEQ	T1, ZERO, nilctxt
	ADD	$gobuf_ctxt, T0, T1
	MOV	T1, 8(X2)
	MOV	ZERO, 16(X2)
	CALL	runtime·writebarrierptr_prewrite(SB)
	MOV	buf+0(FP), T0

nilctxt:
	MOV	gobuf_g(T0), g	// make sure g is not nil
	CALL	runtime·save_g(SB)

	MOV	(g), ZERO // make sure g is not nil
	MOV	gobuf_sp(T0), X2
	MOV	gobuf_lr(T0), RA
	MOV	gobuf_ret(T0), A0
	MOV	gobuf_ctxt(T0), CTXT
	MOV	ZERO, gobuf_sp(T0)
	MOV	ZERO, gobuf_ret(T0)
	MOV	ZERO, gobuf_lr(T0)
	MOV	ZERO, gobuf_ctxt(T0)
	MOV	gobuf_pc(T0), T0
	JALR	ZERO, T0

// func jmpdefer(fv *funcval, argp uintptr)
// called from deferreturn
// 1. grab stored return address from the caller's frame
// 2. sub 12 bytes to get back to JAL deferreturn
// 3. JMP to fn
// TODO(sorear): There are shorter jump sequences.  This function will need to be updated when we use them.
TEXT runtime·jmpdefer(SB), NOSPLIT, $-8-16
	MOV	0(X2), RA
	ADD	$-12, RA

	MOV	fv+0(FP), CTXT
	MOV	argp+8(FP), X2
	ADD	$-8, X2
	MOV	0(CTXT), T0
	JALR	ZERO, T0

// func procyield(cycles uint32)
TEXT runtime·procyield(SB),NOSPLIT,$0-0
	RET

// Switch to m->g0's stack, call fn(g).
// Fn must never return. It should gogo(&g->sched)
// to keep running g.

// func mcall(fn func(*g))
TEXT runtime·mcall(SB), NOSPLIT, $-8-8
	// Save caller state in g->sched
	MOV	X2, (g_sched+gobuf_sp)(g)
	MOV	RA, (g_sched+gobuf_pc)(g)
	MOV	ZERO, (g_sched+gobuf_lr)(g)
	MOV	g, (g_sched+gobuf_g)(g)

	// Switch to m->g0 & its stack, call fn.
	MOV	g, T0
	MOV	g_m(g), T1
	MOV	m_g0(T1), g
	CALL	runtime·save_g(SB)
	BNE	g, T0, 2(PC)
	JMP	runtime·badmcall(SB)
	MOV	fn+0(FP), CTXT			// context
	MOV	0(CTXT), T1			// code pointer
	MOV	(g_sched+gobuf_sp)(g), X2	// sp = m->g0->sched.sp
	ADD	$-16, X2
	MOV	T0, 8(X2)
	MOV	ZERO, 0(X2)
	JALR	RA, T1
	JMP	runtime·badmcall2(SB)

// func gosave(buf *gobuf)
// save state in Gobuf; setjmp
TEXT runtime·gosave(SB), NOSPLIT, $-8-8
	MOV	buf+0(FP), T1
	MOV	X2, gobuf_sp(T1)
	MOV	RA, gobuf_pc(T1)
	MOV	g, gobuf_g(T1)
	MOV	ZERO, gobuf_lr(T1)
	MOV	ZERO, gobuf_ret(T1)
	// Assert ctxt is zero. See func save.
	MOV	gobuf_ctxt(T1), T1
	BEQ	T1, ZERO, 2(PC)
	CALL	runtime·badctxt(SB)
	RET

// func asmcgocall(fn, arg unsafe.Pointer) int32
TEXT ·asmcgocall(SB),NOSPLIT,$0-12
	WORD $0

// func memhash_varlen(p unsafe.Pointer, h uintptr) uintptr
TEXT runtime·memhash_varlen(SB),NOSPLIT,$40-24
	WORD $0

// func asminit()
TEXT runtime·asminit(SB),NOSPLIT,$-8-0
	RET

// reflectcall: call a function with the given argument list
// func call(argtype *_type, f *FuncVal, arg *byte, argsize, retoffset uint32).
// we don't have variable-sized frames, so we use a small number
// of constant-sized-frame functions to encode a few bits of size in the pc.
// Caution: ugly multiline assembly macros in your future!

#define DISPATCH(NAME,MAXSIZE)	\
	MOV	$MAXSIZE, T1	\
	BLTU	T1, T0, 3(PC)	\
	MOV	$NAME(SB), T2;	\
	JALR	ZERO, T2
// Note: can't just "BR NAME(SB)" - bad inlining results.

// func call(argtype *rtype, fn, arg unsafe.Pointer, n uint32, retoffset uint32)
TEXT reflect·call(SB), NOSPLIT, $0-0
	JMP	·reflectcall(SB)

// func reflectcall(argtype *_type, fn, arg unsafe.Pointer, argsize uint32, retoffset uint32)
TEXT ·reflectcall(SB), NOSPLIT, $-8-32
	MOVWU argsize+24(FP), T0
	DISPATCH(runtime·call32, 32)
	DISPATCH(runtime·call64, 64)
	DISPATCH(runtime·call128, 128)
	DISPATCH(runtime·call256, 256)
	DISPATCH(runtime·call512, 512)
	DISPATCH(runtime·call1024, 1024)
	DISPATCH(runtime·call2048, 2048)
	DISPATCH(runtime·call4096, 4096)
	DISPATCH(runtime·call8192, 8192)
	DISPATCH(runtime·call16384, 16384)
	DISPATCH(runtime·call32768, 32768)
	DISPATCH(runtime·call65536, 65536)
	DISPATCH(runtime·call131072, 131072)
	DISPATCH(runtime·call262144, 262144)
	DISPATCH(runtime·call524288, 524288)
	DISPATCH(runtime·call1048576, 1048576)
	DISPATCH(runtime·call2097152, 2097152)
	DISPATCH(runtime·call4194304, 4194304)
	DISPATCH(runtime·call8388608, 8388608)
	DISPATCH(runtime·call16777216, 16777216)
	DISPATCH(runtime·call33554432, 33554432)
	DISPATCH(runtime·call67108864, 67108864)
	DISPATCH(runtime·call134217728, 134217728)
	DISPATCH(runtime·call268435456, 268435456)
	DISPATCH(runtime·call536870912, 536870912)
	DISPATCH(runtime·call1073741824, 1073741824)
	MOV	$runtime·badreflectcall(SB), T2
	JALR	ZERO, T2

#define CALLFN(NAME,MAXSIZE)			\
TEXT NAME(SB), WRAPPER, $MAXSIZE-24;		\
	NO_LOCAL_POINTERS;			\
	/* copy arguments to stack */		\
	MOV	arg+16(FP), A1;			\
	MOVWU	argsize+24(FP), A2;		\
	MOV	X2, A3;				\
	ADD	$8, A3;				\
	ADD	A3, A2;				\
	BEQ	A3, A2, 6(PC);			\
	MOVBU	(A1), A4;			\
	ADD	$1, A1;				\
	MOVB	A4, (A3);			\
	ADD	$1, A3;				\
	JMP	-5(PC);				\
	/* call function */			\
	MOV	f+8(FP), CTXT;			\
	MOV	(CTXT), A4;			\
	PCDATA  $PCDATA_StackMapIndex, $0;	\
	JALR	RA, A4;				\
	/* copy return values back */		\
	MOV	argtype+0(FP), A5;		\
	MOV	arg+16(FP), A1;			\
	MOVWU	n+24(FP), A2;			\
	MOVWU	retoffset+28(FP), A4;		\
	ADD	$8, X2, A3;			\
	ADD	A4, A3; 			\
	ADD	A4, A1;				\
	SUB	A4, A2;				\
	CALL	callRet<>(SB);			\
	RET

// callRet copies return values back at the end of call*. This is a
// separate function so it can allocate stack space for the arguments
// to reflectcallmove. It does not follow the Go ABI; it expects its
// arguments in registers.
TEXT callRet<>(SB), NOSPLIT, $32-0
	MOV	A5, 8(X2)
	MOV	A1, 16(X2)
	MOV	A3, 24(X2)
	MOV	A2, 32(X2)
	CALL	runtime·reflectcallmove(SB)
	RET

CALLFN(·call16, 16)
CALLFN(·call32, 32)
CALLFN(·call64, 64)
CALLFN(·call128, 128)
CALLFN(·call256, 256)
CALLFN(·call512, 512)
CALLFN(·call1024, 1024)
CALLFN(·call2048, 2048)
CALLFN(·call4096, 4096)
CALLFN(·call8192, 8192)
CALLFN(·call16384, 16384)
CALLFN(·call32768, 32768)
CALLFN(·call65536, 65536)
CALLFN(·call131072, 131072)
CALLFN(·call262144, 262144)
CALLFN(·call524288, 524288)
CALLFN(·call1048576, 1048576)
CALLFN(·call2097152, 2097152)
CALLFN(·call4194304, 4194304)
CALLFN(·call8388608, 8388608)
CALLFN(·call16777216, 16777216)
CALLFN(·call33554432, 33554432)
CALLFN(·call67108864, 67108864)
CALLFN(·call134217728, 134217728)
CALLFN(·call268435456, 268435456)
CALLFN(·call536870912, 536870912)
CALLFN(·call1073741824, 1073741824)

// func goexit(neverCallThisFunction)
// The top-most function running on a goroutine
// returns to goexit+PCQuantum.
TEXT runtime·goexit(SB),NOSPLIT,$-8-0
	MOV	ZERO, ZERO	// NOP
	CALL	runtime·goexit1(SB)	// does not return
	// traceback from goexit1 must hit code range of goexit
	MOV	ZERO, ZERO	// NOP

// func setcallerpc(argp unsafe.Pointer, pc uintptr)
TEXT runtime·setcallerpc(SB),NOSPLIT,$8-16
	MOV	pc+8(FP), A1
	MOV	16(X2), A2
	MOV	runtime·stackBarrierPC(SB), A3
	BEQ	A2, A3, setbar
	MOV	A1, 16(X2)		// set LR in caller
	RET
setbar:
	// Set the stack barrier return PC.
	MOV	A1, 8(X2)
	CALL	runtime·setNextBarrierPC(SB)
	RET

// func IndexByte(s []byte, c byte) int
TEXT bytes·IndexByte(SB),NOSPLIT,$0-40
	MOV	s+0(FP), A1
	MOV	s_len+8(FP), A2
	MOVBU	c+24(FP), A3	// byte to find
	MOV	A1, A4		// store base for later
	ADD	A1, A2		// end
	ADD	$-1, A1

loop:
	ADD	$1, A1
	BEQ	A1, A2, notfound
	MOVBU	(A1), A5
	BNE	A3, A5, loop

	SUB	A4, A1		// remove base
	MOV	A1, ret+32(FP)
	RET

notfound:
	MOV	$-1, A1
	MOV	A1, ret+32(FP)
	RET

// func IndexByte(s string, c byte) int
TEXT strings·IndexByte(SB),NOSPLIT,$0-32
	MOV	p+0(FP), A1
	MOV	b_len+8(FP), A2
	MOVBU	c+16(FP), A3	// byte to find
	MOV	A1, A4		// store base for later
	ADD	A1, A2		// end
	ADD	$-1, A1

loop:
	ADD	$1, A1
	BEQ	A1, A2, notfound
	MOVBU	(A1), A5
	BNE	A3, A5, loop

	SUB	A4, A1		// remove base
	MOV	A1, ret+24(FP)
	RET

notfound:
	MOV	$-1, A1
	MOV	A1, ret+24(FP)
	RET

// TODO: share code with memequal?
// func Equal(a, b []byte) bool
TEXT bytes·Equal(SB),NOSPLIT,$0-49
	MOV	a_len+8(FP), A3
	MOV	b_len+32(FP), A4
	BNE	A3, A4, noteq		// unequal lengths are not equal

	MOV	a+0(FP), A1
	MOV	b+24(FP), A2
	ADD	A1, A3		// end

loop:
	BEQ	A1, A3, equal		// reached the end
	MOVBU	(A1), A6
	ADD	$1, A1
	MOVBU	(A2), A7
	ADD	$1, A2
	BEQ	A6, A7, loop

noteq:
	MOVB	ZERO, ret+48(FP)
	RET

equal:
	MOV	$1, A1
	MOVB	A1, ret+48(FP)
	RET

TEXT runtime·stackBarrier(SB),NOSPLIT,$0
	WORD $0

// func cgocallback_gofunc(fv uintptr, frame uintptr, framesize, ctxt uintptr)
TEXT ·cgocallback_gofunc(SB),NOSPLIT,$24-32
	WORD $0

TEXT runtime·prefetcht0(SB),NOSPLIT,$0-8
	RET

TEXT runtime·prefetcht1(SB),NOSPLIT,$0-8
	RET

TEXT runtime·prefetcht2(SB),NOSPLIT,$0-8
	RET

TEXT runtime·prefetchnta(SB),NOSPLIT,$0-8
	RET

TEXT runtime·breakpoint(SB),NOSPLIT,$-8-0
	EBREAK
	RET

// void setg(G*); set g. for use by needm.
TEXT runtime·setg(SB), NOSPLIT, $0-8
	MOV	gg+0(FP), g
	// This only happens if iscgo, so jump straight to save_g
	CALL	runtime·save_g(SB)
	RET

TEXT ·checkASM(SB),NOSPLIT,$0-1
	MOV	$1, T0
	MOV	T0, ret+0(FP)
	RET

DATA	runtime·mainPC+0(SB)/8,$runtime·main(SB)
GLOBL	runtime·mainPC(SB),RODATA,$8
