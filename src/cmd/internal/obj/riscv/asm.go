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
	"cmd/internal/obj"
	"fmt"
)

// stackOffset updates Addr offsets based on the current stack size.
//
// The stack looks like:
// -------------------
// |                 |
// |      PARAMs     |
// |                 |
// |                 |
// -------------------
// |    Parent RA    |   SP on function entry
// -------------------
// |                 |
// |                 |
// |       AUTOs     |
// |                 |
// |                 |
// -------------------
// |        RA       |   SP during function execution
// -------------------
//
// FixedFrameSize makes other packages aware of the space allocated for RA.
//
// Slide 21 on the presention attached to
// https://golang.org/issue/16922#issuecomment-243748180 has a nicer version
// of this diagram.
func stackOffset(a *obj.Addr, stacksize int64) {
	switch a.Name {
	case obj.NAME_AUTO:
		// Adjust to the top of AUTOs.
		a.Offset += stacksize
	case obj.NAME_PARAM:
		// Adjust to the bottom of PARAMs.
		a.Offset += stacksize + 8
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

// movtol converts a MOV mnemonic into the corresponding load instruction.
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
	case AMOVF:
		return AFLW
	case AMOVD:
		return AFLD
	default:
		panic(fmt.Sprintf("%+v is not a MOV", mnemonic))
	}
}

// movtos converts a MOV mnemonic into the corresponding store instruction.
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
	case AMOVF:
		return AFSW
	case AMOVD:
		return AFSD
	default:
		panic(fmt.Sprintf("%+v is not a MOV", mnemonic))
	}
}

// addrtoreg extracts the register from an Addr, handling special Addr.Names.
func addrtoreg(a obj.Addr) int16 {
	switch a.Name {
	case obj.NAME_PARAM, obj.NAME_AUTO:
		return REG_SP
	}
	return a.Reg
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
		case AADD, ASUB, ASLL, AXOR, ASRL, ASRA, AOR, AAND, AMUL, AMULH,
			AMULHU, AMULHSU, AMULW, ADIV, ADIVU, AREM, AREMU, ADIVW,
			ADIVUW, AREMW, AREMUW:
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
		case ASLT:
			p.As = ASLTI
		case ASLTU:
			p.As = ASLTIU
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
	// Turn JMP into JAL ZERO or JALR ZERO.
	case obj.AJMP:
		// p.From is actually an _output_ for this instruction.
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REG_ZERO

		switch p.To.Type {
		case obj.TYPE_BRANCH:
			p.As = AJAL
		case obj.TYPE_MEM:
			switch p.To.Name {
			case obj.NAME_AUTO, obj.NAME_PARAM, obj.NAME_NONE:
				p.As = AJALR
				lowerjalr(p)
			case obj.NAME_EXTERN:
				// JMP to symbol.
				p.As = AJAL
				// We will emit a relocation for this. Until then,
				// we'll encode this as a constant.
				p.To.Type = obj.TYPE_CONST
			default:
				ctxt.Diag("progedit: unsupported name %d for %v", p.To.Name, p)
			}
		default:
			panic(fmt.Sprintf("unhandled type %+v", p.To.Type))
		}

	case obj.ACALL:
		// p.From is actually an _output_ for this instruction.
		p.From.Type = obj.TYPE_REG
		p.From.Reg = REG_RA

		switch p.To.Type {
		case obj.TYPE_MEM:
			p.As = AJAL
			// We will emit a relocation for this. Until then,
			// we'll encode this as a constant.
			p.To.Type = obj.TYPE_CONST
		default:
			ctxt.Diag("unknown destination type %+v (want TYPE_MEM) in CALL: %v", p.To.Type, p)
		}

	case AJALR:
		lowerjalr(p)

	case AECALL, ASCALL, ARDCYCLE, ARDTIME, ARDINSTRET:
		// SCALL is the old name for ECALL.
		if p.As == ASCALL {
			p.As = AECALL
		}

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

	case ASEQZ:
		// SEQZ rs, rd -> SLTIU $1, rs, rd
		p.As = ASLTIU
		*p.From3 = p.From
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 1}

	case ASNEZ:
		// SNEZ rs, rd -> SLTU rs, x0, rd
		p.As = ASLTU
		*p.From3 = obj.Addr{Type: obj.TYPE_REG, Reg: REG_ZERO}

	// For binary float instructions, use From3 and To, not From and
	// To. This helps simplify encoding.
	case AFNEGS:
		// FNEGS rs, rd -> FSGNJNS rs, rs, rd
		p.As = AFSGNJNS
		*p.From3 = p.From
	case AFNEGD:
		// FNEGD rs, rd -> FSGNJND rs, rs, rd
		p.As = AFSGNJND
		*p.From3 = p.From
	case AFSQRTS, AFSQRTD:
		*p.From3 = p.From

		// This instruction expects a zero (i.e., float register 0) to
		// be the second input operand.
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_F0}
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
		pc += encodingForP(p).length
	}
}

// InvertBranch inverts the condition of a conditional branch.
func InvertBranch(i obj.As) obj.As {
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
		panic("InvertBranch: not a branch")
	}
}

// containsCall reports whether the symbol contains a CALL (or equivalent)
// instruction. Must be called after progedit.
func containsCall(sym *obj.LSym) bool {
	// CALLs are JAL with link register RA.
	for p := sym.Text; p != nil; p = p.Link {
		switch p.As {
		case AJAL:
			if p.From.Type == obj.TYPE_REG && p.From.Reg == REG_RA {
				return true
			}
		}
	}

	return false
}

