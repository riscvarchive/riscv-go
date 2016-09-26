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
	//
	// If name is specified, use it rather than the riscv reg number.
	addreg := func(r int, name string) regMask {
		mask := regMask(1) << uint(len(regNamesRISCV))
		if name == "" {
			name = riscv.RegNames[int16(r)]
		}
		regNamesRISCV = append(regNamesRISCV, name)
		regNamed[name] = mask
		return mask
	}

	// General purpose registers.
	for r := riscv.REG_X0; r <= riscv.REG_X31; r++ {
		if r == riscv.REG_RA {
			// RA is not used by regalloc, so we skip it to leave
			// room for pseudo-register SB.
			continue
		}

		mask := addreg(r, "")

		// Add general purpose registers to gpMask.
		switch r {
		// ZERO, g, and TMP are not in any gp mask.
		case riscv.REG_ZERO, riscv.REG_G, riscv.REG_TMP:
		case riscv.REG_SP:
			gpspMask |= mask
			gpspsbMask |= mask
		default:
			gpMask |= mask
			gpspMask |= mask
			gpspsbMask |= mask
		}
	}

	// Floating pointer registers.
	for r := riscv.REG_F0; r <= riscv.REG_F31; r++ {
		mask := addreg(r, "")
		fpMask |= mask
	}

	// Pseudo-register: SB
	mask := addreg(-1, "SB")
	gpspsbMask |= mask

	if len(regNamesRISCV) > 64 {
		// regMask is only 64 bits.
		panic("Too many RISCV registers")
	}

	regCtxt := regNamed["CTXT"]
	callerSave := gpMask | fpMask | regNamed["g"]

	var (
		gpstore = regInfo{inputs: []regMask{gpspsbMask, gpspMask, 0}} // SB in first input so we can load from a global, but not in second to avoid using SB as a temporary register
		gp01    = regInfo{outputs: []regMask{gpMask}}
		// FIXME(prattmic): This is a hack to get things to build, but it probably
		// not correct.
		gp11   = regInfo{inputs: []regMask{gpMask}, outputs: []regMask{gpMask}}
		gp21   = regInfo{inputs: []regMask{gpMask, gpMask}, outputs: []regMask{gpMask}}
		gp20   = regInfo{inputs: []regMask{gpMask, gpMask}, outputs: []regMask{}}
		gpload = regInfo{inputs: []regMask{gpspsbMask, 0}, outputs: []regMask{gpMask}}
		gp11sb = regInfo{inputs: []regMask{gpspsbMask}, outputs: []regMask{gpMask}}

		fp11    = regInfo{inputs: []regMask{fpMask}, outputs: []regMask{fpMask}}
		fp21    = regInfo{inputs: []regMask{fpMask, fpMask}, outputs: []regMask{fpMask}}
		gpfp    = regInfo{inputs: []regMask{gpMask}, outputs: []regMask{fpMask}}
		fpgp    = regInfo{inputs: []regMask{fpMask}, outputs: []regMask{gpMask}}
		fpstore = regInfo{inputs: []regMask{gpspsbMask, fpMask, 0}}
		fpload  = regInfo{inputs: []regMask{gpspsbMask, 0}, outputs: []regMask{fpMask}}
		fp2gp   = regInfo{inputs: []regMask{fpMask, fpMask}, outputs: []regMask{gpMask}}

		call        = regInfo{clobbers: callerSave}
		callClosure = regInfo{inputs: []regMask{gpspMask, regCtxt, 0}, clobbers: callerSave}
		callInter   = regInfo{inputs: []regMask{gpMask}, clobbers: callerSave}
	)

	RISCVops := []opData{
		{name: "ADD", argLength: 2, reg: gp21, asm: "ADD", commutative: true}, // arg0 + arg1
		{name: "ADDI", argLength: 1, reg: gp11sb, asm: "ADDI", aux: "Int64"},  // arg0 + auxint
		{name: "SUB", argLength: 2, reg: gp21, asm: "SUB"},                    // arg0 - arg1

		// M extension. H means high (i.e., it returns the top bits of
		// the result). U means unsigned. W means word (i.e., 32-bit).
		{name: "MUL", argLength: 2, reg: gp21, asm: "MUL", commutative: true, typ: "Int64"}, // arg0 * arg1
		{name: "MULW", argLength: 2, reg: gp21, asm: "MULW", commutative: true, typ: "Int32"},
		{name: "MULH", argLength: 2, reg: gp21, asm: "MULH", commutative: true, typ: "Int64"},
		{name: "MULHU", argLength: 2, reg: gp21, asm: "MULHU", commutative: true, typ: "UInt64"},
		{name: "DIV", argLength: 2, reg: gp21, asm: "DIV", typ: "Int64"}, // arg0 / arg1
		{name: "DIVU", argLength: 2, reg: gp21, asm: "DIVU", typ: "UInt64"},
		{name: "DIVW", argLength: 2, reg: gp21, asm: "DIVW", typ: "Int32"},
		{name: "DIVUW", argLength: 2, reg: gp21, asm: "DIVUW", typ: "UInt32"},
		{name: "REM", argLength: 2, reg: gp21, asm: "REM", typ: "Int64"}, // arg0 % arg1
		{name: "REMU", argLength: 2, reg: gp21, asm: "REMU", typ: "UInt64"},
		{name: "REMW", argLength: 2, reg: gp21, asm: "REMW", typ: "Int32"},
		{name: "REMUW", argLength: 2, reg: gp21, asm: "REMUW", typ: "UInt32"},

		{name: "MOVaddr", argLength: 1, reg: gp11sb, asm: "MOV", aux: "SymOff"}, // arg0 + auxint + offset encoded in aux
		// auxint+aux == add auxint and the offset of the symbol in aux (if any) to the effective address

		{name: "MOVBconst", reg: gp01, asm: "MOV", typ: "UInt8", aux: "Int8", rematerializeable: true},    // 8 low bits of auxint
		{name: "MOVHconst", reg: gp01, asm: "MOV", typ: "UInt16", aux: "Int16", rematerializeable: true},  // 16 low bits of auxint
		{name: "MOVWconst", reg: gp01, asm: "MOV", typ: "UInt32", aux: "Int32", rematerializeable: true},  // 32 low bits of auxint
		{name: "MOVDconst", reg: gp01, asm: "MOV", typ: "UInt64", aux: "Int64", rematerializeable: true},  // auxint
		{name: "MOVSconst", reg: gp01, asm: "MOV", typ: "Float32", aux: "Int32", rematerializeable: true}, // auxint as float

		// Loads: load <size> bits from arg0+auxint+aux and extend to 64 bits; arg1=mem
		{name: "MOVBload", argLength: 2, reg: gpload, asm: "MOVB", aux: "SymOff", typ: "Int8"},     //  8 bits, sign extend
		{name: "MOVHload", argLength: 2, reg: gpload, asm: "MOVH", aux: "SymOff", typ: "Int16"},    // 16 bits, sign extend
		{name: "MOVWload", argLength: 2, reg: gpload, asm: "MOVW", aux: "SymOff", typ: "Int32"},    // 32 bits, sign extend
		{name: "MOVDload", argLength: 2, reg: gpload, asm: "MOV", aux: "SymOff", typ: "Int64"},     // 64 bits
		{name: "MOVBUload", argLength: 2, reg: gpload, asm: "MOVBU", aux: "SymOff", typ: "UInt8"},  //  8 bits, zero extend
		{name: "MOVHUload", argLength: 2, reg: gpload, asm: "MOVHU", aux: "SymOff", typ: "UInt16"}, // 16 bits, zero extend
		{name: "MOVWUload", argLength: 2, reg: gpload, asm: "MOVWU", aux: "SymOff", typ: "UInt32"}, // 32 bits, zero extend

		// Stores: store <size> lowest bits in arg1 to arg0+auxint+aux; arg2=mem
		{name: "MOVBstore", argLength: 3, reg: gpstore, asm: "MOVB", aux: "SymOff", typ: "Mem"}, //  8 bits
		{name: "MOVHstore", argLength: 3, reg: gpstore, asm: "MOVH", aux: "SymOff", typ: "Mem"}, // 16 bits
		{name: "MOVWstore", argLength: 3, reg: gpstore, asm: "MOVW", aux: "SymOff", typ: "Mem"}, // 32 bits
		{name: "MOVDstore", argLength: 3, reg: gpstore, asm: "MOV", aux: "SymOff", typ: "Mem"},  // 64 bits

		// Shift ops
		{name: "SLL", argLength: 2, reg: gp21, asm: "SLL"},                 // arg0 << aux1
		{name: "SRA", argLength: 2, reg: gp21, asm: "SRA"},                 // arg0 >> aux1, signed
		{name: "SRL", argLength: 2, reg: gp21, asm: "SRL"},                 // arg0 >> aux1, unsigned
		{name: "SLLI", argLength: 1, reg: gp11, asm: "SLLI", aux: "Int64"}, // arg0 << auxint
		{name: "SRAI", argLength: 1, reg: gp11, asm: "SRAI", aux: "Int64"}, // arg0 >> auxint, signed
		{name: "SRLI", argLength: 1, reg: gp11, asm: "SRLI", aux: "Int64"}, // arg0 >> auxint, unsigned

		// Bitwise ops
		{name: "XOR", argLength: 2, reg: gp21, asm: "XOR", commutative: true}, // arg0 ^ arg1
		{name: "XORI", argLength: 1, reg: gp11, asm: "XORI", aux: "Int64"},    // arg0 ^ auxint
		{name: "OR", argLength: 2, reg: gp21, asm: "OR", commutative: true},   // arg0 | arg1
		{name: "ORI", argLength: 1, reg: gp11, asm: "ORI", aux: "Int64"},      // arg0 | auxint
		{name: "AND", argLength: 2, reg: gp21, asm: "AND", commutative: true}, // arg0 & arg1
		{name: "ANDI", argLength: 1, reg: gp11, asm: "ANDI", aux: "Int64"},    // arg0 & auxint

		// Generate boolean values
		{name: "SEQZ", argLength: 1, reg: gp11, asm: "SEQZ"},                 // arg0 == 0, result is 0 or 1
		{name: "SNEZ", argLength: 1, reg: gp11, asm: "SNEZ"},                 // arg0 != 0, result is 0 or 1
		{name: "SLT", argLength: 2, reg: gp21, asm: "SLT"},                   // arg0 < arg1, result is 0 or 1
		{name: "SLTI", argLength: 1, reg: gp11, asm: "SLTI", aux: "Int64"},   // arg0 < auxint, result is 0 or 1
		{name: "SLTU", argLength: 2, reg: gp21, asm: "SLTU"},                 // arg0 < arg1, unsigned, result is 0 or 1
		{name: "SLTIU", argLength: 1, reg: gp11, asm: "SLTIU", aux: "Int64"}, // arg0 < auxint, unsigned, result is 0 or 1

		// Flag pseudo-ops.
		// RISC-V doesn't have flags, but SSA wants branches to be flag-based.
		// These are pseudo-ops that contain the arguments to the comparison.
		// There is a single branching block type, BRANCH,
		// which accepts one of these values as its control.
		// During code generation we consult the control value
		// to emit the correct conditional branch instruction.
		{name: "BEQ", argLength: 2, reg: gp20, asm: "BEQ", typ: "Flags"},   // branch if arg0 == arg1
		{name: "BNE", argLength: 2, reg: gp20, asm: "BNE", typ: "Flags"},   // branch if arg0 != arg1
		{name: "BLT", argLength: 2, reg: gp20, asm: "BLT", typ: "Flags"},   // branch if arg0 < arg1
		{name: "BLTU", argLength: 2, reg: gp20, asm: "BLTU", typ: "Flags"}, // branch if arg0 < arg1, unsigned
		{name: "BGE", argLength: 2, reg: gp20, asm: "BGE", typ: "Flags"},   // branch if arg0 >= arg1
		{name: "BGEU", argLength: 2, reg: gp20, asm: "BGEU", typ: "Flags"}, // branch if arg0 >= arg1, unsigned

		// MOVconvert converts between pointers and integers.
		// We have a special op for this so as to not confuse GC
		// (particularly stack maps). It takes a memory arg so it
		// gets correctly ordered with respect to GC safepoints.
		{name: "MOVconvert", argLength: 2, reg: gp11, asm: "MOV"}, // arg0, but converted to int/ptr as appropriate; arg1=mem

		// Calls
		{name: "CALLstatic", argLength: 1, reg: call, aux: "SymOff"},        // call static function aux.(*gc.Sym). arg0=mem, auxint=argsize, returns mem
		{name: "CALLclosure", argLength: 3, reg: callClosure, aux: "Int64"}, // call function via closure. arg0=codeptr, arg1=closure, arg2=mem, auxint=argsize, returns mem
		{name: "CALLdefer", argLength: 1, reg: call, aux: "Int64"},          // call deferproc. arg0=mem, auxint=argsize, returns mem
		{name: "CALLgo", argLength: 1, reg: call, aux: "Int64"},             // call newproc. arg0=mem, auxint=argsize, returns mem
		{name: "CALLinter", argLength: 2, reg: callInter, aux: "Int64"},     // call fn by pointer. arg0=codeptr, arg1=mem, auxint=argsize, returns mem

		// Lowering pass-throughs
		{name: "LoweredNilCheck", argLength: 2, reg: regInfo{inputs: []regMask{gpspMask}}},                                                         // arg0=ptr,arg1=mem, returns void.  Faults if ptr is nil.
		{name: "LoweredGetClosurePtr", reg: regInfo{outputs: []regMask{regCtxt}}},                                                                  // scheduler ensures only at beginning of entry block
		{name: "LoweredExitProc", argLength: 2, typ: "Mem", reg: regInfo{inputs: []regMask{gpMask, 0}, clobbers: regNamed["A0"] | regNamed["A7"]}}, // arg0=mem, auxint=return code

		// F extension.
		{name: "FADDS", argLength: 2, reg: fp21, asm: "FADDS", commutative: true, typ: "Float32"},  // arg0 + arg1
		{name: "FSUBS", argLength: 2, reg: fp21, asm: "FSUBS", commutative: false, typ: "Float32"}, // arg0 - arg1
		{name: "FMULS", argLength: 2, reg: fp21, asm: "FMULS", commutative: true, typ: "Float32"},  // arg0 * arg1
		{name: "FDIVS", argLength: 2, reg: fp21, asm: "FDIVS", commutative: false, typ: "Float32"}, // arg0 / arg1
		{name: "FSQRTS", argLength: 1, reg: fp11, asm: "FSQRTS", typ: "Float32"},                   // sqrt(arg0)
		{name: "FNEGS", argLength: 1, reg: fp11, asm: "FNEGS", typ: "Float32"},                     // -arg0
		{name: "FMVSX", argLength: 1, reg: gpfp, asm: "FMVSX", typ: "Float32"},                     // reinterpret arg0 as float
		{name: "FCVTSW", argLength: 1, reg: gpfp, asm: "FCVTSW", typ: "Float32"},                   // float32(arg0)
		{name: "FCVTSL", argLength: 1, reg: gpfp, asm: "FCVTSL", typ: "Float32"},                   // float32(arg0)
		{name: "FCVTWS", argLength: 1, reg: fpgp, asm: "FCVTWS", typ: "Int32"},                     // int32(arg0)
		{name: "FCVTLS", argLength: 1, reg: fpgp, asm: "FCVTLS", typ: "Int64"},                     // int64(arg0)
		{name: "FMOVWload", argLength: 2, reg: fpload, asm: "MOVF", aux: "SymOff", typ: "Float32"}, // load float32 from arg0+auxint+aux
		{name: "FMOVWstore", argLength: 3, reg: fpstore, asm: "MOVF", aux: "SymOff", typ: "Mem"},   // store float32 to arg0+auxint+aux
		{name: "FEQS", argLength: 2, reg: fp2gp, asm: "FEQS", commutative: true},                   // arg0 == arg1
		{name: "FNES", argLength: 2, reg: fp2gp, asm: "FNES", commutative: true},                   // arg0 != arg1
		{name: "FLTS", argLength: 2, reg: fp2gp, asm: "FLTS"},                                      // arg0 < arg1
		{name: "FLES", argLength: 2, reg: fp2gp, asm: "FLES"},                                      // arg0 <= arg1

		// D extension.
		{name: "FADDD", argLength: 2, reg: fp21, asm: "FADDD", commutative: true, typ: "Float64"},  // arg0 + arg1
		{name: "FSUBD", argLength: 2, reg: fp21, asm: "FSUBD", commutative: false, typ: "Float64"}, // arg0 - arg1
		{name: "FMULD", argLength: 2, reg: fp21, asm: "FMULD", commutative: true, typ: "Float64"},  // arg0 * arg1
		{name: "FDIVD", argLength: 2, reg: fp21, asm: "FDIVD", commutative: false, typ: "Float64"}, // arg0 / arg1
		{name: "FSQRTD", argLength: 1, reg: fp11, asm: "FSQRTD", typ: "Float64"},                   // sqrt(arg0)
		{name: "FNEGD", argLength: 1, reg: fp11, asm: "FNEGD", typ: "Float64"},                     // -arg0
		{name: "FMVDX", argLength: 1, reg: gpfp, asm: "FMVDX", typ: "Float64"},                     // reinterpret arg0 as float
		{name: "FCVTDW", argLength: 1, reg: gpfp, asm: "FCVTDW", typ: "Float64"},                   // float64(arg0)
		{name: "FCVTDL", argLength: 1, reg: gpfp, asm: "FCVTDL", typ: "Float64"},                   // float64(arg0)
		{name: "FCVTWD", argLength: 1, reg: fpgp, asm: "FCVTWD", typ: "Int32"},                     // int32(arg0)
		{name: "FCVTLD", argLength: 1, reg: fpgp, asm: "FCVTLD", typ: "Int64"},                     // int64(arg0)
		{name: "FCVTDS", argLength: 1, reg: fp11, asm: "FCVTDS", typ: "Float64"},                   // float64(arg0)
		{name: "FCVTSD", argLength: 1, reg: fp11, asm: "FCVTSD", typ: "Float32"},                   // float32(arg0)
		{name: "FMOVDload", argLength: 2, reg: fpload, asm: "MOVD", aux: "SymOff", typ: "Float64"}, // load float64 from arg0+auxint+aux
		{name: "FMOVDstore", argLength: 3, reg: fpstore, asm: "MOVD", aux: "SymOff", typ: "Mem"},   // store float6 to arg0+auxint+aux
		{name: "FEQD", argLength: 2, reg: fp2gp, asm: "FEQD", commutative: true},                   // arg0 == arg1
		{name: "FNED", argLength: 2, reg: fp2gp, asm: "FNED", commutative: true},                   // arg0 != arg1
		{name: "FLTD", argLength: 2, reg: fp2gp, asm: "FLTD"},                                      // arg0 < arg1
		{name: "FLED", argLength: 2, reg: fp2gp, asm: "FLED"},                                      // arg0 <= arg1
	}

	RISCVblocks := []blockData{
		{name: "BRANCH"}, // see flag pseudo-ops above.
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
		framepointerreg: -1, // not used
	})
}
