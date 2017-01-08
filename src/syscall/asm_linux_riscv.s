// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

#include "textflag.h"

//
// System calls for riscv, Linux
//

// func Syscall(trap int64, a1, a2, a3 int64) (r1, r2, err int64);

TEXT ·Syscall(SB),NOSPLIT,$0-56
	RET

TEXT ·Syscall6(SB),NOSPLIT,$0-80
	RET

TEXT ·RawSyscall(SB),NOSPLIT,$0-56
	RET

TEXT ·RawSyscall6(SB),NOSPLIT,$0-80
	RET