// preprocess generates prologue and epilogue code, computes PC-relative branch
// and jump offsets, and resolves psuedo-registers.
//
// preprocess is called once per linker symbol.
//
// When preprocess finishes, all instructions in the symbol are either
// concrete, real RISC-V instructions or directive pseudo-ops like TEXT,
// PCDATA, and FUNCDATA.
func preprocess(ctxt *obj.Link, cursym *obj.LSym) {
	// Generate the prologue.
	text := cursym.Text
	if text.As != obj.ATEXT {
		ctxt.Diag("preprocess: found symbol that does not start with TEXT directive")
		return
	}

	stacksize := text.To.Offset

	// We must save RA if there is a CALL.
	saveRA := containsCall(cursym)
	// Unless we're told not to!
	if text.From3.Offset&obj.NOFRAME != 0 {
		saveRA = false
	}
	if saveRA {
		stacksize += 8
	}

	cursym.Args = text.To.Val.(int32)
	cursym.Locals = int32(stacksize)

	prologue := text

	// Insert stack adjustment if necessary.
	if stacksize != 0 {
		prologue = obj.Appendp(ctxt, prologue)
		prologue.As = AADDI
		prologue.From.Type = obj.TYPE_CONST
		prologue.From.Offset = -stacksize
		prologue.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
		prologue.To.Type = obj.TYPE_REG
		prologue.To.Reg = REG_SP
		prologue.Spadj = int32(-stacksize)
	}

	// Actually save RA.
	if saveRA {
		// Source register in From3, destination base register in To,
		// destination offset in From. See MOV TYPE_REG, TYPE_MEM below
		// for details.
		prologue = obj.Appendp(ctxt, prologue)
		prologue.As = ASD
		prologue.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_RA}
		prologue.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
		prologue.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
	}

	// Update stack-based offsets.
	for p := cursym.Text; p != nil; p = p.Link {
		stackOffset(&p.From, stacksize)
		if p.From3 != nil {
			stackOffset(p.From3, stacksize)
		}
		stackOffset(&p.To, stacksize)

		// TODO: update stacksize when instructions that modify SP are
		// found, or disallow it entirely.
	}

	// Additional instruction rewriting. Any rewrites that change the number
	// of instructions must occur here (i.e., before jump target
	// resolution).
	for p := cursym.Text; p != nil; p = p.Link {
		switch p.As {

		// Rewrite MOV. This couldn't be done in progedit, as SP
		// offsets needed to be applied before we split up some of the
		// Addrs.
		case AMOV, AMOVB, AMOVH, AMOVW, AMOVBU, AMOVHU, AMOVWU, AMOVF, AMOVD:
			switch p.From.Type {
			case obj.TYPE_MEM: // MOV c(Rs), Rd -> L $c, Rs, Rd
				switch p.From.Name {
				case obj.NAME_AUTO, obj.NAME_PARAM, obj.NAME_NONE:
					if p.To.Type != obj.TYPE_REG {
						ctxt.Diag("progedit: unsupported load at %v", p)
					}
					p.As = movtol(p.As)
					p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: addrtoreg(p.From)}
					p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: p.From.Offset}
				case obj.NAME_EXTERN:
					// AUIPC $off_hi, R
					// L $off_lo, R
					as := p.As
					to := p.To

					p.As = AAUIPC
					// This offset isn't really encoded
					// with either instruction. It will be
					// extracted for a relocation later.
					p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: p.From.Offset, Sym: p.From.Sym}
					p.From3 = &obj.Addr{}
					p.To = obj.Addr{Type: obj.TYPE_REG, Reg: to.Reg}
					p.Mark |= NEED_PCREL_ITYPE_RELOC
					p = obj.Appendp(ctxt, p)

					p.As = movtol(as)
					p.From = obj.Addr{Type: obj.TYPE_CONST}
					p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: to.Reg}
					p.To = to
				default:
					ctxt.Diag("progedit: unsupported name %d for %v", p.From.Name, p)
				}
			case obj.TYPE_REG:
				switch p.To.Type {
				case obj.TYPE_REG:
					switch p.As {
					case AMOV: // MOV Ra, Rb -> ADDI $0, Ra, Rb
						p.As = AADDI
						*p.From3 = p.From
						p.From = obj.Addr{Type: obj.TYPE_CONST}
					case AMOVF: // MOVF Ra, Rb -> FSGNJS Ra, Ra, Rb
						p.As = AFSGNJS
						*p.From3 = p.From
					case AMOVD: // MOVD Ra, Rb -> FSGNJD Ra, Ra, Rb
						p.As = AFSGNJD
						*p.From3 = p.From
					default:
						ctxt.Diag("progedit: unsupported register-register move at %v", p)
					}
				case obj.TYPE_MEM: // MOV Rs, c(Rd) -> S $c, Rs, Rd
					switch p.As {
					case AMOVBU, AMOVHU, AMOVWU:
						ctxt.Diag("progedit: unsupported unsigned store at %v", p)
					}
					switch p.To.Name {
					case obj.NAME_AUTO, obj.NAME_PARAM, obj.NAME_NONE:
						p.As = movtos(p.As)
						// The destination address goes in p.From and
						// p.To here, with the offset in p.From and the
						// register in p.To. The source register goes in
						// p.From3.
						p.From, *p.From3 = p.To, p.From
						p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: p.From.Offset}
						p.From3.Type = obj.TYPE_REG
						p.To = obj.Addr{Type: obj.TYPE_REG, Reg: addrtoreg(p.To)}
					case obj.NAME_EXTERN:
						// AUIPC $off_hi, TMP
						// S $off_lo, TMP, R
						as := p.As
						from := p.From

						p.As = AAUIPC
						// This offset isn't really encoded
						// with either instruction. It will be
						// extracted for a relocation later.
						p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: p.To.Offset, Sym: p.To.Sym}
						p.From3 = &obj.Addr{}
						p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
						p.Mark |= NEED_PCREL_STYPE_RELOC
						p = obj.Appendp(ctxt, p)

						p.As = movtos(as)
						p.From = obj.Addr{Type: obj.TYPE_CONST}
						p.From3 = &from
						p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
					default:
						ctxt.Diag("progedit: unsupported name %d for %v", p.From.Name, p)
					}
				default:
					ctxt.Diag("progedit: unsupported MOV at %v", p)
				}
			case obj.TYPE_CONST:
				// MOV $c, R
				// If c is small enough, convert to:
				//   ADD $c, ZERO, R
				// If not, convert to:
				//   LUI top20bits(c), R
				//   ADD bottom12bits(c), R, R
				if p.As != AMOV {
					ctxt.Diag("progedit: unsupported constant load at %v", p)
				}
				off := p.From.Offset
				to := p.To

				low, high, err := Split32BitImmediate(off)
				if err != nil {
					// TODO: use a constant pool for 64 bit constants?
					//
					// Or remove REG_TMP from the general purposes registers used by the compiler
					// and emulate riscv.rules, using REG_TMP as the 32 bit value staging ground?
					ctxt.Diag("%v: constant %d too large; see riscv.rules MOVQconst for how to make a 64 bit constant: %v", p, off, err)
				}

				// LUI is only necessary if the offset doesn't fit in 12-bits.
				needLUI := high != 0
				if needLUI {
					p.As = ALUI
					p.To = to
					// Pass top 20 bits to LUI.
					p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: high}
					p = obj.Appendp(ctxt, p)
				}
				p.As = AADDI
				p.To = to
				p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: low}
				p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_ZERO}
				if needLUI {
					p.From3.Reg = to.Reg
				}

			case obj.TYPE_ADDR: // MOV $sym+off(SP/SB), R
				if p.To.Type != obj.TYPE_REG || p.As != AMOV {
					ctxt.Diag("progedit: unsupported addr MOV at %v", p)
				}
				switch p.From.Name {
				case obj.NAME_EXTERN:
					// AUIPC $off_hi, R
					// ADDI $off_lo, R
					to := p.To

					p.As = AAUIPC
					// This offset isn't really encoded
					// with either instruction. It will be
					// extracted for a relocation later.
					p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: p.From.Offset, Sym: p.From.Sym}
					p.From3 = &obj.Addr{}
					p.To = to
					p.Mark |= NEED_PCREL_ITYPE_RELOC
					p = obj.Appendp(ctxt, p)

					p.As = AADDI
					p.From = obj.Addr{Type: obj.TYPE_CONST}
					p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: to.Reg}
					p.To = to
				case obj.NAME_PARAM, obj.NAME_AUTO:
					p.As = AADDI
					p.From3.Type = obj.TYPE_REG
					p.From.Type = obj.TYPE_CONST
					p.From3.Reg = REG_SP
				default:
					ctxt.Diag("progedit: bad addr MOV from name %v at %v", p.From.Name, p)
				}
			default:
				ctxt.Diag("progedit: unsupported MOV at %v", p)
			}

		// Replace RET with epilogue.
		case obj.ARET:
			if saveRA {
				// Restore RA.
				p.As = ALD
				p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
				p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
				p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_RA}
				p = obj.Appendp(ctxt, p)
			}

			if stacksize != 0 {
				p.As = AADDI
				p.From.Type = obj.TYPE_CONST
				p.From.Offset = stacksize
				p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
				p.To.Type = obj.TYPE_REG
				p.To.Reg = REG_SP
				p.Spadj = int32(stacksize)
				p = obj.Appendp(ctxt, p)
			}

			p.As = AJALR
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = 0
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_RA}
			p.To.Type = obj.TYPE_REG
			p.To.Reg = REG_ZERO

		// Replace FNE[SD] with FEQ[SD] and NOT.
		case AFNES:
			if p.To.Type != obj.TYPE_REG {
				ctxt.Diag("progedit: FNES needs an integer register output")
			}
			dst := p.To.Reg
			p.As = AFEQS
			p := obj.Appendp(ctxt, p)
			p.As = AXORI // [bit] xor 1 = not [bit]
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = 1
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: dst}
			p.To.Type = obj.TYPE_REG
			p.To.Reg = dst
		case AFNED:
			if p.To.Type != obj.TYPE_REG {
				ctxt.Diag("progedit: FNED needs an integer register output")
			}
			dst := p.To.Reg
			p.As = AFEQD
			p := obj.Appendp(ctxt, p)
			p.As = AXORI
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = 1
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: dst}
			p.To.Type = obj.TYPE_REG
			p.To.Reg = dst
		}
	}

	// Split immediates larger that 12-bits
	for p := cursym.Text; p != nil; p = p.Link {
		switch p.As {
		// <opi> $imm, FROM3, TO
		case AADDI, AANDI, AORI, AXORI:
			// LUI $high, TMP
			// ADDI $low, TMP, TMP
			// <op> TMP, FROM3, TO
			q := *p
			low, high, err := Split32BitImmediate(p.From.Offset)
			if err != nil {
				ctxt.Diag("%v: constant %d too large", p, p.From.Offset, err)
			}
			if high == 0 {
				break // no need to split
			}

			p.As = ALUI
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: high}
			p.From3 = nil
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p.Spadj = 0 // needed if TO is SP
			p = obj.Appendp(ctxt, p)

			p.As = AADDI
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: low}
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p = obj.Appendp(ctxt, p)

			switch q.As {
			case AADDI:
				p.As = AADD
			case AANDI:
				p.As = AAND
			case AORI:
				p.As = AOR
			case AXORI:
				p.As = AXOR
			default:
				ctxt.Diag("progedit: unsupported inst %v for splitting", q)
			}
			p.Spadj = q.Spadj
			p.To = q.To
			p.From3 = q.From3
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}

		// <load> $imm, FROM3, TO (load $imm+(FROM3), TO)
		case ALD, ALB, ALH, ALW, ALBU, ALHU:
			// LUI $high, TMP
			// ADDI $low, TMP, TMP
			// ADD TMP, FROM3, TMP
			// <load> $0, TMP, TO
			q := *p
			low, high, err := Split32BitImmediate(p.From.Offset)
			if err != nil {
				ctxt.Diag("%v: constant %d too large", p, p.From.Offset)
			}
			if high == 0 {
				break // no need to split
			}

			p.As = ALUI
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: high}
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p = obj.Appendp(ctxt, p)

			p.As = AADDI
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: low}
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p = obj.Appendp(ctxt, p)

			p.As = AADD
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p.From3 = q.From3
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p = obj.Appendp(ctxt, p)

			p.As = q.As
			p.To = q.To
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}

		// <store> $imm, FROM3, TO (store $imm+(TO), FROM3)
		case ASD, ASB, ASH, ASW:
			// LUI $high, TMP
			// ADDI $low, TMP, TMP
			// ADD TMP, TO, TMP
			// <store> $0, FROM3, TMP
			q := *p
			low, high, err := Split32BitImmediate(p.From.Offset)
			if err != nil {
				ctxt.Diag("%v: constant %d too large", p, p.From.Offset)
			}
			if high == 0 {
				break // no need to split
			}

			p.As = ALUI
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: high}
			p.From3 = nil
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p = obj.Appendp(ctxt, p)

			p.As = AADDI
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: low}
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p = obj.Appendp(ctxt, p)

			p.As = AADD
			p.From = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: q.To.Reg}
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p = obj.Appendp(ctxt, p)

			p.As = q.As
			p.From3 = q.From3
			p.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_TMP}
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
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

				p.As = InvertBranch(p.As)
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
	// At this point, instruction rewriting which changes the number of
	// instructions will break everything--don't do it!
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

	// Validate all instructions. This provides nice error messages.
	for p := cursym.Text; p != nil; p = p.Link {
		encodingForP(p).validate(p)
	}
}

