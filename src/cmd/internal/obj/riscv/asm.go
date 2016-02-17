// Copyright © 2015 The Go Authors.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package riscv

import (
	"log"

	"cmd/internal/obj"
)

const (
	// Things which the assembler treats as instructions but which do not
	// correspond to actual RISC-V instructions (e.g., the TEXT directive at
	// the start of each symbol).
	type_pseudo = iota
)

type Optab struct {
	as    int16
	op1   uint8
	op2   uint8
	op3   uint8
	type_ int8 // internal instruction type used to dispatch in asmout
	size  int8 // bytes
}

var optab = []Optab{
	// This is a Go (liblink) NOP, not a RISC-V NOP; it's only used to make
	// the assembler happy with otherwise empty symbols.  It thus occupies
	// zero bytes.  (RISC-V NOPs are not currently supported.)
	//
	// TODO(bbaren, mpratt): Can we strip these out in progedit or
	// preprocess?
	{obj.ANOP, C_NONE, C_NONE, C_NONE, type_pseudo, 0},
}

// progedit is called individually for each Prog.
// TODO(myenik)
func progedit(ctxt *obj.Link, p *obj.Prog) {
	log.Printf("progedit: ctxt: %+v p: %#v p: %s", ctxt, p, p)

	// Rewrite branches as TYPE_BRANCH
	switch p.As {
	case AJAL,
		AJALR,
		ABEQ,
		ABNE,
		ABLT,
		ABLTU,
		ABGE,
		ABGEU,
		obj.ARET,
		obj.ADUFFZERO,
		obj.ADUFFCOPY:
		if p.To.Sym != nil {
			p.To.Type = obj.TYPE_BRANCH
		}
	}
}

// TODO(myenik)
func follow(ctxt *obj.Link, s *obj.LSym) {
	log.Printf("follow: ctxt: %+v", ctxt)

	for ; s != nil; s = s.Next {
		log.Printf("s: %+v", s)
	}
}

// preprocess is responsible for:
// * Updating the SP on function entry and exit
// * Rewriting RET to a real return instruction
func preprocess(ctxt *obj.Link, cursym *obj.LSym) {
	log.Printf("preprocess: ctxt: %+v", ctxt)

	ctxt.Cursym = cursym

	if cursym.Text == nil || cursym.Text.Link == nil {
		return
	}

	stackSize := cursym.Text.To.Offset

	// TODO(prattmic): explain what these are really for,
	// once I figure it out.
	cursym.Args = cursym.Text.To.Val.(int32)
	cursym.Locals = int32(stackSize)

	var q *obj.Prog
	for p := cursym.Text; p != nil; p = p.Link {
		log.Printf("p: %+v", p)

		switch p.As {
		case obj.ATEXT:
			// Function entry. Setup stack.
			// TODO(prattmic): handle calls to morestack.
			q = p
			q = obj.Appendp(ctxt, q)
			q.As = AADDI
			q.From.Type = obj.TYPE_REG
			q.From.Reg = REG_SP
			q.From3 = &obj.Addr{}
			q.From3.Type = obj.TYPE_CONST
			q.From3.Offset = -stackSize
			q.To.Type = obj.TYPE_REG
			q.To.Reg = REG_SP
			q.Spadj = int32(-stackSize)
		case obj.ARET:
			// Function exit. Stack teardown and exit.
			q = p
			q = obj.Appendp(ctxt, q)
			q.As = AADDI
			q.From.Type = obj.TYPE_REG
			q.From.Reg = REG_SP
			q.From3 = &obj.Addr{}
			q.From3.Type = obj.TYPE_CONST
			q.From3.Offset = stackSize
			q.To.Type = obj.TYPE_REG
			q.To.Reg = REG_SP
			q.Spadj = int32(stackSize)

			q = obj.Appendp(ctxt, q)
			q.As = AJAL
			q.From.Type = obj.TYPE_REG
			q.From.Reg = REG_RA
			q.To.Type = obj.TYPE_REG
			q.To.Reg = REG_ZERO
		}
	}
}

// Looks up an operation in the operation table.
func oplook(ctxt *obj.Link, p *obj.Prog) *Optab {
	log.Printf("oplook: ctxt: %+v p: %+v", ctxt, p)
	return &optab[0] // Just make everything a NOP.
}

// Encodes a machine instruction.
func asmout(ctxt *obj.Link, p *obj.Prog, o *Optab) uint32 {
	log.Printf("asmout: ctxt: %+v p: %+v o: %+v", ctxt, p, o)

	result := uint32(0)
	switch o.type_ {
	default:
		ctxt.Diag("unknown type %d", o.type_)
	case type_pseudo:
		break
	}
	return result
}

func assemble(ctxt *obj.Link, cursym *obj.LSym) {
	log.Printf("assemble: ctxt: %+v", ctxt)

	if cursym.Text == nil || cursym.Text.Link == nil {
		// We're being asked to assemble an external function or an ELF
		// section symbol.  Do nothing.
		return
	}

	ctxt.Cursym = cursym

	// Determine how many bytes this symbol will wind up using.
	pc := int64(0) // program counter relative to the start of the symbol
	ctxt.Autosize = int32(cursym.Text.To.Offset + 4)
	for p := cursym.Text; p != nil; p = p.Link {
		ctxt.Curp = p
		ctxt.Pc = pc
		p.Pc = pc

		m := oplook(ctxt, p).size

		// All operations should be 32 bits wide.
		if m%4 != 0 || p.Pc%4 != 0 {
			ctxt.Diag("!pc invalid: %v size=%d", p, m)
		}

		if m == 0 {
			// TODO(bbaren): Once everything's all done, do something like
			//   if not a nop {
			//     bail out
			//   }
			continue
		}

		pc += int64(m)
	}
	cursym.Size = pc // remember the size of this symbol

	// Allocate for the symbol.
	obj.Symgrow(ctxt, cursym, cursym.Size)

	// Lay out code.
	bp := cursym.P
	for p := cursym.Text; p != nil; p = p.Link {
		ctxt.Curp = p
		ctxt.Pc = p.Pc

		o := oplook(ctxt, p)
		if o.size != 0 {
			ctxt.Arch.ByteOrder.PutUint32(bp, asmout(ctxt, p, o))
			bp = bp[4:]
		}
	}
}
