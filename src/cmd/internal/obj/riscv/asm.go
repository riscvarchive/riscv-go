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
//
// The Go assembler framework occasionally abuses certain fields in the Prog and
// Addr structs.  For instance, the instruction
//
//   JAL T1, label
//
// jumps to the address ZERO+label and stores a linkage pointer in T1.  Since
// ZERO is an input register and T1 is an output register, you might expect the
// assembler's parser to set From to be ZERO and To to be T1--but you'd be
// wrong!  Instead, From is T1 and To is ZERO.  Repairing this infelicity would
// require changes to the parser and every assembler backend, so until that
// cleanup occurs, the authors have tried to document specific gotchas where
// they occur.  Be on the lookout.

package riscv

import (
	"fmt"

	"cmd/internal/obj"
)

// resolvepseudoreg concretizes pseudo-registers in an Addr.
func resolvepseudoreg(a *obj.Addr) {
	if a.Type == obj.TYPE_MEM {
		switch a.Name {
		case obj.NAME_PARAM:
			a.Reg = REG_FP
		}
	}
}

// lowerjalr normalizes a JALR instruction.
func lowerjalr(p *obj.Prog) {
	if p.As != AJALR {
		panic("lowerjalr: not a JALR")
	}

	// JALR gets parsed like JAL--the linkage pointer goes in From, and the
	// target is in To.  However, we need to assemble it as an I-type
	// instruction--the linkage pointer will go in To, the target register
	// in From3, and the offset in From.
	//
	// TODO(bbaren): Handle sym, symkind, index, and scale.
	p.From, *p.From3, p.To = p.To, p.To, p.From
	p.From.Type = obj.TYPE_CONST
	p.From3.Type = obj.TYPE_REG
}

// movtol converts a MOV[BHW]?U? mnemonic into the corresponding L[BHWD]
// instruction.
func movtol(mnemonic obj.As) obj.As {
	switch mnemonic {
	case AMOV:
		return ALD
	case AMOVB:
		return ALB
	case AMOVH:
		return ALH
	case AMOVW:
		return ALW
	case AMOVBU:
		return ALBU
	case AMOVHU:
		return ALHU
	case AMOVWU:
		return ALWU
	default:
		panic(fmt.Sprintf("%+v is not a MOV", mnemonic))
	}
}

