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

import "cmd/internal/obj"

const (
	// Things which the assembler treats as instructions but which do not
	// correspond to actual RISC-V instructions (e.g., the TEXT directive at
	// the start of each symbol, MOVs).
	type_pseudo = iota

	// Integer register-immediate instructions, such as ADDI.
	type_regi_immi

	// Integer register-register instructions, such as ADD.
	type_regi2

	// Instructions which get compiled as jump-and-link, including JMP.
	type_jal

	// Instructions which get compiled as register jump-and-link.
	type_jalr

	// Conditional branches.
	type_branch

	// System instructions (read counters).  These are encoded using a
	// variant of the I-type encoding.
	type_system
)

type Optab struct {
	as    obj.As
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

	{obj.ATEXT, C_MEM, C_IMMI, C_TEXTSIZE, type_pseudo, 0},

	{ABEQ, C_REGI, C_REGI, C_RELADDR, type_branch, 4},
	{ABNE, C_REGI, C_REGI, C_RELADDR, type_branch, 4},
	{ABLT, C_REGI, C_REGI, C_RELADDR, type_branch, 4},
	{ABGE, C_REGI, C_REGI, C_RELADDR, type_branch, 4},
	{ABLTU, C_REGI, C_REGI, C_RELADDR, type_branch, 4},
	{ABGEU, C_REGI, C_REGI, C_RELADDR, type_branch, 4},

	// Note that these are backwards from what one would expect.
	// The link destination register is in src1 because the Go toolchain
	// requires the jump address to be in dest.
	{AJAL, C_REGI, C_NONE, C_RELADDR, type_jal, 4},
	{AJALR, C_REGI, C_NONE, C_MEM, type_jalr, 4},

	{AADDI, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{ASLLI, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{AXORI, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{ASRLI, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{ASRAI, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{AORI, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},
	{AANDI, C_IMMI, C_REGI, C_REGI, type_regi_immi, 4},

	{AADD, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASUB, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASLL, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{AXOR, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASRL, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{ASRA, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{AOR, C_REGI, C_REGI, C_REGI, type_regi2, 4},
	{AAND, C_REGI, C_REGI, C_REGI, type_regi2, 4},

	{ASCALL, C_NONE, C_NONE, C_NONE, type_system, 4},

	{ARDCYCLE, C_NONE, C_NONE, C_REGI, type_system, 4},
	{ARDTIME, C_NONE, C_NONE, C_REGI, type_system, 4},
	{ARDINSTRET, C_NONE, C_NONE, C_REGI, type_system, 4},
}

// progedit is called individually for each Prog.
// TODO(myenik)
func progedit(ctxt *obj.Link, p *obj.Prog) {
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

	// Populate the Class field in the operands.
	aclass(ctxt, &p.From)
	if p.From3 == nil {
		// There is no third operand for this operation.  Create one to
		// make other code have to deal with fewer special cases.
		p.From3 = &obj.Addr{}
		if p.As != AJAL && p.As != AJALR && (p.From.Class == C_REGI || p.From.Class == C_IMMI) {
			p.From3.Reg = p.To.Reg
			p.From3.Class = C_REGI
		} else {
			p.From3.Class = C_NONE
		}
	} else {
		aclass(ctxt, p.From3)
	}
	aclass(ctxt, &p.To)

	// Rewrite pseudoinstructions.
	if p.From.Class == C_IMMI {
		switch p.As {
		case AADD:
			p.As = AADDI
		case AAND:
			p.As = AANDI
		case AOR:
			p.As = AORI
		case ASLL:
			p.As = ASLLI
		case ASRA:
			p.As = ASRAI
		case ASRL:
			p.As = ASRLI
		case AXOR:
			p.As = AXORI
		}
	}
	switch p.As {
	case AMOV:
		switch p.From.Class {
		case C_REGI:
			p.As = AADDI
			*p.From3 = p.From
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = 0
			aclass(ctxt, &p.From)
		case C_IMMI:
			p.As = AADDI
			p.From3.Type = obj.TYPE_REG
			p.From3.Reg = REG_ZERO
			aclass(ctxt, p.From3)
		default:
			ctxt.Diag("progedit: unsupported MOV")
		}
	case obj.AJMP:
		// Convert "JMP label" into "JAL ZERO, label".
		p.As = AJAL
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REG_ZERO
		aclass(ctxt, &p.From)
	}
}

// TODO(myenik)
func follow(ctxt *obj.Link, s *obj.LSym) {
}

// Given an Addr, reads the Addr's high-level Type and converts it to a
// low-level Class.
func aclass(ctxt *obj.Link, a *obj.Addr) {
	switch a.Type {
	case obj.TYPE_NONE:
		a.Class = C_NONE

	case obj.TYPE_REG:
		if REG_X0 <= a.Reg && a.Reg <= REG_X31 {
			a.Class = C_REGI
		} else if REG_F0 <= a.Reg && a.Reg <= REG_F31 {
			ctxt.Diag("aclass: floating-point registers are unsupported")
		} else {
			ctxt.Diag("aclass: unknown register %v", a.Reg)
		}

	case obj.TYPE_CONST:
		a.Class = C_IMMI

	case obj.TYPE_BRANCH:
		a.Class = C_RELADDR

	case obj.TYPE_TEXTSIZE:
		a.Class = C_TEXTSIZE

	case obj.TYPE_MEM:
		a.Class = C_MEM

	default:
		ctxt.Diag("aclass: unsupported type %v", a.Type)
	}
}

// preprocess is responsible for:
// * Updating the SP on function entry and exit
// * Rewriting RET to a real return instruction
func preprocess(ctxt *obj.Link, cursym *obj.LSym) {
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
		switch p.As {
		case obj.ATEXT:
			// Function entry. Setup stack.
			// TODO(prattmic): handle calls to morestack.
			q = p
			q = obj.Appendp(ctxt, q)
			q.As = AADDI
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
			q.As = AADDI
			q.From.Type = obj.TYPE_CONST
			q.From.Offset = stackSize
			q.From3 = &obj.Addr{}
			q.From3.Type = obj.TYPE_REG
			q.From3.Reg = REG_SP
			q.To.Type = obj.TYPE_REG
			q.To.Reg = REG_SP
			q.Spadj = int32(stackSize)
			aclass(ctxt, &q.From)
			aclass(ctxt, q.From3)
			aclass(ctxt, &q.To)

			q = obj.Appendp(ctxt, q)
			q.As = AJALR
			q.From.Type = obj.TYPE_REG
			q.From.Reg = REG_ZERO
			q.From3 = &obj.Addr{}
			q.From3.Class = C_NONE
			q.To.Type = obj.TYPE_MEM
			q.To.Reg = REG_RA
		}
		aclass(ctxt, &q.From)
		aclass(ctxt, q.From3)
		aclass(ctxt, &q.To)
	}
}

// Looks up an operation in the operation table.
func oplook(ctxt *obj.Link, p *obj.Prog) *Optab {
	for i := 0; i < len(optab); i++ {
		o := optab[i]
		if o.as == p.As &&
			o.src1 == p.From.Class &&
			o.src2 == p.From3.Class &&
			o.dest == p.To.Class {
			return &o
		}
	}
	ctxt.Diag("oplook: could not find op %#v (%#v)", p, *p.From3)
	return nil
}

// Encodes a register.
func reg(ctxt *obj.Link, r int16) uint32 {
	if r < REG_X0 || REG_END <= r {
		ctxt.Diag("reg: invalid register %d", r)
	}
	return uint32(r - obj.RBaseRISCV)
}

// Encodes a signed integer immediate.
func immi(ctxt *obj.Link, i int64, nbits uint) uint32 {
	if i < -(1<<(nbits-1)) || (1<<(nbits-1))-1 < i {
		// The immediate will not fit in the bits allotted to it in the
		// instruction.
		ctxt.Diag("immi: too large immediate %d", i)
	}
	return uint32(i)
}

// Encodes an R-type instruction.
func instr_r(ctxt *obj.Link, funct7 uint32, rs2 int16, rs1 int16, funct3 uint32, rd int16, opcode uint32) uint32 {
	if funct7>>7 != 0 {
		ctxt.Diag("instr_r: too large funct7 %#x", funct7)
	}
	if funct3>>3 != 0 {
		ctxt.Diag("instr_r: too large funct3 %#x", funct3)
	}
	if opcode>>7 != 0 {
		ctxt.Diag("instr_r: too large opcode %#x", opcode)
	}
	return funct7<<25 | reg(ctxt, rs2)<<20 | reg(ctxt, rs1)<<15 | funct3<<12 | reg(ctxt, rd)<<7 | opcode
}

// Encodes an I-type instruction.
func instr_i(ctxt *obj.Link, imm int64, rs1 int16, funct3 uint32, rd int16, opcode uint32) uint32 {
	if funct3>>3 != 0 {
		ctxt.Diag("instr_i: too large funct3 %#x", funct3)
	}
	if opcode>>7 != 0 {
		ctxt.Diag("instr_i: too large opcode %#x", opcode)
	}
	return immi(ctxt, imm, 12)<<20 | reg(ctxt, rs1)<<15 | funct3<<12 | reg(ctxt, rd)<<7 | opcode
}

// Encodes an SB-type instruction.
func instr_sb(ctxt *obj.Link, imm64 int64, rs2 int16, rs1 int16, funct3 uint32, opcode uint32) uint32 {
	if funct3>>3 != 0 {
		ctxt.Diag("instr_sb: too large funct3 %#x", funct3)
	}
	if opcode>>7 != 0 {
		ctxt.Diag("instr_sb: too large opcode %#x", opcode)
	}
	imm := immi(ctxt, imm64, 13)
	return (imm>>12)<<31 |
		((imm>>5)&0x3f)<<25 |
		reg(ctxt, rs2)<<20 |
		reg(ctxt, rs1)<<15 |
		funct3<<12 |
		((imm>>1)&0xf)<<8 |
		((imm>>11)&0x1)<<7 |
		opcode
}

// Encodes a UJ-type instruction.
func instr_uj(ctxt *obj.Link, imm64 int64, rd int16, opcode uint32) uint32 {
	if opcode>>7 != 0 {
		ctxt.Diag("instr_i: too large opcode %#x", opcode)
	}
	imm := immi(ctxt, imm64, 21)
	return (imm>>20)<<31 |
		((imm>>1)&0x3ff)<<21 |
		((imm>>11)&0x1)<<20 |
		((imm>>12)&0xff)<<12 |
		reg(ctxt, rd)<<7 |
		opcode
}

// Encodes a machine instruction.
func asmout(ctxt *obj.Link, p *obj.Prog, o *Optab) uint32 {
	result := uint32(0)
	switch o.type_ {
	default:
		ctxt.Diag("unknown type %d", o.type_)
	case type_pseudo:
		ctxt.Diag("asmout: found pseudoinstruction %v", o.as)
	case type_regi_immi:
		encoded := encode(o.as)
		// TODO(bbaren): Do something reasonable if immediate is too large.
		result = instr_i(ctxt, p.From.Offset, p.From3.Reg, encoded.funct3, p.To.Reg, encoded.opcode)
	case type_regi2:
		encoded := encode(o.as)
		result = instr_r(ctxt, encoded.funct7, p.From.Reg, p.From3.Reg, encoded.funct3, p.To.Reg, encoded.opcode)
	case type_branch:
		encoded := encode(o.as)
		// Compute the branch offset.  We couldn't do this in progedit,
		// because the offset is relative and we didn't know what the
		// code would look like laid out.
		offset := p.Pcond.Pc - p.Pc
		// TODO(bbaren): Do something reasonable if offset is too large.
		if offset%4 != 0 {
			ctxt.Diag("asmout: misaligned branch offset %d", offset)
		}
		result = instr_sb(ctxt, offset, p.Reg, p.From.Reg, encoded.funct3, encoded.opcode)
	case type_jal:
		// Compute the jump offset.  We couldn't do this in progedit,
		// because the offset is relative and we didn't know what the
		// code would look like laid out.
		offset := p.Pcond.Pc - p.Pc
		// TODO(bbaren): Do something reasonable if immediate is too large.
		if offset%4 != 0 {
			ctxt.Diag("asmout: misaligned jump offset %d", offset)
		}
		result = instr_uj(ctxt, offset, p.From.Reg, encode(o.as).opcode)
	case type_jalr:
		encoded := encode(o.as)
		if p.From.Name != obj.NAME_NONE {
			ctxt.Diag("asmout: unsupported symbol in addr: %#v", p)
		}
		if p.From.Scale != 0 {
			ctxt.Diag("asmout: unsupported scale in addr: %#v", p)
		}
		result = instr_i(ctxt, p.To.Offset, p.To.Reg, encoded.funct3, p.From.Reg, encoded.opcode)
	case type_system:
		encoded := encode(o.as)
		switch p.To.Class {
		case C_REGI:
			result = instr_i(ctxt, encoded.csr, REG_ZERO, encoded.funct3, p.To.Reg, encoded.opcode)
		case C_NONE:
			result = instr_i(ctxt, encoded.csr, REG_ZERO, encoded.funct3, REG_ZERO, encoded.opcode)
		default:
			ctxt.Diag("unknown instruction %d", o.as)
		}
	}
	return result
}

func assemble(ctxt *obj.Link, cursym *obj.LSym) {
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
	cursym.Grow(cursym.Size)

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
