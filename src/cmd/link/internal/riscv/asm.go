package riscv

import (
	"cmd/link/internal/ld"
	"log"
)

func adddynrel(s *ld.LSym, r *ld.Reloc) {
	// TODO(bbaren): Implement
	log.Printf("adddynrel: s: %+v r: %+v", s, r)
}

func archreloc(r *ld.Reloc, s *ld.LSym, val *int64) int {
	// TODO(bbaren): Implement
	log.Printf("archreloc: r: %+v s: %+v val: %+v", r, s, val)
	return 0
}

func archrelocvariant(r *ld.Reloc, s *ld.LSym, t int64) int64 {
	// TODO(bbaren): Implement
	log.Printf("archrelocvariant: r: %+v s: %+v t: %+v", r, s, t)
	return t
}

func asmb() {
	// TODO(bbaren): Implement
	log.Printf("asmb")
}

func elfreloc1(r *ld.Reloc, sectoff int64) int {
	// TODO(bbaren): Implement
	log.Printf("elfreloc1: r: %+v sectoff: %+v", r, sectoff)
	return 0
}

func elfsetupplt() {
	// TODO(bbaren): Implement
	log.Printf("elfsetupplt")
}

func gentext() {
	// TODO(bbaren): Implement
	log.Printf("gentext")
}

func machoreloc1(r *ld.Reloc, sectoff int64) int {
	// TODO(bbaren): Implement
	log.Printf("machoreloc1: r: %+v sectoff: %+v", r, sectoff)
	return 0
}
