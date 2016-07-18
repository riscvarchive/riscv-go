// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/riscv"
)

// gins generates one instruction.
func gins(as obj.As, from *gc.Node, to *gc.Node) *obj.Prog {
	p := gc.Prog(as)
	gc.Naddr(&p.From, from)
	gc.Naddr(&p.To, to)
	return p
}

func ginsnop() {
	// Hardware nop is ADD $0, x0
	p := gc.Prog(riscv.AADD)
	p.To.Type = obj.TYPE_REG
	p.To.Reg = riscv.REG_X0
}
