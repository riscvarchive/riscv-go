// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import "cmd/internal/obj/riscv"

func init() {
	var regNamesRISCV []string
	var gpMask, fpMask, gpspMask, gpspsbMask regMask
	regNamed := make(map[string]regMask)

	// Build the list of register names, creating an appropriately indexed
	// regMask for the gp and fp registers as we go.
	addreg := func(r int) regMask {
		mask := regMask(1) << uint(len(regNamesRISCV))
		name := riscv.RegNames[int16(r)]
		regNamesRISCV = append(regNamesRISCV, name)
		regNamed[name] = mask
		return mask
	}
	for r := riscv.REG_X0; r <= riscv.REG_X31; r++ {
		mask := addreg(r)
		// Add general purpose registers to gpMask.
		switch r {
		// Special registers that we must leave alone.
		// TODO: Is this list right?
		case riscv.REG_ZERO, riscv.REG_RA, riscv.REG_G:
		case riscv.REG_SB:
			gpspsbMask |= mask
		case riscv.REG_SP:
			gpspMask |= mask
			gpspsbMask |= mask
		default:
			gpMask |= mask
			gpspMask |= mask
			gpspsbMask |= mask
		}
	}
	for r := riscv.REG_F0; r <= riscv.REG_F31; r++ {
		mask := addreg(r)
		fpMask |= mask
	}

	if len(regNamesRISCV) > 64 {
		// regMask is only 64 bits.
		panic("Too many RISCV registers")
	}

	var (
		gpstore = regInfo{inputs: []regMask{gpspsbMask, gpspMask, 0}} // SB in first input so we can load from a global, but not in second to avoid using SB as a temporary register
		gp01    = regInfo{outputs: []regMask{gpMask}}
		// FIXME(prattmic): This is a hack to get things to build, but it probably
		// not correct.
		gp21   = regInfo{inputs: []regMask{gpMask, gpMask}, outputs: []regMask{gpMask}}
		gpload = regInfo{inputs: []regMask{gpspsbMask, 0}, outputs: []regMask{gpMask}}
		gp11sb = regInfo{inputs: []regMask{gpspsbMask}, outputs: []regMask{gpMask}}
	)

	RISCVops := []opData{
		{name: "ADD", argLength: 2, reg: gp21, asm: "ADD", commutative: true, resultInArg0: true}, // arg0 + arg1
		{name: "MOVmem", argLength: 1, reg: gp11sb, asm: "MOV", aux: "SymOff"},                    // arg0 + auxint + offset encoded in aux
		// auxint+aux == add auxint and the offset of the symbol in aux (if any) to the effective address

		{name: "MOVBconst", reg: gp01, asm: "MOV", typ: "UInt8", aux: "Int8", rematerializeable: true},   // 8 low bits of auxint
		{name: "MOVWconst", reg: gp01, asm: "MOV", typ: "UInt16", aux: "Int16", rematerializeable: true}, // 16 low bits of auxint
		{name: "MOVLconst", reg: gp01, asm: "MOV", typ: "UInt32", aux: "Int32", rematerializeable: true}, // 32 low bits of auxint
		{name: "MOVQconst", reg: gp01, asm: "MOV", typ: "UInt64", aux: "Int64", rematerializeable: true}, // auxint

		{name: "MOVload", argLength: 2, reg: gpload, asm: "MOV", aux: "SymOff"},               // load from arg0+auxint+aux. arg1=mem
		{name: "MOVstore", argLength: 3, reg: gpstore, asm: "MOV", aux: "SymOff", typ: "Mem"}, // store value in arg1 to arg0+auxint+aux. arg2=mem
		{name: "LoweredNilCheck", argLength: 2, reg: regInfo{inputs: []regMask{gpspMask}}},    // arg0=ptr,arg1=mem, returns void.  Faults if ptr is nil.

		{name: "LoweredExitProc", argLength: 2, typ: "Mem", reg: regInfo{inputs: []regMask{gpMask, 0}, clobbers: regNamed[".A0"] | regNamed[".A7"]}}, // arg0=mem, auxint=return code
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
		name:            "RISCV",
		pkg:             "cmd/internal/obj/riscv",
		genfile:         "../../riscv/ssa.go",
		ops:             RISCVops,
		blocks:          RISCVblocks,
		regnames:        regNamesRISCV,
		gpregmask:       gpMask,
		fpregmask:       fpMask,
		flagmask:        0,  // no flags
		framepointerreg: -1, // not used
	})
}