// signExtend sign extends val starting at bit bit.
func signExtend(val int64, bit uint) int64 {
	// Mask off the bits to keep.
	low := val
	low &= 1<<bit - 1

	// Generate upper sign bits, leaving space for the bottom bits.
	val >>= bit - 1
	val <<= 63
	val >>= 64 - bit
	val |= low // put the low bits into place.

	return val
}

// Split32BitImmediate splits a signed 32-bit immediate into a signed 20-bit
// upper immediate and a signed 12-bit lower immediate to be added to the upper
// result.
//
// For example, high may be used in LUI and low in a following ADDI to generate
// a full 32-bit constant.
func Split32BitImmediate(imm int64) (low, high int64, err error) {
	if !immFits(imm, 32) {
		return 0, 0, fmt.Errorf("immediate does not fit in 32-bits: %d", imm)
	}

	// Nothing special needs to be done if the immediate fits in 12-bits.
	if immFits(imm, 12) {
		return imm, 0, nil
	}

	high = imm >> 12
	// The bottom 12 bits will be treated as signed.
	//
	// If that will result in a negative 12 bit number, add 1 to
	// our upper bits to adjust for the borrow.
	//
	// It is not possible for this increment to overflow. To
	// overflow, the 20 top bits would be 1, and the sign bit for
	// the low 12 bits would be set, in which case the entire 32
	// bit pattern fits in a 12 bit signed value.
	if imm&(1<<11) != 0 {
		high++
	}

	high = signExtend(high, 20)
	low = signExtend(imm, 12)

	return
}

