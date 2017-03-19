// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// System calls and other sys.stuff for mips64, Linux
//

#include "textflag.h"
#include "go_asm.h"

#define AT_FDCWD -100

#define SYS_clock_gettime	113
#define SYS_clone		220
#define SYS_close		57
#define SYS_connect		203
#define SYS_epoll_create1	20
#define SYS_epoll_ctl		21
#define SYS_epoll_pwait		22
#define SYS_exit		93
#define SYS_exit_group		94
#define SYS_faccessat		48
#define SYS_fcntl		25
#define SYS_futex		98
#define SYS_getpid		172
#define SYS_getrlimit		163
#define SYS_gettid		178
#define SYS_gettimeofday	169
#define SYS_kill		129
#define SYS_madvise		233
#define SYS_mincore		232
#define SYS_mmap		222
#define SYS_munmap		215
#define SYS_nanosleep		101
#define SYS_openat		56
#define SYS_pselect6		72
#define SYS_read		63
#define SYS_rt_sigaction	134
#define SYS_rt_sigprocmask	135
#define SYS_rt_sigreturn	139
#define SYS_sched_getaffinity	123
#define SYS_sched_yield		124
#define SYS_setitimer		103
#define SYS_sigaltstack		132
#define SYS_socket		198
#define SYS_tkill		130
#define SYS_write		64

#define ERR_ABORT \
	MOV	$-4096, T0 \
	BLTU	A0, T0, 2(PC) \
	WORD $0

#define ERR_RETURN_M1(slot) \
	MOV	$-4096, T0 \
	BGEU	T0, A0, 2(PC) \
	MOV	$-1, A0 \
	MOVW	A0, slot(FP)

#define ERR_RETURN_M1_64(slot) \
	MOV	$-4096, T0 \
	BGEU	T0, A0, 2(PC) \
	MOV	$-1, A0 \
	MOV	A0, slot(FP)

#define ERR_RETURN_ABS(slot) \
	MOV	$-4096, T0 \
	BGEU	T0, A0, 2(PC) \
	SUB	A0, ZERO, A0 \
	MOVW	A0, slot(FP)

#define ERR_RETURN_ABS_64(slot) \
	MOV	$-4096, T0 \
	BGEU	T0, A0, 2(PC) \
	SUB	A0, ZERO, A0 \
	MOV	A0, slot(FP)

// func exit(code int32)
TEXT runtime·exit(SB),NOSPLIT,$-8-4
	MOVW	code+0(FP), A0
	MOV	$SYS_exit_group, A7
	ECALL
	RET

// func exit1(code int32)
//TEXT runtime·exit1(SB),NOSPLIT,$-8-4
//	MOVW	code+0(FP), R0
//	MOVD	$SYS_exit, R8
//	SVC
//	RET

// func open(name *byte, mode, perm int32) int32
TEXT runtime·open(SB),NOSPLIT,$-8-20
	MOV	$AT_FDCWD, A0
	MOV	name+0(FP), A1
	MOVW	mode+8(FP), A2
	MOVW	perm+12(FP), A3
	MOV	$SYS_openat, A7
	ECALL
	ERR_RETURN_M1(ret+16)
	RET

// func closefd(fd int32) int32
TEXT runtime·closefd(SB),NOSPLIT,$-8-12
	MOVW	fd+0(FP), A0
	MOV	$SYS_close, A7
	ECALL
	ERR_RETURN_M1(ret+8)
	RET

// func write(fd uintptr, p unsafe.Pointer, n int32) int32
TEXT runtime·write(SB),NOSPLIT,$-8-28
	MOV	fd+0(FP), A0
	MOV	p+8(FP), A1
	MOVW	n+16(FP), A2
	MOV	$SYS_write, A7
	ECALL
	ERR_RETURN_M1(ret+24)
	RET

