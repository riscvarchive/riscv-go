//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
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
	"cmd/internal/obj"
	"cmd/internal/obj/riscv"
	"cmd/link/internal/ld"
	"fmt"
	"log"
)

func gentext(ctxt *ld.Link) {
}

func adddynrela(ctxt *ld.Link, rel *ld.Symbol, s *ld.Symbol, r *ld.Reloc) {
	log.Fatalf("adddynrela not implemented")
}

func adddynrel(ctxt *ld.Link, s *ld.Symbol, r *ld.Reloc) {
	log.Fatalf("adddynrel not implemented")
}

func elfreloc1(ctxt *ld.Link, r *ld.Reloc, sectoff int64) int {
	log.Printf("elfreloc1")
	return -1
}

func elfsetupplt(ctxt *ld.Link) {
	log.Printf("elfsetuplt")
	return
}

func machoreloc1(ctxt *ld.Link, r *ld.Reloc, sectoff int64) int {
	log.Fatalf("machoreloc1 not implemented")
	return -1
}

func archreloc(ctxt *ld.Link, r *ld.Reloc, s *ld.Symbol, val *int64) int {
	switch r.Type {
	case obj.R_CALLRISCV:
		pc := s.Value + int64(r.Off)
		off := ld.Symaddr(ctxt, r.Sym) + r.Add - pc

		// This is always a JAL instruction, we just need
		// to replace the immediate.
		//
		// TODO(prattmic): sanity check the opcode.
		if off&1 != 0 {
			ctxt.Diag("R_CALLRISCV relocation for %s is not aligned: %#x", r.Sym.Name, off)
			return 0
		}

		imm, err := riscv.EncodeUJImmediate(off)
		if err != nil {
			// TODO(mpratt): We can only fit 1MB offsets in the
			// immediate, which we will quickly outgrow.
			ctxt.Diag("cannot encode R_CALLRISCV relocation offset for %s: %v", r.Sym.Name, err)
			return 0
		}

		// The assembler is encoding a normal JAL instruction with the
		// immediate as whatever p.To.Offset. We need to replace that
		// immediate with the relocated value.
		*val = (*val &^ riscv.UJTypeImmMask) | int64(imm)

	case obj.R_RISCV_PCREL_ITYPE, obj.R_RISCV_PCREL_STYPE:
		pc := s.Value + int64(r.Off)
		off := ld.Symaddr(ctxt, r.Sym) + r.Add - pc

		// Generate AUIPC and second instruction immediates.
		low, high, err := riscv.Split32BitImmediate(off)
		if err != nil {
			ctxt.Diag("R_RISCV_PCREL_ relocation does not fit in 32-bits: %d", off)
			return 0
		}

		auipcImm, err := riscv.EncodeUImmediate(high)
		if err != nil {
			ctxt.Diag("cannot encode R_RISCV_PCREL_ AUIPC relocation offset for %s: %v", r.Sym.Name, err)
			return 0
		}

		var secondImm, secondImmMask int64
		switch r.Type {
		case obj.R_RISCV_PCREL_ITYPE:
			secondImmMask = riscv.ITypeImmMask
			secondImm, err = riscv.EncodeIImmediate(low)
			if err != nil {
				ctxt.Diag("cannot encode R_RISCV_PCREL_ITYPE I-type instruction relocation offset for %s: %v", r.Sym.Name, err)
				return 0
			}
		case obj.R_RISCV_PCREL_STYPE:
			secondImmMask = riscv.STypeImmMask
			secondImm, err = riscv.EncodeSImmediate(low)
			if err != nil {
				ctxt.Diag("cannot encode R_RISCV_PCREL_STYPE S-type instruction relocation offset for %s: %v", r.Sym.Name, err)
				return 0
			}
		default:
			panic(fmt.Sprintf("Unknown relocation type: %v", r.Type))
		}

		auipc := int64(uint32(*val))
		second := int64(uint32(*val >> 32))

		auipc = (auipc &^ riscv.UTypeImmMask) | auipcImm
		second = (second &^ secondImmMask) | secondImm

		*val = second<<32 | auipc

	default:
		return -1
	}

	return 0
}

func archrelocvariant(ctxt *ld.Link, r *ld.Reloc, s *ld.Symbol, t int64) int64 {
	log.Printf("archrelocvariant")
	return -1
}