// movtos converts a MOV[BHW]? mnemonic into the corresponding S[BHWD]
// instruction.
func movtos(mnemonic obj.As) obj.As {
	switch mnemonic {
	case AMOV:
		return ASD
	case AMOVB:
		return ASB
	case AMOVH:
		return ASH
	case AMOVW:
		return ASW
	default:
		panic(fmt.Sprintf("%+v is not a MOV", mnemonic))
	}
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
	if p.From3.Type == obj.TYPE_CONST {
		switch p.As {
		case ASLT:
			p.As = ASLTI
		case ASLTU:
			p.As = ASLTIU
		}
	}

	// Concretize pseudo-registers.
	resolvepseudoreg(&p.From)
	resolvepseudoreg(p.From3)
	resolvepseudoreg(&p.To)

	// Do additional single-instruction rewriting.
	switch p.As {
	// Turn JMP into JAL ZERO or JALR ZERO.
	case obj.AJMP:
		// p.From is actually an _output_ for this instruction.
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REG_ZERO
		switch p.To.Type {
		case obj.TYPE_BRANCH:
			p.As = AJAL
		case obj.TYPE_MEM:
			p.As = AJALR
			lowerjalr(p)
		default:
			panic(fmt.Sprintf("unhandled type %+v", p.To.Type))
		}

	case AJALR:
		lowerjalr(p)

	case ASCALL, ARDCYCLE, ARDTIME, ARDINSTRET:
		i, ok := encode(p.As)
		if !ok {
			panic("progedit: tried to rewrite nonexistent instruction")
		}
		p.From.Type = obj.TYPE_CONST
		// The CSR isn't exactly an offset, but it winds up in the
		// immediate area of the encoded instruction, so record it in
		// the Offset field.
		p.From.Offset = i.csr
		p.From3.Type = obj.TYPE_REG
		p.From3.Reg = REG_ZERO
		if p.To.Type == obj.TYPE_NONE {
			p.To.Type = obj.TYPE_REG
			p.To.Reg = REG_ZERO
		}

	// Rewrite MOV.
	case AMOV, AMOVB, AMOVH, AMOVW, AMOVBU, AMOVHU, AMOVWU:
		switch p.From.Type {
		case obj.TYPE_MEM: // MOV c(Rs), Rd -> L $c, Rs, Rd
			if p.To.Type != obj.TYPE_REG {
				ctxt.Diag("progedit: unsupported load at %v", p)
			}
			p.As = movtol(p.As)
			*p.From3 = p.From
			p.From.Type = obj.TYPE_CONST
			p.From3.Type = obj.TYPE_REG
		case obj.TYPE_REG:
			switch p.To.Type {
			case obj.TYPE_REG: // MOV Ra, Rb -> ADDI $0, Ra, Rb
				if p.As != AMOV {
					ctxt.Diag("progedit: unsupported register-register move at %v", p)
				}
				p.As = AADDI
				*p.From3 = p.From
				p.From.Type = obj.TYPE_CONST
				p.From.Offset = 0
			case obj.TYPE_MEM: // MOV Rs, c(Rd) -> S $c, Rs, Rd
				switch p.As {
				case AMOVBU, AMOVHU, AMOVWU:
					ctxt.Diag("progedit: unsupported unsigned store at %v", p)
				}
				p.As = movtos(p.As)
				// The destination address goes in p.From and
				// p.To here, with the offset in p.From and the
				// register in p.To.  The data register goes in
				// p.From3.
				p.From, *p.From3 = p.To, p.From
				p.From.Type = obj.TYPE_CONST
				p.From3.Type = obj.TYPE_REG
				p.To.Type = obj.TYPE_REG
			default:
				ctxt.Diag("progedit: unsupported MOV at %v", p)
			}
		case obj.TYPE_CONST: // MOV $c, R -> ADD $c, ZERO, R
			if p.As != AMOV {
				ctxt.Diag("progedit: unsupported constant load at %v", p)
			}
			p.As = AADDI
			p.From3.Type = obj.TYPE_REG
			p.From3.Reg = REG_ZERO
		default:
			ctxt.Diag("progedit: unsupported MOV at %v", p)
		}

	// The semantics for SLT are designed to make sense when writing
	// assembly from right to left--for instance, slt t2,t1,t0 sets t2 if
	// t1 < t0.  Go assembly is written from left to right, though, so
	// switch the operands around so you can write SLT T0, T1, T2 instead.
	case ASLT, ASLTI, ASLTU, ASLTIU:
		p.From, *p.From3 = *p.From3, p.From
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

// preprocess is called once for each linker symbol.  It generates prologue and
// epilogue code and computes PC-relative branch and jump offsets.  By the time
// preprocess finishes, all instructions in the symbol are concrete, real RISC-V
// instructions.
func preprocess(ctxt *obj.Link, cursym *obj.LSym) {
	// Generate the prologue.
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

	// Delete unneeded instructions.
	var prev *obj.Prog
	for p := cursym.Text; p != nil; p = p.Link {
		switch p.As {
		case obj.AFUNCDATA:
			if prev != nil {
				prev.Link = p.Link
			} else {
				cursym.Text = p.Link
			}
		default:
			prev = p
		}
	}

	// Replace RET with epilogue.
	for p := cursym.Text; p != nil; p = p.Link {
		if p.As == obj.ARET {
			p.As = AADDI
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = stacksize
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
			p.To.Type = obj.TYPE_REG
			p.To.Reg = REG_SP
			p.Spadj = int32(stacksize)

			p = obj.Appendp(ctxt, p)
			p.As = AJALR
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = 0
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_RA}
			p.To.Type = obj.TYPE_REG
			p.To.Reg = REG_ZERO
		}
	}

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

func position(p *obj.Prog, a obj.Addr) string {
	switch {
	case p.From == a:
		return "from"
	case p.To == a:
		return "to"
	case p.From3 != nil && *p.From3 == a:
		return "from3"
	}
	return "???"
}

// regival validates an integer register.
func regival(r int16) uint32 {
	if r < REG_X0 || REG_X31 < r {
		panic("register out of range")
	}
	return uint32(r - obj.RBaseRISCV)
}

// regi extracts the integer register from an Addr.
func regi(p *obj.Prog, a obj.Addr) uint32 {
	if a.Type != obj.TYPE_REG {
		p.Ctxt.Diag("%v\texpected register in %s position but got %s", p, position(p, a), p.Ctxt.Dconv(&a))
		return 0
	}
	return regival(a.Reg)
}

// immi extracts the integer literal of the specified size from an Addr.
func immi(p *obj.Prog, a obj.Addr, nbits uint) uint32 {
	if a.Type != obj.TYPE_CONST {
		p.Ctxt.Diag("%v\texpected immediate in %s position but got %s", p, position(p, a), p.Ctxt.Dconv(&a))
	}
	if a.Offset < -(1<<(nbits-1)) || (1<<(nbits-1))-1 < a.Offset {
		p.Ctxt.Diag("%v\timmediate %d does not fit in %d bits", p, a.Offset, nbits)
	}
	return uint32(a.Offset)
}

func instr_r(p *obj.Prog) uint32 {
	rs2 := regi(p, p.From)
	rs1 := regi(p, *p.From3)
	rd := regi(p, p.To)
	i, ok := encode(p.As)
	if !ok {
		panic("instr_r: could not encode instruction")
	}
	return i.funct7<<25 | rs2<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func instr_i(p *obj.Prog) uint32 {
	imm := immi(p, p.From, 12)
	rs1 := regi(p, *p.From3)
	rd := regi(p, p.To)
	i, ok := encode(p.As)
	if !ok {
		panic("instr_i: could not encode instruction")
	}
	return imm<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func instr_s(p *obj.Prog) uint32 {
	imm := immi(p, p.From, 12)
	rs2 := regi(p, *p.From3)
	rs1 := regi(p, p.To)
	i, ok := encode(p.As)
	if !ok {
		panic("instr_i: could not encode instruction")
	}
	return (imm>>5)<<25 |
		rs2<<20 |
		rs1<<15 |
		i.funct3<<12 |
		(imm&0x1f)<<7 |
		i.opcode
}

func instr_sb(p *obj.Prog) uint32 {
	imm := immi(p, p.To, 13)
	rs2 := regival(p.Reg)
	rs1 := regi(p, p.From)
	i, ok := encode(p.As)
	if !ok {
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

func instr_u(p *obj.Prog) uint32 {
	imm := immi(p, p.To, 21)
	rd := regi(p, p.From)
	i, ok := encode(p.As)
	if !ok {
		panic("instr_u: could not encode instruction")
	}
	return imm<<12 | rd<<7 | i.opcode
}

func instr_uj(p *obj.Prog) uint32 {
	imm := immi(p, p.To, 21)
	rd := regi(p, p.From)
	i, ok := encode(p.As)
	if !ok {
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
	case AADD, ASUB, ASLL, AXOR, ASRL, ASRA, AOR, AAND, ASLT, ASLTU:
		return instr_r(p)
	case AADDI, ASLLI, AXORI, ASRLI, ASRAI, AORI, AANDI, AJALR, ASCALL,
		ARDCYCLE, ARDTIME, ARDINSTRET, ALB, ALH, ALW, ALD, ALBU, ALHU,
		ALWU, ASLTI, ASLTIU:
		return instr_i(p)
	case ASB, ASH, ASW, ASD:
		return instr_s(p)
	case ABEQ, ABNE, ABLT, ABGE, ABLTU, ABGEU:
		return instr_sb(p)
	case AAUIPC, ALUI:
		return instr_u(p)
	case AJAL:
		return instr_uj(p)
	}
	panic(fmt.Sprintf("asmout: unrecognized instruction %s", obj.Aconv(p.As)))
}

// assemble is called at the very end of the assembly process.  It actually
// emits machine code.
func assemble(ctxt *obj.Link, cursym *obj.LSym) {
	var symcode []uint32 // machine code for this symbol
	for p := cursym.Text; p != nil; p = p.Link {
		symcode = append(symcode, asmout(p))
	}
	cursym.Size = int64(4 * len(symcode))

	cursym.Grow(cursym.Size)
	for p, i := cursym.P, 0; i < len(symcode); p, i = p[4:], i+1 {
		ctxt.Arch.ByteOrder.PutUint32(p, symcode[i])
	}
}
