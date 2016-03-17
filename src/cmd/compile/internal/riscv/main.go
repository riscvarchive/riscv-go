// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj/riscv"
)

func betypeinit() {
	gc.Widthptr = 8
	gc.Widthint = 8
	gc.Widthreg = 8
}

func Main() {
	gc.Thearch.Thechar = 'V'
	gc.Thearch.Thestring = "riscv"
	gc.Thearch.Thelinkarch = &riscv.LinkRISCV
	gc.Thearch.Betypeinit = betypeinit

	gc.Thearch.REGSP = riscv.REG_SP
	gc.Thearch.REGCTXT = riscv.REG_CTXT
	gc.Thearch.REGMIN = riscv.REG_X0
	gc.Thearch.REGMAX = riscv.REG_X31
	gc.Thearch.FREGMIN = riscv.REG_F0
	gc.Thearch.FREGMAX = riscv.REG_F31
	// TODO(prattmic): all the other arches use 50 bits, even though
	// they have 48-bit vaddrs. why?
	gc.Thearch.MAXWIDTH = 1 << 50

	gc.Thearch.Gins = gins

	// TODO(prattmic): other fields?

	gc.Thearch.SSARegToReg = ssaRegToReg
	gc.Thearch.SSAMarkMoves = ssaMarkMoves
	gc.Thearch.SSAGenValue = ssaGenValue
	gc.Thearch.SSAGenBlock = ssaGenBlock

	gc.Main()
	gc.Exit(0)
}
