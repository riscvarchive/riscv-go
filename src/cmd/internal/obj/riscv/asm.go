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

// Like all Go assemblers, this assembler proceeds in four steps: progedit,
// follow, preprocess, and assemble.

package riscv

import (
	"fmt"

	"cmd/internal/obj"
)

// err is the type of errors which intermediate assembler steps can produce.
type err string

func (e err) Error() string {
	return string(e)
}

// progedit is called individually for each Prog.  It normalizes instruction
// formats and eliminates as many pseudoinstructions as it can.
func progedit(ctxt *obj.Link, p *obj.Prog) {
	// Ensure everything has a From3 to eliminate a ton of nil-pointer
	// checks later.
	if p.From3 == nil {
		p.From3 = &obj.Addr{Type: obj.TYPE_NONE}
	}

	// Expand binary instructions to ternary ones.
	if p.From3.Type == obj.TYPE_NONE {
		switch p.As {
		case AADD, ASUB, ASLL, AXOR, ASRL, ASRA, AOR, AAND:
			p.From3.Type = obj.TYPE_REG
			p.From3.Reg = p.To.Reg
		}
	}

	// Rewrite instructions with constant operands to refer to the immediate
	// form of the instruction.
	if p.From.Type == obj.TYPE_CONST {
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

	// Do additional single-instruction rewriting.
	switch p.As {
	// Turn JMP into JAL ZERO.
	case obj.AJMP:
		p.As = AJAL
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REG_ZERO

	case ASCALL, ARDCYCLE, ARDTIME, ARDINSTRET:
		i := encode(p.As)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = i.csr
		p.From3.Type = obj.TYPE_REG
		p.From3.Reg = REG_ZERO
		if p.To.Type == obj.TYPE_NONE {
			p.To.Type = obj.TYPE_REG
			p.To.Reg = REG_ZERO
		}

	// Rewrite MOV.
	case AMOV:
		switch p.From.Type {
		case obj.TYPE_REG: // MOV Ra, Rb -> ADDI $0, Ra, Rb
			p.As = AADDI
			*p.From3 = p.From
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = 0
		case obj.TYPE_CONST: // MOV $c, R -> ADD $c, ZERO, R
			p.As = AADDI
			p.From3.Type = obj.TYPE_REG
			p.From3.Reg = REG_ZERO
		}
	}
}

// follow can do some optimization on the structure of the program.  Currently,
// follow does nothing.
func follow(ctxt *obj.Link, s *obj.LSym) {}

// setpcs sets the Pc field in all instructions reachable from p.  It uses pc as
// the initial value.
func setpcs(p *obj.Prog, pc int64) {
	for ; p != nil; p = p.Link {
		p.Pc = pc
		if p.As != obj.ATEXT { // if this is a real instruction
			pc += 4
		}
	}
}

// invbr inverts the condition of a conditional branch.
func invbr(i obj.As) obj.As {
	switch i {
	case ABEQ:
		return ABNE
	case ABNE:
		return ABEQ
	case ABLT:
		return ABGE
	case ABGE:
		return ABLT
	case ABLTU:
		return ABGEU
	case ABGEU:
		return ABLTU
	default:
		panic("invbr: not a branch")
	}
}

// preprocess is called once for each linker symbol.  It generates prolog and
// epilog code and computes PC-relative branch and jump offsets.  By the time
// preprocess finishes, all instructions in the symbol are concrete, real RISC-V
// instructions.
func preprocess(ctxt *obj.Link, cursym *obj.LSym) {
	// Generate the prolog.
	text := cursym.Text
	if text.As != obj.ATEXT {
		ctxt.Diag("preprocess: found symbol that does not start with TEXT directive")
		return
	}
	stacksize := text.To.Offset
	// Insert stack adjustment.  Do not overwrite the TEXT directive itself;
	// other parts of the assembler assume it's there.
	spadj := obj.Appendp(ctxt, text)
	spadj.As = AADDI
	spadj.From.Type = obj.TYPE_CONST
	spadj.From.Offset = -stacksize
	spadj.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
	spadj.To.Type = obj.TYPE_REG
	spadj.To.Reg = REG_SP
	spadj.Spadj = int32(-stacksize)
	// Do, however, skip over the TEXT directive when generating assembly.
	// (It's not a valid RISC-V instruction, after all.)
	cursym.Text = spadj

	// Expand each long branch into a short branch and a jump.  This is a
	// fairly inefficient algorithm in theory, but it's only pathological
	// when there are a large quantity of long branches, which is unusual.
	setpcs(cursym.Text, 0)
	for p := cursym.Text; p != nil; {
		switch p.As {
		case ABEQ, ABNE, ABLT, ABGE, ABLTU, ABGEU:
			if p.To.Type != obj.TYPE_BRANCH {
				panic("assemble: instruction with branch-like opcode lacks destination")
				p = p.Link
				continue
			}
			offset := p.Pcond.Pc - p.Pc
			if offset < -4096 || 4096 <= offset {
				// Branch is long.  Replace it with a jump.
				jmp := obj.Appendp(ctxt, p)
				jmp.As = AJAL
				jmp.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_ZERO}
				jmp.To = obj.Addr{Type: obj.TYPE_BRANCH}
				jmp.Pcond = p.Pcond

				p.As = invbr(p.As)
				p.Pcond = jmp.Link
				// All future PCs are now invalid, so recompute
				// them.
				setpcs(jmp, p.Pc+4)
				// We may have made previous branches too long,
				// so recheck them.
				p = cursym.Text
			} else {
				// Branch is short.  No big deal.
				p = p.Link
			}
		default:
			p = p.Link
		}
	}

	// Now that there are no long branches, resolve branch and jump targets.
	for p := cursym.Text; p != nil; p = p.Link {
		switch p.As {
		case ABEQ, ABNE, ABLT, ABGE, ABLTU, ABGEU, AJAL:
			switch p.To.Type {
			case obj.TYPE_BRANCH:
				p.To.Type = obj.TYPE_CONST
				p.To.Offset = p.Pcond.Pc - p.Pc
			case obj.TYPE_MEM:
				panic("unhandled type")
			}
		}
	}
}