// func read(fd int32, p unsafe.Pointer, n int32) int32
TEXT runtime·read(SB),NOSPLIT,$-8-28
	MOVW	fd+0(FP), A0
	MOV	p+8(FP), A1
	MOVW	n+16(FP), A2
	MOV	$SYS_read, A7
	ECALL
	ERR_RETURN_M1(ret+24)
	RET

// func getrlimit(kind int32, limit unsafe.Pointer) int32
TEXT runtime·getrlimit(SB),NOSPLIT,$-8-20
	MOVW	kind+0(FP), A0
	MOV	limit+8(FP), A1
	MOV	$SYS_getrlimit, A7
	ECALL
	MOVW	A0, ret+16(FP)
	RET

// func usleep(usec uint32)
TEXT runtime·usleep(SB),NOSPLIT,$24-4
	MOVWU	usec+0(FP), A0
	MOV	$1000, A1
	MUL	A1, A0, A0
	MOV	$1000000000, A1
	DIV	A1, A0, A2
	MOV	A2, 8(X2)
	REM	A1, A0, A3
	MOV	A3, 16(X2)
	ADD	$8, X2, A0
	MOV	ZERO, A1
	MOV	$SYS_nanosleep, A7
	ECALL
	RET

// func gettid() uint32
TEXT runtime·gettid(SB),NOSPLIT,$0-4
	MOV	$SYS_gettid, A7
	ECALL
	MOVW	A0, ret+0(FP)
	RET

// func raise(sig uint32)
TEXT runtime·raise(SB),NOSPLIT,$-8
	MOV	$SYS_gettid, A7
	ECALL
	// MOVW	A0, A0	// arg 1 tid
	MOVW	sig+0(FP), A1	// arg 2
	MOV	$SYS_tkill, A7
	ECALL
	RET

// func raiseproc(sig uint32)
TEXT runtime·raiseproc(SB),NOSPLIT,$-8
	MOV	$SYS_getpid, A7
	ECALL
	// MOVW	A0, A0		// arg 1 pid
	MOVW	sig+0(FP), A1	// arg 2
	MOV	$SYS_kill, A7
	ECALL
	RET

// func setitimer(mode int32, new, old *itimerval)
TEXT runtime·setitimer(SB),NOSPLIT,$-8-24
	MOVW	mode+0(FP), A0
	MOV	new+8(FP), A1
	MOV	old+16(FP), A2
	MOV	$SYS_setitimer, A7
	ECALL
	RET

// func mincore(addr unsafe.Pointer, n uintptr, dst *byte) int32
TEXT runtime·mincore(SB),NOSPLIT,$-8-28
	MOV	addr+0(FP), A0
	MOV	n+8(FP), A1
	MOV	dst+16(FP), A2
	MOV	$SYS_mincore, A7
	ECALL
	MOVW	A0, ret+24(FP)
	RET

// func walltime() (sec int64, nsec int32)
TEXT runtime·walltime(SB),NOSPLIT,$24-12
	MOV	$0, A0 // CLOCK_REALTIME
	MOV	X2, A1
	MOV	$SYS_clock_gettime, A7
	ECALL
	MOV	0(X2), T0	// sec
	MOV	8(X2), T1	// nsec
	MOV	T0, sec+0(FP)
	MOVW	T1, nsec+8(FP)
	RET

// func nanotime() int64
TEXT runtime·nanotime(SB),NOSPLIT,$24-8
	MOV	$1, A0 // CLOCK_MONOTONIC
	MOV	X2, A1
	MOV	$SYS_clock_gettime, A7
	ECALL
	MOV	0(X2), T0	// sec
	MOV	8(X2), T1	// nsec
	// sec is in T0, nsec in T1
	// return nsec in T0
	MOV	$1000000000, T2
	MUL	T2, T0
	ADD	T1, T0
	MOV	T0, ret+0(FP)
	RET