func regval(r int16, min int16, max int16) uint32 {
	if r < min || max < r {
		panic(fmt.Sprintf("register out of range, want %d < %d < %d", min, r, max))
	}
	return uint32(r - min)
}

func reg(a obj.Addr, min int16, max int16) uint32 {
	if a.Type != obj.TYPE_REG {
		panic(fmt.Sprintf("ill typed: %+v", a))
	}
	return regval(a.Reg, min, max)
}

// regi extracts the integer register from an Addr.
func regi(a obj.Addr) uint32 { return reg(a, REG_X0, REG_X31) }

// regf extracts the float register from an Addr.
func regf(a obj.Addr) uint32 { return reg(a, REG_F0, REG_F31) }

func wantReg(p *obj.Prog, pos string, a obj.Addr, descr string, min int16, max int16) {
	if a.Type != obj.TYPE_REG {
		p.Ctxt.Diag("%v\texpected register in %s position but got %s",
			p, pos, p.Ctxt.Dconv(&a))
		return
	}
	if a.Reg < min || max < a.Reg {
		p.Ctxt.Diag("%v\texpected %s register in %s position but got non-%s register %s",
			p, descr, pos, descr, p.Ctxt.Dconv(&a))
	}
}

// wantIntReg checks that a contains an integer register.
func wantIntReg(p *obj.Prog, pos string, a obj.Addr) {
	wantReg(p, pos, a, "integer", REG_X0, REG_X31)
}

