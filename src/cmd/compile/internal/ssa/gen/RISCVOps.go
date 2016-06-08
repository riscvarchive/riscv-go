// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import "cmd/internal/obj/riscv"

func init() {
	var regNamesRISCV []string
	var gpMask regMask
	var fpMask regMask

	// Build the list of register names, creating an appropriately indexed
	// regMask for the gp and fp registers as we go.
	for r := riscv.REG_X0; r <= riscv.REG_X31; r++ {
		gpMask |= regMask(1) << uint(len(regNamesRISCV))
		regNamesRISCV = append(regNamesRISCV, "."+riscv.RegNames[int16(r)])
	}
	for r := riscv.REG_F0; r <= riscv.REG_F31; r++ {
		fpMask |= regMask(1) << uint(len(regNamesRISCV))
		regNamesRISCV = append(regNamesRISCV, "."+riscv.RegNames[int16(r)])
	}

	if len(regNamesRISCV) > 64 {
		// regMask is only 64 bits.
		panic("Too many RISCV registers")
	}


	gp := regInfo{inputs: []regMask{gpMask}, outputs: []regMask{gpMask}}
	// FIXME(prattmic): This is a hack to get things to build, but it probably
	// not correct.
	gp2 := regInfo{inputs: []regMask{gpMask, gpMask}, outputs: []regMask{gpMask}}

	RISCVops := []opData{
		{name: "ADD", argLength: 2, reg: gp2, asm: "ADD", commutative: true, resultInArg0: true}, // arg0 + arg1
		{name: "MOVmem", argLength: 1, reg: gp, asm: "MOV", aux: "SymOff"}, // arg0 + auxint + offset encoded in aux
		// auxint+aux == add auxint and the offset of the symbol in aux (if any) to the effective address
		{name: "MOVload", argLength: 2, reg: gp, asm: "MOV", aux: "SymOff"},  // load from arg0+auxint+aux. arg1=mem
		{name: "MOVstore", argLength: 3, reg: gp, asm: "MOV", aux: "SymOff", typ: "Mem"},  // store value in arg1 to arg0+auxint+aux. arg2=mem
		{name: "LoweredNilCheck", argLength: 2, reg: gp},  //arg0=ptr,arg1=mem, returns void.  Faults if ptr is nil.
	}

	RISCVblocks := []blockData{
		{name: "EQ"},
		{name: "NE"},
		{name: "LT"},
		{name: "GE"},
		{name: "LTU"},
		{name: "GEU"},
	}

	archs = append(archs, arch{
		name:     "RISCV",
		pkg:      "cmd/internal/obj/riscv",
		genfile:  "../../riscv/ssa.go",
		ops:      RISCVops,
		blocks:   RISCVblocks,
		regnames: regNamesRISCV,
	})
}
