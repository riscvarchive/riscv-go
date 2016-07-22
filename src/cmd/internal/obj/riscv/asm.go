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

// addrtoreg extracts the register from an addr, handling SB and SP.
func addrtoreg(a obj.Addr) int16 {
	switch a.Name {
	case obj.NAME_EXTERN:
		return REG_SB
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

	// Rewrite MOV.
	case AMOV, AMOVB, AMOVH, AMOVW, AMOVBU, AMOVHU, AMOVWU:
		switch p.From.Type {
		case obj.TYPE_MEM: // MOV c(Rs), Rd -> L $c, Rs, Rd
			if p.To.Type != obj.TYPE_REG {
				ctxt.Diag("progedit: unsupported load at %v", p)
			}
			p.As = movtol(p.As)
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: addrtoreg(p.From)}
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: p.From.Offset}
		case obj.TYPE_REG:
			switch p.To.Type {
			case obj.TYPE_REG: // MOV Ra, Rb -> ADDI $0, Ra, Rb
				if p.As != AMOV {
					ctxt.Diag("progedit: unsupported register-register move at %v", p)
				}
				p.As = AADDI
				*p.From3 = p.From
				p.From = obj.Addr{Type: obj.TYPE_CONST}
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
				p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: p.From.Offset}
				p.From3.Type = obj.TYPE_REG
				p.To = obj.Addr{Type: obj.TYPE_REG, Reg: addrtoreg(p.To)}
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
			// If the immediate can be generated by sign-extending a 12 bit number, an ADDI suffices.
			needLUI := !immFits(off, 12)
			if needLUI {
				if !immFits(off, 32) {
					// TODO: use a constant pool for 64 bit constants?
					// Or remove REG_TMP from the general purposes registers used by the compiler
					// and emulate riscv.rules, using REG_TMP as the 32 bit value staging ground?
					ctxt.Diag("%v: constant %d too large; see riscv.rules MOVQconst for how to make a 64 bit constant ", p, off)
				}
				p.As = ALUI
				p.To = to
				// Pass top 20 bits to LUI.
				p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: off >> 12}
				// The bottom 12 bits will be treated as signed.
				// If that will result in a negative 12 bit number,
				// add 1 to our upper bits to adjust for the borrow.
				// It is not possible for this increment to overflow.
				// To overflow, the 20 top bits would be 1,
				// and the sign bit for the low 12 bits would be set,
				// in which case the entire 32 bit pattern fits in
				// a 12 bit signed value.
				if off&(1<<11) != 0 {
					p.From.Offset++
				}
				p = obj.Appendp(ctxt, p)

				// Generate our new low 12 bit value.
				low := off
				low &= 1<<12 - 1 // mask off the bits we just handled
				// Generate upper sign bits, leaving space for the bottom 12 bits.
				off = int64(low >> 11)
				off <<= 63
				off >>= 64 - 12
				off |= low // put the low bits into place
			}
			p.As = AADDI
			p.To = to
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: off}
			p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_ZERO}
			if needLUI {
				p.From3.Reg = to.Reg
			}

		case obj.TYPE_ADDR: // MOV $sym+off(SP/SB), R
			if p.To.Type != obj.TYPE_REG || p.As != AMOV {
				ctxt.Diag("progedit: unsupported addr MOV at %v", p)
			}
			p.As = AADDI
			p.From3.Type = obj.TYPE_REG
			p.From.Type = obj.TYPE_CONST
			switch p.From.Name {
			case obj.NAME_EXTERN:
				p.From3.Reg = REG_SB
			case obj.NAME_PARAM, obj.NAME_AUTO:
				p.From3.Reg = REG_SP
			default:
				ctxt.Diag("progedit: bad addr MOV from name %v at %v", p.From.Name, p)
			}
		default:
			ctxt.Diag("progedit: unsupported MOV at %v", p)
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

