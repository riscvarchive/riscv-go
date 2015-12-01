package riscv

import (
	"cmd/link/internal/ld"
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
	// TODO(bbaren): Implement
	log.Printf("archinit")
}
