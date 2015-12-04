package riscv

import (
	"cmd/internal/obj"
	"cmd/link/internal/ld"
	"fmt"
	"log"
)

func Main() {
	linkarchinit()
	ld.Ldmain()
}

func linkarchinit() {
	ld.Thestring = "riscv"
	ld.Thelinkarch = &ld.Linkriscv

	ld.Thearch.Thechar = thechar
	ld.Thearch.Ptrsize = ld.Thelinkarch.Ptrsize
	ld.Thearch.Intsize = ld.Thelinkarch.Ptrsize
	ld.Thearch.Regsize = ld.Thelinkarch.Regsize
	ld.Thearch.Funcalign = FuncAlign
	ld.Thearch.Maxalign = MaxAlign
	ld.Thearch.Minlc = MINLC
	ld.Thearch.Dwarfregsp = DWARFREGSP
	ld.Thearch.Dwarfreglr = DWARFREGLR

	ld.Thearch.Adddynrel = adddynrel
	ld.Thearch.Archinit = archinit
	ld.Thearch.Archreloc = archreloc
	ld.Thearch.Archrelocvariant = archrelocvariant
	ld.Thearch.Asmb = asmb
	ld.Thearch.Elfreloc1 = elfreloc1
	ld.Thearch.Elfsetupplt = elfsetupplt
	ld.Thearch.Gentext = gentext
	ld.Thearch.Machoreloc1 = machoreloc1
	ld.Thearch.Lput = ld.Lputl
	ld.Thearch.Wput = ld.Wputl
	ld.Thearch.Vput = ld.Vputl

	// A quick Google search suggests that ld.so.1 is the preferred dynamic
	// linker name for RISC-V Linux.
	ld.Thearch.Linuxdynld = "/lib/ld.so.1"
	ld.Thearch.Netbsddynld = "/libexec/ld.elf_so"
	ld.Thearch.Freebsddynld = "XXX"   // port exists, but progress unclear
	ld.Thearch.Openbsddynld = "XXX"   // no port known
	ld.Thearch.Dragonflydynld = "XXX" // no port known
	ld.Thearch.Solarisdynld = "XXX"   // no port known
}

func archinit() {
	log.Printf("archinit")

	// TODO(bbaren): Support external linking
	if ld.Linkmode == ld.LinkExternal {
		log.Fatalf("-linkmode=external is unsupported on RISC-V")
	} else {
		ld.Linkmode = ld.LinkInternal
	}

	// TODO(bbaren): This is mostly cargo-culted off the arm and amd64
	// backends.  Figure out what this does and document it.
	switch ld.HEADTYPE {
	default:
		ld.Exitf("unknown -H option: %v", ld.HEADTYPE)

	case obj.Hlinux:
		ld.Elfinit()
		ld.HEADR = ld.ELFRESERVE
		if ld.INITTEXT == -1 {
			ld.INITTEXT = (1 << 22) + int64(ld.HEADR)
		}
		if ld.INITDAT == -1 {
			ld.INITDAT = 0
		}
		if ld.INITRND == -1 {
			ld.INITRND = 4096
		}
	}

	if ld.INITDAT != 0 && ld.INITRND != 0 {
		fmt.Printf("warning: -D0x%x is ignored because of -R0x%x\n",
			uint64(ld.INITDAT), uint32(ld.INITRND))
	}
}
