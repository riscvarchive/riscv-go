// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

//
// System calls and other sys.stuff for riscv, Linux
//

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

TEXT runtime·exit(SB),NOSPLIT,$-8-4
	RET

TEXT runtime·exit1(SB),NOSPLIT,$-8-4
	RET

TEXT runtime·open(SB),NOSPLIT,$-8-20
	RET

TEXT runtime·closefd(SB),NOSPLIT,$-8-12
	RET

TEXT runtime·write(SB),NOSPLIT,$-8-28
	RET

TEXT runtime·read(SB),NOSPLIT,$-8-28
	RET

TEXT runtime·getrlimit(SB),NOSPLIT,$-8-20
	RET

TEXT runtime·usleep(SB),NOSPLIT,$16-4
	RET

TEXT runtime·gettid(SB),NOSPLIT,$0-4
	RET

TEXT runtime·raise(SB),NOSPLIT,$-8
	RET

TEXT runtime·raiseproc(SB),NOSPLIT,$-8
	RET

TEXT runtime·setitimer(SB),NOSPLIT,$-8-24
	RET

TEXT runtime·mincore(SB),NOSPLIT,$-8-28
	RET

// func walltime() (sec int64, nsec int32)
TEXT runtime·walltime(SB),NOSPLIT,$16
	RET

// func now() (sec int64, nsec int32)
TEXT runtime·nanotime(SB),NOSPLIT,$16
	RET

TEXT runtime·rtsigprocmask(SB),NOSPLIT,$-8-28
	RET

TEXT runtime·rt_sigaction(SB),NOSPLIT,$-8-36
	RET

TEXT runtime·sigfwd(SB),NOSPLIT,$0-32
	RET

TEXT runtime·sigtramp(SB),NOSPLIT,$64
	RET

TEXT runtime·cgoSigtramp(SB),NOSPLIT,$0
	JMP	runtime·sigtramp(SB)

TEXT runtime·mmap(SB),NOSPLIT,$-8
	RET

TEXT runtime·munmap(SB),NOSPLIT,$-8
	RET

TEXT runtime·madvise(SB),NOSPLIT,$-8
	RET

// int64 futex(int32 *uaddr, int32 op, int32 val,
//	struct timespec *timeout, int32 *uaddr2, int32 val2);
TEXT runtime·futex(SB),NOSPLIT,$-8
	RET

// int64 clone(int32 flags, void *stk, M *mp, G *gp, void (*fn)(void));
TEXT runtime·clone(SB),NOSPLIT,$-8
	RET

TEXT runtime·sigaltstack(SB),NOSPLIT,$-8
	RET

TEXT runtime·osyield(SB),NOSPLIT,$-8
	RET

TEXT runtime·sched_getaffinity(SB),NOSPLIT,$-8
	RET

// int32 runtime·epollcreate(int32 size);
TEXT runtime·epollcreate(SB),NOSPLIT,$-8
	RET

// int32 runtime·epollcreate1(int32 flags);
TEXT runtime·epollcreate1(SB),NOSPLIT,$-8
	RET

// func epollctl(epfd, op, fd int32, ev *epollEvent) int
TEXT runtime·epollctl(SB),NOSPLIT,$-8
	RET

// int32 runtime·epollwait(int32 epfd, EpollEvent *ev, int32 nev, int32 timeout);
TEXT runtime·epollwait(SB),NOSPLIT,$-8
	RET

// void runtime·closeonexec(int32 fd);
TEXT runtime·closeonexec(SB),NOSPLIT,$-8
	RET