// preprocess generates prologue and epilogue code and computes PC-relative branch and jump offsets.
// preprocess is called once for each linker symbol.
// When preprocess finishes, all instructions in the symbol are either concrete, real RISC-V instructions
// or directive pseudo-ops like TEXT, PCDATA, and FUNCDATA.
func preprocess(ctxt *obj.Link, cursym *obj.LSym) {
	// Generate the prologue.
	text := cursym.Text
	if text.As != obj.ATEXT {
		ctxt.Diag("preprocess: found symbol that does not start with TEXT directive")
		return
	}
	stacksize := text.To.Offset
	// Insert stack adjustment if necessary.
	if stacksize != 0 {
		spadj := obj.Appendp(ctxt, text)
		spadj.As = AADDI
		spadj.From.Type = obj.TYPE_CONST
		spadj.From.Offset = -stacksize
		spadj.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: REG_SP}
		spadj.To.Type = obj.TYPE_REG
		spadj.To.Reg = REG_SP
		spadj.Spadj = int32(-stacksize)
	}

	// Replace RET with epilogue.
	for p := cursym.Text; p != nil; p = p.Link {
		if p.As == obj.ARET {
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

// regival validates an integer register.
func regival(r int16) uint32 {
	if r < REG_X0 || REG_X31 < r {
		panic(fmt.Sprintf("register out of range, want %d < %d < %d", REG_X0, r, REG_X31))
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

// wantIntReg checks that a contains an integer register.
func wantIntReg(p *obj.Prog, pos string, a obj.Addr) {
	if a.Type != obj.TYPE_REG {
		p.Ctxt.Diag("%v\texpected register in %s position but got %s", p, pos, p.Ctxt.Dconv(&a))
		return
	}
	if a.Reg < REG_X0 || REG_X31 < a.Reg {
		p.Ctxt.Diag("%v\texpected integer register in %s position but got non-integer register %s", p, pos, p.Ctxt.Dconv(&a))
	}
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

func validateR(p *obj.Prog) {
	wantIntReg(p, "from", p.From)
	wantIntReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func encodeR(p *obj.Prog) uint32 {
	rs2 := regi(p.From)
	rs1 := regi(*p.From3)
	rd := regi(p.To)
	i, ok := encode(p.As)
	if !ok {
		panic("encodeR: could not encode instruction")
	}
	return i.funct7<<25 | rs2<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func validateI(p *obj.Prog) {
	wantImm(p, "from", p.From, 12)
	wantIntReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func encodeI(p *obj.Prog) uint32 {
	imm := immi(p.From, 12)
	rs1 := regi(*p.From3)
	rd := regi(p.To)
	i, ok := encode(p.As)
	if !ok {
		panic("encodeI: could not encode instruction")
	}
	imm |= uint32(i.csr)
	return imm<<20 | rs1<<15 | i.funct3<<12 | rd<<7 | i.opcode
}

func validateS(p *obj.Prog) {
	wantImm(p, "from", p.From, 12)
	wantIntReg(p, "from3", *p.From3)
	wantIntReg(p, "to", p.To)
}

func encodeS(p *obj.Prog) uint32 {
	imm := immi(p.From, 12)
	rs2 := regi(*p.From3)
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
	rs2 := regival(p.Reg)
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

type encoding struct {
	encode   func(*obj.Prog) uint32 // encode returns the machine code for a Prog
	validate func(*obj.Prog)        // validate validates a Prog, calling ctxt.Diag for any issues
	length   int64                  // length of encoded instruction; 0 for pseudo-ops, 4 otherwise
}

var (
	rEncoding  = encoding{encode: encodeR, validate: validateR, length: 4}
	iEncoding  = encoding{encode: encodeI, validate: validateI, length: 4}
	sEncoding  = encoding{encode: encodeS, validate: validateS, length: 4}
	sbEncoding = encoding{encode: encodeSB, validate: validateSB, length: 4}
	uEncoding  = encoding{encode: encodeU, validate: validateU, length: 4}
	ujEncoding = encoding{encode: encodeUJ, validate: validateUJ, length: 4}

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
	AADD & obj.AMask:    rEncoding,
	ASUB & obj.AMask:    rEncoding,
	ASLL & obj.AMask:    rEncoding,
	AXOR & obj.AMask:    rEncoding,
	ASRL & obj.AMask:    rEncoding,
	ASRA & obj.AMask:    rEncoding,
	AOR & obj.AMask:     rEncoding,
	AAND & obj.AMask:    rEncoding,
	ASLT & obj.AMask:    rEncoding,
	ASLTU & obj.AMask:   rEncoding,
	AMUL & obj.AMask:    rEncoding,
	AMULH & obj.AMask:   rEncoding,
	AMULHU & obj.AMask:  rEncoding,
	AMULHSU & obj.AMask: rEncoding,
	AMULW & obj.AMask:   rEncoding,
	ADIV & obj.AMask:    rEncoding,
	ADIVU & obj.AMask:   rEncoding,
	AREM & obj.AMask:    rEncoding,
	AREMU & obj.AMask:   rEncoding,
	ADIVW & obj.AMask:   rEncoding,
	ADIVUW & obj.AMask:  rEncoding,
	AREMW & obj.AMask:   rEncoding,
	AREMUW & obj.AMask:  rEncoding,

	AADDI & obj.AMask:      iEncoding,
	ASLLI & obj.AMask:      iEncoding,
	AXORI & obj.AMask:      iEncoding,
	ASRLI & obj.AMask:      iEncoding,
	ASRAI & obj.AMask:      iEncoding,
	AORI & obj.AMask:       iEncoding,
	AANDI & obj.AMask:      iEncoding,
	AJALR & obj.AMask:      iEncoding,
	AECALL & obj.AMask:     iEncoding,
	ARDCYCLE & obj.AMask:   iEncoding,
	ARDTIME & obj.AMask:    iEncoding,
	ARDINSTRET & obj.AMask: iEncoding,
	ALB & obj.AMask:        iEncoding,
	ALH & obj.AMask:        iEncoding,
	ALW & obj.AMask:        iEncoding,
	ALD & obj.AMask:        iEncoding,
	ALBU & obj.AMask:       iEncoding,
	ALHU & obj.AMask:       iEncoding,
	ALWU & obj.AMask:       iEncoding,
	ASLTI & obj.AMask:      iEncoding,
	ASLTIU & obj.AMask:     iEncoding,

	ASB & obj.AMask: sEncoding,
	ASH & obj.AMask: sEncoding,
	ASW & obj.AMask: sEncoding,
	ASD & obj.AMask: sEncoding,

	ABEQ & obj.AMask:  sbEncoding,
	ABNE & obj.AMask:  sbEncoding,
	ABLT & obj.AMask:  sbEncoding,
	ABGE & obj.AMask:  sbEncoding,
	ABLTU & obj.AMask: sbEncoding,
	ABGEU & obj.AMask: sbEncoding,

	AAUIPC & obj.AMask: uEncoding,
	ALUI & obj.AMask:   uEncoding,

	AJAL & obj.AMask: ujEncoding,

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
				// This is a CALL which needs a relocation.
				rel := obj.Addrel(cursym)
				rel.Off = int32(p.Pc)
				rel.Siz = 4
				rel.Sym = p.To.Sym
				rel.Add = p.To.Offset
				rel.Type = obj.R_CALLRISCV
			}
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
