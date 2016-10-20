// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj/riscv"
)

func Main() {
	gc.Thearch.LinkArch = &riscv.LinkRISCV

	gc.Thearch.REGSP = riscv.REG_SP
	// TODO(prattmic): all the other arches use 50 bits, even though
	// they have 48-bit vaddrs. why?
	gc.Thearch.MAXWIDTH = 1 << 50

	gc.Thearch.Defframe = defframe
	gc.Thearch.Proginfo = proginfo

	// TODO(prattmic): other fields?

	gc.Thearch.SSAMarkMoves = ssaMarkMoves
	gc.Thearch.SSAGenValue = ssaGenValue
	gc.Thearch.SSAGenBlock = ssaGenBlock

	gc.Main()
	gc.Exit(0)
}
