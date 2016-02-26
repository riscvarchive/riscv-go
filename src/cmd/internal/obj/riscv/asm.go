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

	// Integer register-immediate instructions, such as ADDI.
	type_regi_immi

	// Integer register-register instructions, such as ADD.
	type_regi2

	// Instructions which get compiled as jump-and-link, including JMP.
	type_jal

	// System instructions (read counters).  These are encoded using a
	// variant of the I-type encoding.
	type_system
)

type Optab struct {
	as    int16
	src1  int8
	src2  int8
	dest  int8
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

	{AADD, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{AADD, C_REGI, C_NONE, C_REGI, type_regi2, 4},
	{AADD, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{AADD, C_IMMI, C_NONE, C_REGI, type_regi_immi, 4},

	{ASUB, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASUB, C_REGI, C_NONE, C_REGI, type_regi2, 4},

	{ASLL, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASLL, C_REGI, C_NONE, C_REGI, type_regi2, 4},
	{ASLL, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{ASLL, C_IMMI, C_NONE, C_REGI, type_regi_immi, 4},
	{ASRL, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASRL, C_REGI, C_NONE, C_REGI, type_regi2, 4},
	{ASRL, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{ASRL, C_IMMI, C_NONE, C_REGI, type_regi_immi, 4},
	{ASRA, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASRA, C_REGI, C_NONE, C_REGI, type_regi2, 4},
	{ASRA, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{ASRA, C_IMMI, C_NONE, C_REGI, type_regi_immi, 4},

	{AAND, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{AAND, C_REGI, C_NONE, C_REGI, type_regi2, 4},
	{AAND, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{AAND, C_IMMI, C_NONE, C_REGI, type_regi_immi, 4},
	{AOR, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{AOR, C_REGI, C_NONE, C_REGI, type_regi2, 4},
	{AOR, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{AOR, C_IMMI, C_NONE, C_REGI, type_regi_immi, 4},
	{AXOR, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{AXOR, C_REGI, C_NONE, C_REGI, type_regi2, 4},
	{AXOR, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{AXOR, C_IMMI, C_NONE, C_REGI, type_regi_immi, 4},

	{obj.AJMP, C_NONE, C_NONE, C_RELADDR, type_jal, 4},

	{ARDCYCLE, C_NONE, C_NONE, C_REGI, type_system, 4},
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

// Given an Addr, reads the Addr's high-level Type and converts it to a
// low-level Class.
func aclass(a *obj.Addr) {
	switch a.Type {
	case obj.TYPE_NONE:
		a.Class = C_NONE

	case obj.TYPE_REG:
		if REG_X0 <= a.Reg && a.Reg <= REG_X31 {
			a.Class = C_REGI
		} else if REG_F0 <= a.Reg && a.Reg <= REG_F31 {
			log.Printf("aclass: floating-point registers are unsupported")
		}

	case obj.TYPE_CONST:
		a.Class = C_IMMI

	case obj.TYPE_BRANCH:
		a.Class = C_RELADDR

	default:
		log.Printf("aclass: unsupported type")
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
			q.As = AADD
			q.From.Type = obj.TYPE_CONST
			q.From.Offset = -stackSize
			q.From3 = &obj.Addr{}
			q.From3.Type = obj.TYPE_REG
			q.From3.Reg = REG_SP
			q.To.Type = obj.TYPE_REG
			q.To.Reg = REG_SP
			q.Spadj = int32(-stackSize)
		case obj.ARET:
			// Function exit. Stack teardown and exit.
			q = p
			q = obj.Appendp(ctxt, q)
			q.As = AADD
			q.From.Type = obj.TYPE_CONST
			q.From.Offset = stackSize
			q.From3 = &obj.Addr{}
			q.From3.Type = obj.TYPE_REG
			q.From3.Reg = REG_SP
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

	// Normalize all the instructions.
	for p := cursym.Text; p != nil; p = p.Link {
		// Populate the Class field in the operands.
		aclass(&p.From)
		if p.From3 == nil {
			// There is no third operand for this operation.  Create an
			// empty one to make other code have to deal with fewer special
			// cases.
			p.From3 = &obj.Addr{}
			p.From3.Class = C_NONE
		} else {
			aclass(p.From3)
		}
		aclass(&p.To)
	}
}

// Looks up an operation in the operation table.
func oplook(ctxt *obj.Link, p *obj.Prog) *Optab {
	log.Printf("oplook: ctxt: %+v p: %+v", ctxt, p)

	for i := 0; i < len(optab); i++ {
		o := optab[i]
		if o.as == p.As &&
			o.src1 == p.From.Class &&
			o.src2 == p.From3.Class &&
			o.dest == p.To.Class {
			return &o
		}
	}

	// No instruction found, so just return a placeholder NOP.
	log.Printf("oplook: could not find op, returning NOP")
	return &optab[0]
}

// Encodes a register.
func reg(r int16) uint32 {
	if r < REG_X0 || REG_END <= r {
		log.Fatalf("reg: invalid register %d", r)
	}
	return uint32(r - obj.RBaseRISCV)
}

// Encodes a signed integer immediate.
func immi(i int64, nbits uint) uint32 {
	if i < -(1<<(nbits-1)) || (1<<(nbits-1))-1 < i {
		// The immediate will not fit in the bits allotted to it in the
		// instruction.
		log.Fatalf("immi: too large immediate %d", i)
	}
	return uint32(i)
}

// Encodes an R-type instruction.
func instr_r(funct7 uint32, rs2 int16, rs1 int16, funct3 uint32, rd int16, opcode uint32) uint32 {
	if funct7>>7 != 0 {
		log.Fatalf("instr_r: too large funct7 %#x", funct7)
	}
	if funct3>>3 != 0 {
		log.Fatalf("instr_r: too large funct3 %#x", funct3)
	}
	if opcode>>7 != 0 {
		log.Fatalf("instr_r: too large opcode %#x", opcode)
	}
	return funct7<<25 | reg(rs2)<<20 | reg(rs1)<<15 | funct3<<12 | reg(rd)<<7 | opcode
}

// Encodes an I-type instruction.
func instr_i(imm int64, rs1 int16, funct3 uint32, rd int16, opcode uint32) uint32 {
	if funct3>>3 != 0 {
		log.Fatalf("instr_i: too large funct3 %#x", funct3)
	}
	if opcode>>7 != 0 {
		log.Fatalf("instr_i: too large opcode %#x", opcode)
	}
	return immi(imm, 12)<<20 | reg(rs1)<<15 | funct3<<12 | reg(rd)<<7 | opcode
}

// Encodes a UJ-type instruction.
func instr_uj(imm64 int64, rd int16, opcode uint32) uint32 {
	if opcode>>7 != 0 {
		log.Fatalf("instr_i: too large opcode %#x", opcode)
	}
	imm := immi(imm64, 21)
	return (imm>>20)<<31 |
		((imm>>1)&0x3ff)<<21 |
		((imm>>11)&0x1)<<20 |
		((imm>>12)&0xff)<<12 |
		reg(rd)<<7 |
		opcode
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
	case type_regi_immi:
		var encoded *inst
		switch o.as {
		case AADD:
			encoded = encode(AADDI)
		case AAND:
			encoded = encode(AANDI)
		case AOR:
			encoded = encode(AORI)
		case ASLL:
			encoded = encode(ASLLI)
		case ASRA:
			encoded = encode(ASRAI)
		case ASRL:
			encoded = encode(ASRLI)
		case AXOR:
			encoded = encode(AXORI)
		default:
			ctxt.Diag("unknown instruction %d", o.as)
		}
		if p.From3.Class == C_NONE {
			p.From3.Reg = p.To.Reg
		}
		// TODO(bbaren): Do something reasonable if immediate is too large.
		result = instr_i(p.From.Offset, p.From3.Reg, encoded.funct3, p.To.Reg, encoded.opcode)
	case type_regi2:
		if p.From3.Class == C_NONE {
			p.From3.Reg = p.To.Reg
		}
		encoded := encode(o.as)
		result = instr_r(encoded.funct7, p.From.Reg, p.From3.Reg, encoded.funct3, p.To.Reg, encoded.opcode)
	case type_jal:
		var encoded *inst
		var rd int16
		switch o.as {
		case obj.AJMP:
			encoded = encode(AJAL)
			rd = REG_ZERO
		default:
			ctxt.Diag("unknown instruction %d", o.as)
		}
		var offset int64
		if p.Pcond == nil {
			offset = 0
		} else {
			offset = p.Pcond.Pc - p.Pc
		}
		// TODO(bbaren): Do something reasonable if immediate is too large.
		if offset%4 != 0 {
			ctxt.Diag("asmout: misaligned jump offset %d", offset)
		}
		result = instr_uj(offset, rd, encoded.opcode)
	case type_system:
		encoded := encode(o.as)
		result = instr_i(int64(encoded.csr), REG_ZERO, encoded.funct3, p.To.Reg, encoded.opcode)
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
