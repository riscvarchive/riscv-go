// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

	RISCVops := []opData{
		{name: "ADD", argLength: 2, reg: gp, asm: "ADD", commutative: true, resultInArg0: true}, // arg0 + arg1
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