// wantFloatReg checks that a contains a floating-point register.
func wantFloatReg(p *obj.Prog, pos string, a obj.Addr) {
	wantReg(p, pos, a, "float", REG_F0, REG_F31)
}

// immFits reports whether immediate value x fits in nbits bits.
func immFits(x int64, nbits uint) bool {
	nbits--
	var min int64 = -1 << nbits
	var max int64 = 1<<nbits - 1
	return min <= x && x <= max
}

// immi extracts the integer literal of the specified size from an Addr.
func immi(a obj.Addr, nbits uint) uint32 {
	if a.Type != obj.TYPE_CONST {
		panic(fmt.Sprintf("ill typed: %+v", a))
	}
	if !immFits(a.Offset, nbits) {
		panic(fmt.Sprintf("immediate %d in %v cannot fit in %d bits", a.Offset, a, nbits))
	}
	return uint32(a.Offset)
}

func wantImm(p *obj.Prog, pos string, a obj.Addr, nbits uint) {
	if a.Type != obj.TYPE_CONST {
		p.Ctxt.Diag("%v\texpected immediate in %s position but got %s", p, pos, p.Ctxt.Dconv(&a))
		return
	}
	if !immFits(a.Offset, nbits) {
		p.Ctxt.Diag("%v\timmediate in %s position cannot be larger than %d bits but got %d", p, pos, nbits, a.Offset)
	}
}

func wantEvenJumpOffset(p *obj.Prog) {
	if p.To.Offset%1 != 0 {
		p.Ctxt.Diag("%v\tjump offset %v must be even", p, p.Ctxt.Dconv(&p.To))
	}
}

