// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

#include "textflag.h"

TEXT _rt0_riscv_linux(SB),NOSPLIT,$0
	JMP	main(SB)

TEXT main(SB),NOSPLIT,$-8
	MOV	$runtime·rt0_go(SB), T0
	JMP	(T0)