// regival validates an integer register.
func regival(r int16) uint32 {
	if r < REG_X0 || REG_X31 < r {
		panic("register out of range")
	}
	return uint32(r - obj.RBaseRISCV)
}

// regi extracts the integer register from an Addr.
func regi(a obj.Addr) uint32 {
	if a.Type != obj.TYPE_REG {
		panic(fmt.Sprintf("ill typed: %+v", a))
	}
	return regival(a.Reg)
}

// immi extracts the integer literal of the specified size from an Addr.
func immi(a obj.Addr, nbits uint) uint32 {
	if a.Type != obj.TYPE_CONST {
		panic(fmt.Sprintf("ill typed: %+v", a))
	}
	if a.Offset < -(1<<(nbits-1)) || (1<<(nbits-1))-1 < a.Offset {
		panic(fmt.Sprintf("immediate cannot fit in %d bits", nbits))
	}
	return uint32(a.Offset)
}

func instr_r(p *obj.Prog) uint32 {
	rs2 := regi(p.From)
	rs1 := regi(*p.From3)
	rd := regi(p.To)
	i := encode(p.As)
	if i == nil {
		panic("instr_r: could not encode instruction")
	}
	return i.funct7<<25 | rs2<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func instr_i(p *obj.Prog) uint32 {
	imm := immi(p.From, 12)
	rs1 := regi(*p.From3)
	rd := regi(p.To)
	i := encode(p.As)
	if i == nil {
		panic("instr_i: could not encode instruction")
	}
	return imm<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func instr_sb(p *obj.Prog) uint32 {
	imm := immi(p.To, 13)
	rs2 := regival(p.Reg)
	rs1 := regi(p.From)
	i := encode(p.As)
	if i == nil {
		panic("instr_sb: could not encode instruction")
	}
	return (imm>>12)<<31 |
		((imm>>5)&0x3f)<<25 |
		rs2<<20 |
		rs1<<15 |
		i.funct3<<12 |
		((imm>>1)&0xf)<<8 |
		((imm>>11)&0x1)<<7 |
		i.opcode
}

func instr_uj(p *obj.Prog) uint32 {
	imm := immi(p.To, 21)
	rd := regi(p.From)
	i := encode(p.As)
	if i == nil {
		panic("instr_uj: could not encode instruction")
	}
	return (imm>>20)<<31 |
		((imm>>1)&0x3ff)<<21 |
		((imm>>11)&0x1)<<20 |
		((imm>>12)&0xff)<<12 |
		rd<<7 |
		i.opcode
}

// asmout generates the machine code for a Prog.
func asmout(p *obj.Prog) uint32 {
	switch p.As {
	case AADD, ASUB, ASLL, AXOR, ASRL, ASRA, AOR, AAND:
		return instr_r(p)
	case AADDI, ASLLI, AXORI, ASRLI, ASRAI, AORI, AANDI, AJALR, ASCALL,
		ARDCYCLE, ARDTIME, ARDINSTRET:
		return instr_i(p)
	case ABEQ, ABNE, ABLT, ABGE, ABLTU, ABGEU:
		return instr_sb(p)
	case AJAL:
		return instr_uj(p)
	}
	panic("asmout: unrecognized instruction")
	return 0
}

// assemble is called at the very end of the assembly process.  It actually
// emits machine code.
func assemble(ctxt *obj.Link, cursym *obj.LSym) {
	var symcode []uint32 // machine code for this symbol
	for p := cursym.Text; p != nil; p = p.Link {
		symcode = append(symcode, asmout(p))
	}

	cursym.Grow(int64(4 * len(symcode)))
	for p, i := cursym.P, 0; i < len(symcode); p, i = p[4:], i+1 {
		ctxt.Arch.ByteOrder.PutUint32(p, symcode[i])
	}
}