// TODO(mpratt): Refactor asmb for other archs. This worked literally copied
// and pasted from mips64.
func asmb(ctxt *ld.Link) {
	if ctxt.Debugvlog != 0 {
		fmt.Fprintf(ctxt.Bso, "%5.2f asmb\n", obj.Cputime())
	}
	ctxt.Bso.Flush()

	if ld.Iself {
		ld.Asmbelfsetup(ctxt)
	}

	sect := ld.Segtext.Sect
	ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
	ld.Codeblk(ctxt, int64(sect.Vaddr), int64(sect.Length))
	for sect = sect.Next; sect != nil; sect = sect.Next {
		ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
		ld.Datblk(ctxt, int64(sect.Vaddr), int64(sect.Length))
	}

	if ld.Segrodata.Filelen > 0 {
		if ctxt.Debugvlog != 0 {
			fmt.Fprintf(ctxt.Bso, "%5.2f rodatblk\n", obj.Cputime())
		}
		ctxt.Bso.Flush()

		ld.Cseek(int64(ld.Segrodata.Fileoff))
		ld.Datblk(ctxt, int64(ld.Segrodata.Vaddr), int64(ld.Segrodata.Filelen))
	}

	if ctxt.Debugvlog != 0 {
		fmt.Fprintf(ctxt.Bso, "%5.2f datblk\n", obj.Cputime())
	}
	ctxt.Bso.Flush()

	ld.Cseek(int64(ld.Segdata.Fileoff))
	ld.Datblk(ctxt, int64(ld.Segdata.Vaddr), int64(ld.Segdata.Filelen))

	ld.Cseek(int64(ld.Segdwarf.Fileoff))
	ld.Dwarfblk(ctxt, int64(ld.Segdwarf.Vaddr), int64(ld.Segdwarf.Filelen))

	/* output symbol table */
	ld.Symsize = 0

	ld.Lcsize = 0
	symo := uint32(0)
	if !*ld.FlagS {
		if ctxt.Debugvlog != 0 {
			fmt.Fprintf(ctxt.Bso, "%5.2f sym\n", obj.Cputime())
		}
		ctxt.Bso.Flush()
		switch ld.HEADTYPE {
		default:
			if ld.Iself {
				symo = uint32(ld.Segdwarf.Fileoff + ld.Segdata.Filelen)
				symo = uint32(ld.Rnd(int64(symo), int64(*ld.FlagRound)))
			}

		case obj.Hplan9:
			log.Fatalf("Plan 9 unsupported")
		}

		ld.Cseek(int64(symo))
		switch ld.HEADTYPE {
		default:
			if ld.Iself {
				if ctxt.Debugvlog != 0 {
					fmt.Fprintf(ctxt.Bso, "%5.2f elfsym\n", obj.Cputime())
				}
				ld.Asmelfsym(ctxt)
				ld.Cflush()
				ld.Cwrite(ld.Elfstrdat)

				if ld.Linkmode == ld.LinkExternal {
					ld.Elfemitreloc(ctxt)
				}
			}

		case obj.Hplan9:
			log.Fatalf("Plan 9 unsupported")
		}
	}

	ctxt.Cursym = nil
	if ctxt.Debugvlog != 0 {
		fmt.Fprintf(ctxt.Bso, "%5.2f header\n", obj.Cputime())
	}
	ctxt.Bso.Flush()
	ld.Cseek(0)
	switch ld.HEADTYPE {
	default:
		log.Fatalf("Unsupported OS: %v", ld.HEADTYPE)

	case obj.Hlinux,
		obj.Hfreebsd,
		obj.Hnetbsd,
		obj.Hopenbsd,
		obj.Hnacl:
		ld.Asmbelf(ctxt, int64(symo))
	}

	ld.Cflush()
	if *ld.FlagC {
		fmt.Printf("textsize=%d\n", ld.Segtext.Filelen)
		fmt.Printf("datsize=%d\n", ld.Segdata.Filelen)
		fmt.Printf("bsssize=%d\n", ld.Segdata.Length-ld.Segdata.Filelen)
		fmt.Printf("symsize=%d\n", ld.Symsize)
		fmt.Printf("lcsize=%d\n", ld.Lcsize)
		fmt.Printf("total=%d\n", ld.Segtext.Filelen+ld.Segdata.Length+uint64(ld.Symsize)+uint64(ld.Lcsize))
	}
}