// func rtsigprocmask(how int32, new, old *sigset, size int32)
TEXT runtime·rtsigprocmask(SB),NOSPLIT,$-8-28
	MOVW	how+0(FP), A0
	MOV	new+8(FP), A1
	MOV	old+16(FP), A2
	MOVW	size+24(FP), A3
	MOV	$SYS_rt_sigprocmask, A7
	ECALL
	ERR_ABORT
	RET

// func rt_sigaction(sig uintptr, new, old *sigactiont, size uintptr) int32
TEXT runtime·rt_sigaction(SB),NOSPLIT,$-8-36
	MOV	sig+0(FP), A0
	MOV	new+8(FP), A1
	MOV	old+16(FP), A2
	MOV	size+24(FP), A3
	MOV	$SYS_rt_sigaction, A7
	ECALL
	MOVW	A0, ret+32(FP)
	RET

// func sigfwd(fn uintptr, sig uint32, info *siginfo, ctx unsafe.Pointer)
TEXT runtime·sigfwd(SB),NOSPLIT,$0-32
	MOVW	sig+8(FP), A0
	MOV	info+16(FP), A1
	MOV	ctx+24(FP), A2
	MOV	fn+0(FP), T1
	JALR	RA, T1
	RET

// func sigtramp(ureg, note unsafe.Pointer)
TEXT runtime·sigtramp(SB),NOSPLIT,$24
	// this might be called in external code context,
	// where g is not set.
	// first save R0, because runtime·load_g will clobber it
	MOVW	A0, 8(X2)
	MOVBU	runtime·iscgo(SB), A0
	BEQ	A0, ZERO, 2(PC)
	CALL	runtime·load_g(SB)

	MOV	A1, 16(X2)
	MOV	A2, 24(X2)
	MOV	$runtime·sigtrampgo(SB), A0
	JALR	RA, A0
	RET

// func cgoSigtramp()
TEXT runtime·cgoSigtramp(SB),NOSPLIT,$0
	MOV	$runtime·sigtramp(SB), T1
	JALR	ZERO, T1

// func mmap(addr unsafe.Pointer, n uintptr, prot, flags, fd int32, off uint32) unsafe.Pointer
TEXT runtime·mmap(SB),NOSPLIT,$-8
	MOV	addr+0(FP), A0
	MOV	n+8(FP), A1
	MOVW	prot+16(FP), A2
	MOVW	flags+20(FP), A3
	MOVW	fd+24(FP), A4
	MOVW	off+28(FP), A5
	MOV	$SYS_mmap, A7
	ECALL
	ERR_RETURN_ABS_64(ret+32)
	RET

// func munmap(addr unsafe.Pointer, n uintptr)               {}
TEXT runtime·munmap(SB),NOSPLIT,$-8
	MOV	addr+0(FP), A0
	MOV	n+8(FP), A1
	MOV	$SYS_munmap, A7
	ECALL
	ERR_ABORT
	RET

// func madvise(addr unsafe.Pointer, n uintptr, flags int32) {}
TEXT runtime·madvise(SB),NOSPLIT,$-8
	MOV	addr+0(FP), A0
	MOV	n+8(FP), A1
	MOVW	flags+16(FP), A2
	MOV	$SYS_madvise, A7
	ECALL
	// ignore failure - maybe pages are locked
	RET

// func futex(addr unsafe.Pointer, op int32, val uint32, ts, addr2 unsafe.Pointer, val3 uint32) int32
TEXT runtime·futex(SB),NOSPLIT,$-8
	MOV	addr+0(FP), A0
	MOVW	op+8(FP), A1
	MOVW	val+12(FP), A2
	MOV	ts+16(FP), A3
	MOV	addr2+24(FP), A4
	MOVW	val3+32(FP), A5
	MOV	$SYS_futex, A7
	ECALL
	MOVW	A0, ret+40(FP)
	RET

