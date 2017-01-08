// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

// If !iscgo, this is a no-op.
TEXT runtime·save_g(SB),NOSPLIT,$-8-0
	RET

TEXT runtime·load_g(SB),NOSPLIT,$-8-0
	RET

GLOBL runtime·tls_g(SB), TLSBSS, $8
