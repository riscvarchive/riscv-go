// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// System calls and other sys.stuff for mips64, Linux
//

#include "textflag.h"

// func exit(code int32)
TEXT runtime·exit(SB),NOSPLIT,$-8-4
	WORD $0

// func exit1(code int32)
TEXT runtime·exit1(SB),NOSPLIT,$-8-4
	WORD $0

// func open(name *byte, mode, perm int32) int32
TEXT runtime·open(SB),NOSPLIT,$-8-20
	WORD $0

// func closefd(fd int32) int32
TEXT runtime·closefd(SB),NOSPLIT,$-8-12
	WORD $0

// func write(fd uintptr, p unsafe.Pointer, n int32) int32
TEXT runtime·write(SB),NOSPLIT,$-8-28
	WORD $0

// func read(fd int32, p unsafe.Pointer, n int32) int32
TEXT runtime·read(SB),NOSPLIT,$-8-28
	WORD $0

// func getrlimit(kind int32, limit unsafe.Pointer) int32
TEXT runtime·getrlimit(SB),NOSPLIT,$-8-20
	WORD $0

// func usleep(usec uint32)
TEXT runtime·usleep(SB),NOSPLIT,$16-4
	WORD $0

// func gettid() uint32
TEXT runtime·gettid(SB),NOSPLIT,$0-4
	WORD $0

// func raise(sig uint32)
TEXT runtime·raise(SB),NOSPLIT,$-8
	WORD $0

// func raiseproc(sig uint32)
TEXT runtime·raiseproc(SB),NOSPLIT,$-8
	WORD $0

// func setitimer(mode int32, new, old *itimerval)
TEXT runtime·setitimer(SB),NOSPLIT,$-8-24
	WORD $0

// func mincore(addr unsafe.Pointer, n uintptr, dst *byte) int32
TEXT runtime·mincore(SB),NOSPLIT,$-8-28
	WORD $0

// func walltime() (sec int64, nsec int32)
TEXT runtime·walltime(SB),NOSPLIT,$16
	WORD $0

// func nanotime() int64
TEXT runtime·nanotime(SB),NOSPLIT,$16
	WORD $0

// func rtsigprocmask(how int32, new, old *sigset, size int32)
TEXT runtime·rtsigprocmask(SB),NOSPLIT,$-8-28
	WORD $0

// func rt_sigaction(sig uintptr, new, old *sigactiont, size uintptr) int32
TEXT runtime·rt_sigaction(SB),NOSPLIT,$-8-36
	WORD $0

// func sigfwd(fn uintptr, sig uint32, info *siginfo, ctx unsafe.Pointer)
TEXT runtime·sigfwd(SB),NOSPLIT,$0-32
	WORD $0

// func sigtramp(ureg, note unsafe.Pointer)
TEXT runtime·sigtramp(SB),NOSPLIT,$64
	WORD $0

// func cgoSigtramp()
TEXT runtime·cgoSigtramp(SB),NOSPLIT,$0
	WORD $0

// func mmap(addr unsafe.Pointer, n uintptr, prot, flags, fd int32, off uint32) unsafe.Pointer
TEXT runtime·mmap(SB),NOSPLIT,$-8
	WORD $0

// func munmap(addr unsafe.Pointer, n uintptr)               {}
TEXT runtime·munmap(SB),NOSPLIT,$-8
	WORD $0

// func madvise(addr unsafe.Pointer, n uintptr, flags int32) {}
TEXT runtime·madvise(SB),NOSPLIT,$-8
	WORD $0

// func futex(addr unsafe.Pointer, op int32, val uint32, ts, addr2 unsafe.Pointer, val3 uint32) int32
TEXT runtime·futex(SB),NOSPLIT,$-8
	WORD $0

// func clone(flags int32, stk, mp, gp, fn unsafe.Pointer) int32
TEXT runtime·clone(SB),NOSPLIT,$-8
	WORD $0

// func sigaltstack(new, old *stackt)
TEXT runtime·sigaltstack(SB),NOSPLIT,$-8
	WORD $0

// func osyield()
TEXT runtime·osyield(SB),NOSPLIT,$-8
	WORD $0

// func sched_getaffinity(pid, len uintptr, buf *uintptr) int32
TEXT runtime·sched_getaffinity(SB),NOSPLIT,$-8
	WORD $0

// func epollcreate(size int32) int32
TEXT runtime·epollcreate(SB),NOSPLIT,$-8
	WORD $0

// func epollcreate1(flags int32) int32
TEXT runtime·epollcreate1(SB),NOSPLIT,$-8
	WORD $0

// func epollctl(epfd, op, fd int32, ev *epollevent) int32
TEXT runtime·epollctl(SB),NOSPLIT,$-8
	WORD $0

// func epollwait(epfd int32, ev *epollevent, nev, timeout int32) int32
TEXT runtime·epollwait(SB),NOSPLIT,$-8
	WORD $0

// func closeonexec(int32)                                   {}
TEXT runtime·closeonexec(SB),NOSPLIT,$-8
	WORD $0