// func clone(flags int32, stk, mp, gp, fn unsafe.Pointer) int32
TEXT runtime·clone(SB),NOSPLIT,$-8
	MOVW	flags+0(FP), A0
	MOV	stk+8(FP), A1

	// Copy mp, gp, fn off parent stack for use by child.
	MOV	mp+16(FP), T0
	MOV	gp+24(FP), T1
	MOV	fn+32(FP), T2

	MOV	T0, -8(A1)
	MOV	T1, -16(A1)
	MOV	T2, -24(A1)
	MOV	$1234, T0
	MOV	T0, -32(A1)

	MOV	$SYS_clone, A7
	ECALL

	// In parent, return.
	BEQ	ZERO, A0, child
	MOVW	ZERO, ret+40(FP)
	RET
child:

	// In child, on new stack.
	MOV	-32(X2), T0
	MOV	$1234, A0
	BEQ	A0, T0, good
	WORD	$0	//crash
good:

	// Initialize m->procid to Linux tid
	MOV	$SYS_gettid, A7
	ECALL

	MOV	-24(X2), T2     // fn
	MOV	-16(X2), T1     // g
	MOV	-8(X2), T0      // m

	BEQ	ZERO, T0, nog
	BEQ	ZERO, T1, nog

	MOV	A0, m_procid(T0)

	// In child, set up new stack
	MOV	T0, g_m(T1)
	MOV	T1, g

nog:
	// Call fn
	JALR	RA, T2
	WORD	$0

// func sigaltstack(new, old *stackt)
TEXT runtime·sigaltstack(SB),NOSPLIT,$-8
	MOV	new+0(FP), A0
	MOV	old+8(FP), A1
	MOV	$SYS_sigaltstack, A7
	ECALL
	ERR_ABORT
	RET

// func osyield()
TEXT runtime·osyield(SB),NOSPLIT,$-8
	MOV	$SYS_sched_yield, A7
	ECALL
	RET

// func sched_getaffinity(pid, len uintptr, buf *uintptr) int32
TEXT runtime·sched_getaffinity(SB),NOSPLIT,$-8
	MOV	pid+0(FP), A0
	MOV	len+8(FP), A1
	MOV	buf+16(FP), A2
	MOV	$SYS_sched_getaffinity, A7
	ECALL
	MOV	A0, ret+24(FP)
	RET

// func epollcreate(size int32) int32
TEXT runtime·epollcreate(SB),NOSPLIT,$-8
	MOV	$0, A0
	MOV	$SYS_epoll_create1, A7
	ECALL
	MOVW	A0, ret+8(FP)
	RET

// func epollcreate1(flags int32) int32
TEXT runtime·epollcreate1(SB),NOSPLIT,$-8
	MOVW	flags+0(FP), A0
	MOV	$SYS_epoll_create1, A7
	ECALL
	MOVW	A0, ret+8(FP)
	RET

// func epollctl(epfd, op, fd int32, ev *epollevent) int32
TEXT runtime·epollctl(SB),NOSPLIT,$-8
	MOVW	epfd+0(FP), A0
	MOVW	op+4(FP), A1
	MOVW	fd+8(FP), A2
	MOV	ev+16(FP), A3
	MOV	$SYS_epoll_ctl, A7
	ECALL
	MOVW	A0, ret+24(FP)
	RET

// func epollwait(epfd int32, ev *epollevent, nev, timeout int32) int32
TEXT runtime·epollwait(SB),NOSPLIT,$-8
	MOVW	epfd+0(FP), A0
	MOV	ev+8(FP), A1
	MOVW	nev+16(FP), A2
	MOVW	timeout+20(FP), A3
	MOV	$0, A4
	MOV	$SYS_epoll_pwait, A7
	ECALL
	MOVW	A0, ret+24(FP)
	RET

// func closeonexec(int32)                                   {}
TEXT runtime·closeonexec(SB),NOSPLIT,$-8
	MOVW	fd+0(FP), A0  // fd
	MOV	$2, A1	// F_SETFD
	MOV	$1, A2	// FD_CLOEXEC
	MOV	$SYS_fcntl, A7
	ECALL
	RET
