// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/riscv"
)

func defframe(ptxt *obj.Prog) {
	// fill in argument size, stack size
	ptxt.To.Type = obj.TYPE_TEXTSIZE

	ptxt.To.Val = int32(gc.Rnd(gc.Curfn.Type.ArgWidth(), int64(gc.Widthptr)))
	frame := uint32(gc.Rnd(gc.Stksize+gc.Maxarg, int64(gc.Widthreg)))

	ptxt.To.Offset = int64(frame)

	// insert code to zero ambiguously live variables
	// so that the garbage collector only sees initialized values
	// when it looks for pointers.
	p := ptxt

	hi := int64(0)
	lo := hi

	// iterate through declarations - they are sorted in decreasing xoffset order.
	for _, n := range gc.Curfn.Func.Dcl {
		if !n.Name.Needzero {
			continue
		}
		if n.Class != gc.PAUTO {
			gc.Fatalf("needzero class %d", n.Class)
		}
		if n.Type.Width%int64(gc.Widthptr) != 0 || n.Xoffset%int64(gc.Widthptr) != 0 || n.Type.Width == 0 {
			gc.Fatalf("var %v has size %d offset %d", n, int(n.Type.Width), int(n.Xoffset))
		}

		if lo != hi && n.Xoffset+n.Type.Width >= lo-int64(2*gc.Widthreg) {
			// merge with range we already have
			lo = n.Xoffset
			continue
		}

		// zero old range
		p = zerorange(p, int64(frame), lo, hi)

		// set new range
		hi = n.Xoffset + n.Type.Width
		lo = n.Xoffset
	}

	// zero final range
	zerorange(p, int64(frame), lo, hi)
}

// FIXME: This is incredibly inefficient, but nice and simple. Optimize.
// See other zerorange implementations for ideas.
func zerorange(p *obj.Prog, frame int64, lo int64, hi int64) *obj.Prog {
	cnt := hi - lo
	if cnt == 0 {
		return p
	}
	// Loop, zeroing one byte at a time.
	// ADD	$(frame+lo), SP, T0
	// ADD	$(cnt), T0, T1
	// loop:
	// 	MOVB	ZERO, (T0)
	// 	ADD	$1, T0
	//	BNE	T0, T1, loop

	// lo is an offset relative to the frame pointer, which we can't use from this function,
	// but adding the true frame size makes it into an offset from the stack pointer.  frame
	// is the local variable size, get the true frame pointer by adding the size of the saved
	// return address.
	p = appendpp(p, riscv.AADD,
		obj.Addr{Type: obj.TYPE_CONST, Offset: int64(gc.Widthptr) + frame + lo},
		&obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_SP},
		obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_T0},
		0)
	p = appendpp(p, riscv.AADD,
		obj.Addr{Type: obj.TYPE_CONST, Offset: cnt},
		&obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_T0},
		obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_T1},
		0)
	p = appendpp(p, riscv.AMOVB,
		obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_ZERO},
		nil,
		obj.Addr{Type: obj.TYPE_MEM, Reg: riscv.REG_T0},
		0)
	loop := p
	p = appendpp(p, riscv.AADD,
		obj.Addr{Type: obj.TYPE_CONST, Offset: 1},
		nil,
		obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_T0},
		0)
	p = appendpp(p, riscv.ABNE,
		obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_T0},
		nil,
		obj.Addr{Type: obj.TYPE_BRANCH},
		riscv.REG_T1)
	gc.Patch(p, loop)
	return p
}

func appendpp(p *obj.Prog, as obj.As, from obj.Addr, from3 *obj.Addr, to obj.Addr, reg int16) *obj.Prog {
	q := gc.Ctxt.NewProg()
	q.As = as
	q.Pos = p.Pos
	q.From = from
	q.From3 = from3
	q.To = to
	q.Reg = reg
	q.Link = p.Link
	p.Link = q
	return q
}