func validateRIII(p *obj.Prog) {
	wantIntReg(p, "from", p.From)
	wantIntReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func validateRFFF(p *obj.Prog) {
	wantFloatReg(p, "from", p.From)
	wantFloatReg(p, "from3", *p.From3)
	wantFloatReg(p, "to", p.To)
}

func validateRFFI(p *obj.Prog) {
	wantFloatReg(p, "from", p.From)
	wantFloatReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func validateRFI(p *obj.Prog) {
	wantFloatReg(p, "from", p.From)
	wantIntReg(p, "to", p.To)
}

func validateRIF(p *obj.Prog) {
	wantIntReg(p, "from", p.From)
	wantFloatReg(p, "to", p.To)
}

func validateRFF(p *obj.Prog) {
	wantFloatReg(p, "from", p.From)
	wantFloatReg(p, "to", p.To)
}

func encodeR(p *obj.Prog, rs1 uint32, rs2 uint32, rd uint32) uint32 {
	i, ok := encode(p.As)
	if !ok {
		panic("encodeR: could not encode instruction")
	}
	if i.rs2 != 0 && rs2 != 0 {
		panic("encodeR: instruction uses rs2, but rs2 was nonzero")
	}

	return i.funct7<<25 | i.rs2<<20 | rs2<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func encodeRIII(p *obj.Prog) uint32 {
	return encodeR(p, regi(*p.From3), regi(p.From), regi(p.To))
}

func encodeRFFF(p *obj.Prog) uint32 {
	return encodeR(p, regf(*p.From3), regf(p.From), regf(p.To))
}

func encodeRFFI(p *obj.Prog) uint32 {
	return encodeR(p, regf(*p.From3), regf(p.From), regi(p.To))
}

func encodeRFI(p *obj.Prog) uint32 {
	return encodeR(p, regf(p.From), 0, regi(p.To))
}

func encodeRIF(p *obj.Prog) uint32 {
	return encodeR(p, regi(p.From), 0, regf(p.To))
}

func encodeRFF(p *obj.Prog) uint32 {
	return encodeR(p, regf(p.From), 0, regf(p.To))
}

func validateII(p *obj.Prog) {
	wantImm(p, "from", p.From, 12)
	wantIntReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func validateIF(p *obj.Prog) {
	wantImm(p, "from", p.From, 12)
	wantIntReg(p, "from3", *p.From3)
	wantFloatReg(p, "to", p.To)
}

func encodeI(p *obj.Prog, rd uint32) uint32 {
	imm := immi(p.From, 12)
	rs1 := regi(*p.From3)
	i, ok := encode(p.As)
	if !ok {
		panic("encodeI: could not encode instruction")
	}
	imm |= uint32(i.csr)
	return imm<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func encodeII(p *obj.Prog) uint32 {
	return encodeI(p, regi(p.To))
}

func encodeIF(p *obj.Prog) uint32 {
	return encodeI(p, regf(p.To))
}

func validateSI(p *obj.Prog) {
	wantImm(p, "from", p.From, 12)
	wantIntReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func validateSF(p *obj.Prog) {
	wantImm(p, "from", p.From, 12)
	wantFloatReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func EncodeSImmediate(imm int64) (int64, error) {
	if !immFits(imm, 12) {
		return 0, fmt.Errorf("immediate %#x does not fit in 12 bits", imm)
	}

	return ((imm >> 5) << 25) | ((imm & 0x1f) << 7), nil
}

func encodeS(p *obj.Prog, rs2 uint32) uint32 {
	imm := immi(p.From, 12)
	rs1 := regi(p.To)
	i, ok := encode(p.As)
	if !ok {
		panic("encodeS: could not encode instruction")
	}
	return (imm>>5)<<25 |
		rs2<<20 |
		rs1<<15 |
		i.funct3<<12 |
		(imm&0x1f)<<7 |
		i.opcode
}

func encodeSI(p *obj.Prog) uint32 {
	return encodeS(p, regi(*p.From3))
}

func encodeSF(p *obj.Prog) uint32 {
	return encodeS(p, regf(*p.From3))
}

func validateSB(p *obj.Prog) {
	// Offsets are multiples of two, so accept 13 bit immediates for the 12 bit slot.
	// We implicitly drop the least significant bit in encodeSB.
	wantEvenJumpOffset(p)
	wantImm(p, "to", p.To, 13)
	// TODO: validate that the register from p.Reg is in range
	wantIntReg(p, "from", p.From)
}

func encodeSB(p *obj.Prog) uint32 {
	imm := immi(p.To, 13)
	rs2 := regval(p.Reg, REG_X0, REG_X31)
	rs1 := regi(p.From)
	i, ok := encode(p.As)
	if !ok {
		panic("encodeSB: could not encode instruction")
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

func validateU(p *obj.Prog) {
	wantImm(p, "from", p.From, 20)
	wantIntReg(p, "to", p.To)
}

func encodeU(p *obj.Prog) uint32 {
	// The immediates for encodeU are the upper 20 bits of a 32 bit value.
	// Rather than have the user/compiler generate a 32 bit constant,
	// the bottommost bits of which must all be zero,
	// instead accept just the top bits.
	imm := immi(p.From, 20)
	rd := regi(p.To)
	i, ok := encode(p.As)
	if !ok {
		panic("encodeU: could not encode instruction")
	}
	return imm<<12 | rd<<7 | i.opcode
}

func EncodeIImmediate(imm int64) (int64, error) {
	if !immFits(imm, 12) {
		return 0, fmt.Errorf("immediate %#x does not fit in 12 bits", imm)
	}

	return imm << 20, nil
}

func EncodeUImmediate(imm int64) (int64, error) {
	if !immFits(imm, 20) {
		return 0, fmt.Errorf("immediate %#x does not fit in 20 bits", imm)
	}

	return imm << 12, nil
}

func validateUJ(p *obj.Prog) {
	// Offsets are multiples of two, so accept 21 bit immediates for the 20 bit slot.
	// We implicitly drop the least significant bit in encodeUJ.
	wantEvenJumpOffset(p)
	wantImm(p, "to", p.To, 21)
	wantIntReg(p, "from", p.From)
}

// encodeUJImmediate encodes a UJ-type immediate. imm must fit in 21-bits.
func encodeUJImmediate(imm uint32) uint32 {
	return (imm>>20)<<31 |
		((imm>>1)&0x3ff)<<21 |
		((imm>>11)&0x1)<<20 |
		((imm>>12)&0xff)<<12
}

// EncodeUJImmediate encodes a UJ-type immediate.
func EncodeUJImmediate(imm int64) (uint32, error) {
	if !immFits(imm, 21) {
		return 0, fmt.Errorf("immediate %#x does not fit in 21 bits", imm)
	}
	return encodeUJImmediate(uint32(imm)), nil
}

func encodeUJ(p *obj.Prog) uint32 {
	imm := encodeUJImmediate(immi(p.To, 21))
	rd := regi(p.From)
	i, ok := encode(p.As)
	if !ok {
		panic("encodeUJ: could not encode instruction")
	}
	return imm | rd<<7 | i.opcode
}

func validateRaw(p *obj.Prog) {
	wantImm(p, "raw", p.From, 32)
}

func encodeRaw(p *obj.Prog) uint32 {
	return immi(p.From, 32)
}

type encoding struct {
	encode   func(*obj.Prog) uint32 // encode returns the machine code for a Prog
	validate func(*obj.Prog)        // validate validates a Prog, calling ctxt.Diag for any issues
	length   int64                  // length of encoded instruction; 0 for pseudo-ops, 4 otherwise
}

var (
	// Encodings have the following naming convention:
	//	1. the instruction encoding (R/I/S/SB/U/UJ), in lowercase
	//	2. zero or more register operand identifiers (I = integer
	//	   register, F = float register), in uppercase
	//	3. the word "Encoding"
	// For example, rIIIEncoding indicates an R-type instruction with two
	// integer register inputs and an integer register output; sFEncoding
	// indicates an S-type instruction with rs2 being a float register.

	rIIIEncoding = encoding{encode: encodeRIII, validate: validateRIII, length: 4}
	rFFFEncoding = encoding{encode: encodeRFFF, validate: validateRFFF, length: 4}
	rFFIEncoding = encoding{encode: encodeRFFI, validate: validateRFFI, length: 4}
	rFIEncoding  = encoding{encode: encodeRFI, validate: validateRFI, length: 4}
	rIFEncoding  = encoding{encode: encodeRIF, validate: validateRIF, length: 4}
	rFFEncoding  = encoding{encode: encodeRFF, validate: validateRFF, length: 4}

	iIEncoding = encoding{encode: encodeII, validate: validateII, length: 4}
	iFEncoding = encoding{encode: encodeIF, validate: validateIF, length: 4}

	sIEncoding = encoding{encode: encodeSI, validate: validateSI, length: 4}
	sFEncoding = encoding{encode: encodeSF, validate: validateSF, length: 4}

	sbEncoding = encoding{encode: encodeSB, validate: validateSB, length: 4}

	uEncoding = encoding{encode: encodeU, validate: validateU, length: 4}

	ujEncoding = encoding{encode: encodeUJ, validate: validateUJ, length: 4}

	rawEncoding = encoding{encode: encodeRaw, validate: validateRaw, length: 4}

	// pseudoOpEncoding panics if encoding is attempted, but does no validation.
	pseudoOpEncoding = encoding{encode: nil, validate: func(*obj.Prog) {}, length: 0}

	// badEncoding is used when an invalid op is encountered.
	// An error has already been generated, so let anything else through.
	badEncoding = encoding{encode: func(*obj.Prog) uint32 { return 0 }, validate: func(*obj.Prog) {}, length: 0}
)

// encodingForAs contains the encoding for a RISC-V instruction.
// Instructions are masked with obj.AMask to keep indices small.
// TODO: merge this with the encoding table in inst.go.
// TODO: add other useful per-As info, like whether it is a branch (used in preprocess).
var encodingForAs = [...]encoding{
	// 2.5: Control Transfer Instructions
	AJAL & obj.AMask:  ujEncoding,
	AJALR & obj.AMask: iIEncoding,
	ABEQ & obj.AMask:  sbEncoding,
	ABNE & obj.AMask:  sbEncoding,
	ABLT & obj.AMask:  sbEncoding,
	ABLTU & obj.AMask: sbEncoding,
	ABGE & obj.AMask:  sbEncoding,
	ABGEU & obj.AMask: sbEncoding,

	// 2.9: Environment Call and Breakpoints
	AECALL & obj.AMask: iIEncoding,

	// 4.2: Integer Computational Instructions
	AADDI & obj.AMask:  iIEncoding,
	ASLTI & obj.AMask:  iIEncoding,
	ASLTIU & obj.AMask: iIEncoding,
	AANDI & obj.AMask:  iIEncoding,
	AORI & obj.AMask:   iIEncoding,
	AXORI & obj.AMask:  iIEncoding,
	ASLLI & obj.AMask:  iIEncoding,
	ASRLI & obj.AMask:  iIEncoding,
	ASRAI & obj.AMask:  iIEncoding,
	ALUI & obj.AMask:   uEncoding,
	AAUIPC & obj.AMask: uEncoding,
	AADD & obj.AMask:   rIIIEncoding,
	ASLT & obj.AMask:   rIIIEncoding,
	ASLTU & obj.AMask:  rIIIEncoding,
	AAND & obj.AMask:   rIIIEncoding,
	AOR & obj.AMask:    rIIIEncoding,
	AXOR & obj.AMask:   rIIIEncoding,
	ASLL & obj.AMask:   rIIIEncoding,
	ASRL & obj.AMask:   rIIIEncoding,
	ASUB & obj.AMask:   rIIIEncoding,
	ASRA & obj.AMask:   rIIIEncoding,

	// 4.3: Load and Store Instructions
	ALD & obj.AMask:  iIEncoding,
	ALW & obj.AMask:  iIEncoding,
	ALWU & obj.AMask: iIEncoding,
	ALH & obj.AMask:  iIEncoding,
	ALHU & obj.AMask: iIEncoding,
	ALB & obj.AMask:  iIEncoding,
	ALBU & obj.AMask: iIEncoding,
	ASD & obj.AMask:  sIEncoding,
	ASW & obj.AMask:  sIEncoding,
	ASH & obj.AMask:  sIEncoding,
	ASB & obj.AMask:  sIEncoding,

	// 4.4: System Instructions
	ARDCYCLE & obj.AMask:   iIEncoding,
	ARDTIME & obj.AMask:    iIEncoding,
	ARDINSTRET & obj.AMask: iIEncoding,

	// 5.1: Multiplication Operations
	AMUL & obj.AMask:    rIIIEncoding,
	AMULH & obj.AMask:   rIIIEncoding,
	AMULHU & obj.AMask:  rIIIEncoding,
	AMULHSU & obj.AMask: rIIIEncoding,
	AMULW & obj.AMask:   rIIIEncoding,
	ADIV & obj.AMask:    rIIIEncoding,
	ADIVU & obj.AMask:   rIIIEncoding,
	AREM & obj.AMask:    rIIIEncoding,
	AREMU & obj.AMask:   rIIIEncoding,
	ADIVW & obj.AMask:   rIIIEncoding,
	ADIVUW & obj.AMask:  rIIIEncoding,
	AREMW & obj.AMask:   rIIIEncoding,
	AREMUW & obj.AMask:  rIIIEncoding,

	// 7.5: Single-Precision Load and Store Instructions
	AFLW & obj.AMask: iFEncoding,
	AFSW & obj.AMask: sFEncoding,

	// 7.6: Single-Precision Floating-Point Computational Instructions
	AFADDS & obj.AMask:  rFFFEncoding,
	AFSUBS & obj.AMask:  rFFFEncoding,
	AFMULS & obj.AMask:  rFFFEncoding,
	AFDIVS & obj.AMask:  rFFFEncoding,
	AFSQRTS & obj.AMask: rFFFEncoding,

	// 7.7: Single-Precision Floating-Point Conversion and Move Instructions
	AFCVTWS & obj.AMask:  rFIEncoding,
	AFCVTLS & obj.AMask:  rFIEncoding,
	AFCVTSW & obj.AMask:  rIFEncoding,
	AFCVTSL & obj.AMask:  rIFEncoding,
	AFSGNJS & obj.AMask:  rFFFEncoding,
	AFSGNJNS & obj.AMask: rFFFEncoding,
	AFSGNJXS & obj.AMask: rFFFEncoding,
	AFMVSX & obj.AMask:   rIFEncoding,

	// 7.8: Single-Precision Floating-Point Compare Instructions
	AFEQS & obj.AMask: rFFIEncoding,
	AFLTS & obj.AMask: rFFIEncoding,
	AFLES & obj.AMask: rFFIEncoding,

	// 8.2: Double-Precision Load and Store Instructions
	AFLD & obj.AMask: iFEncoding,
	AFSD & obj.AMask: sFEncoding,

	// 8.3: Double-Precision Floating-Point Computational Instructions
	AFADDD & obj.AMask:  rFFFEncoding,
	AFSUBD & obj.AMask:  rFFFEncoding,
	AFMULD & obj.AMask:  rFFFEncoding,
	AFDIVD & obj.AMask:  rFFFEncoding,
	AFSQRTD & obj.AMask: rFFFEncoding,

	// 8.4: Double-Precision Floating-Point Conversion and Move Instructions
	AFCVTWD & obj.AMask:  rFIEncoding,
	AFCVTLD & obj.AMask:  rFIEncoding,
	AFCVTDW & obj.AMask:  rIFEncoding,
	AFCVTDL & obj.AMask:  rIFEncoding,
	AFCVTSD & obj.AMask:  rFFEncoding,
	AFCVTDS & obj.AMask:  rFFEncoding,
	AFSGNJD & obj.AMask:  rFFFEncoding,
	AFSGNJND & obj.AMask: rFFFEncoding,
	AFSGNJXD & obj.AMask: rFFFEncoding,
	AFMVDX & obj.AMask:   rIFEncoding,

	// 8.5: Double-Precision Floating-Point Compare Instructions
	AFEQD & obj.AMask: rFFIEncoding,
	AFLTD & obj.AMask: rFFIEncoding,
	AFLED & obj.AMask: rFFIEncoding,

	// Escape hatch
	AWORD & obj.AMask: rawEncoding,

	// Pseudo-operations
	obj.AFUNCDATA: pseudoOpEncoding,
	obj.APCDATA:   pseudoOpEncoding,
	obj.ATEXT:     pseudoOpEncoding,
	obj.AUNDEF:    pseudoOpEncoding,
}

// encodingForP returns the encoding (encode+validate funcs) for a Prog.
func encodingForP(p *obj.Prog) encoding {
	if base := p.As &^ obj.AMask; base != obj.ABaseRISCV && base != 0 {
		p.Ctxt.Diag("encodingForP: not a RISC-V instruction %s", p.As)
		return badEncoding
	}
	as := p.As & obj.AMask
	if int(as) >= len(encodingForAs) {
		p.Ctxt.Diag("encodingForP: bad RISC-V instruction %s", p.As)
		return badEncoding
	}
	enc := encodingForAs[as]
	if enc.validate == nil {
		p.Ctxt.Diag("encodingForP: no encoding for instruction %s", p.As)
		return badEncoding
	}
	return enc
}

// assemble emits machine code.
// It is called at the very end of the assembly process.
func assemble(ctxt *obj.Link, cursym *obj.LSym) {
	var symcode []uint32 // machine code for this symbol
	for p := cursym.Text; p != nil; p = p.Link {
		switch p.As {
		case AJAL:
			if p.To.Sym != nil {
				// This is a CALL/JMP which needs a relocation.
				rel := obj.Addrel(cursym)
				rel.Off = int32(p.Pc)
				rel.Siz = 4
				rel.Sym = p.To.Sym
				rel.Add = p.To.Offset
				rel.Type = obj.R_CALLRISCV
			}
		case AAUIPC:
			var t obj.RelocType
			if p.Mark&NEED_PCREL_ITYPE_RELOC == NEED_PCREL_ITYPE_RELOC {
				t = obj.R_RISCV_PCREL_ITYPE
			} else if p.Mark&NEED_PCREL_STYPE_RELOC == NEED_PCREL_STYPE_RELOC {
				t = obj.R_RISCV_PCREL_STYPE
			} else {
				break
			}
			if p.Link == nil {
				ctxt.Diag("AUIPC needing PC-relative reloc missing following instruction")
				break
			}
			if p.From.Sym == nil {
				ctxt.Diag("AUIPC needing PC-relative reloc missing symbol")
				break
			}

			rel := obj.Addrel(cursym)
			rel.Off = int32(p.Pc)
			rel.Siz = 8
			rel.Sym = p.From.Sym
			rel.Add = p.From.Offset
			rel.Type = t
		}

		enc := encodingForP(p)
		if enc.length > 0 {
			symcode = append(symcode, enc.encode(p))
		}
	}
	cursym.Size = int64(4 * len(symcode))

	cursym.Grow(cursym.Size)
	for p, i := cursym.P, 0; i < len(symcode); p, i = p[4:], i+1 {
		ctxt.Arch.ByteOrder.PutUint32(p, symcode[i])
	}
}
