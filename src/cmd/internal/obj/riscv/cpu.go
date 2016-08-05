//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2008 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2008 Lucent Technologies Inc. and others
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

import "cmd/internal/obj"

//go:generate go run ../stringer.go -i $GOFILE -o anames.go -p riscv

const (
	// Base register numberings.
	REG_X0 = obj.RBaseRISCV + iota
	REG_X1
	REG_X2
	REG_X3
	REG_X4
	REG_X5
	REG_X6
	REG_X7
	REG_X8
	REG_X9
	REG_X10
	REG_X11
	REG_X12
	REG_X13
	REG_X14
	REG_X15
	REG_X16
	REG_X17
	REG_X18
	REG_X19
	REG_X20
	REG_X21
	REG_X22
	REG_X23
	REG_X24
	REG_X25
	REG_X26
	REG_X27
	REG_X28
	REG_X29
	REG_X30
	REG_X31

	// FP register numberings.
	REG_F0
	REG_F1
	REG_F2
	REG_F3
	REG_F4
	REG_F5
	REG_F6
	REG_F7
	REG_F8
	REG_F9
	REG_F10
	REG_F11
	REG_F12
	REG_F13
	REG_F14
	REG_F15
	REG_F16
	REG_F17
	REG_F18
	REG_F19
	REG_F20
	REG_F21
	REG_F22
	REG_F23
	REG_F24
	REG_F25
	REG_F26
	REG_F27
	REG_F28
	REG_F29
	REG_F30
	REG_F31

	// Special/control registers.
	// TODO(myenik) Read more and add the ones we need...

	// This marks the end of the register numbering.
	REG_END

	// General registers reassigned to ABI names.
	REG_ZERO = REG_X0
	REG_RA   = REG_X1
	REG_SP   = REG_X2
	REG_GP   = REG_X3 // aka REG_SB
	REG_TP   = REG_X4 // aka REG_G
	REG_T0   = REG_X5
	REG_T1   = REG_X6
	REG_T2   = REG_X7
	REG_S0   = REG_X8
	REG_FP   = REG_X8 // S0 and FP are the same.
	REG_S1   = REG_X9
	REG_A0   = REG_X10
	REG_A1   = REG_X11
	REG_A2   = REG_X12
	REG_A3   = REG_X13
	REG_A4   = REG_X14
	REG_A5   = REG_X15
	REG_A6   = REG_X16
	REG_A7   = REG_X17
	REG_S2   = REG_X18
	REG_S3   = REG_X19
	REG_S4   = REG_X20
	REG_S5   = REG_X21
	REG_S6   = REG_X22
	REG_S7   = REG_X23
	REG_S8   = REG_X24
	REG_S9   = REG_X25
	REG_S10  = REG_X26
	REG_S11  = REG_X27
	REG_T3   = REG_X28
	REG_T4   = REG_X29
	REG_T5   = REG_X30
	REG_T6   = REG_X31

	// Go runtime register names.
	REG_SB   = REG_X3 // Static base.
	REG_G    = REG_X4 // G pointer.
	REG_RT1  = REG_S2 // Reserved for runtime (duffzero and duffcopy).
	REG_RT2  = REG_S3 // Reserved for runtime (duffcopy).
	REG_CTXT = REG_S4 // Context for closures.

	// ABI names for floating point registers.
	REG_FT0  = REG_F0
	REG_FT1  = REG_F1
	REG_FT2  = REG_F2
	REG_FT3  = REG_F3
	REG_FT4  = REG_F4
	REG_FT5  = REG_F5
	REG_FT6  = REG_F6
	REG_FT7  = REG_F7
	REG_FS0  = REG_F8
	REG_FS1  = REG_F9
	REG_FA0  = REG_F10
	REG_FA1  = REG_F11
	REG_FA2  = REG_F12
	REG_FA3  = REG_F13
	REG_FA4  = REG_F14
	REG_FA5  = REG_F15
	REG_FA6  = REG_F16
	REG_FA7  = REG_F17
	REG_FS2  = REG_F18
	REG_FS3  = REG_F19
	REG_FS4  = REG_F20
	REG_FS5  = REG_F21
	REG_FS6  = REG_F22
	REG_FS7  = REG_F23
	REG_FS8  = REG_F24
	REG_FS9  = REG_F25
	REG_FS10 = REG_F26
	REG_FS11 = REG_F27
	REG_FT8  = REG_F28
	REG_FT9  = REG_F29
	REG_FT10 = REG_F30
	REG_FT11 = REG_F31
)

// TEXTFLAG definitions.
const (
	/* mark flags */
	LABEL   = 1 << 0
	LEAF    = 1 << 1
	FLOAT   = 1 << 2
	BRANCH  = 1 << 3
	LOAD    = 1 << 4
	FCMP    = 1 << 5
	SYNC    = 1 << 6
	LIST    = 1 << 7
	FOLL    = 1 << 8
	NOSCHED = 1 << 9
)

// RISC-V mnemonics, as defined in the "opcodes" and "opcodes-pseudo" files of
// riscv-opcodes, as well as some fake mnemonics (e.g., MOV) used only in the
// assembler.
//
// If you modify this table, you MUST run 'go generate' to regenerate anames.go!
const (
	ABEQ = obj.ABaseRISCV + obj.A_ARCHSPECIFIC + iota
	ABNE
	ABLT
	ABGE
	ABLTU
	ABGEU
	AJALR
	AJAL
	ALUI
	AAUIPC
	AADDI
	ASLLI
	ASLTI
	ASLTIU
	AXORI
	ASRLI
	ASRAI
	AORI
	AANDI
	AADD
	ASUB
	ASLL
	ASLT
	ASLTU
	AXOR
	ASRL
	ASRA
	AOR
	AAND
	AADDIW
	ASLLIW
	ASRLIW
	ASRAIW
	AADDW
	ASUBW
	ASLLW
	ASRLW
	ASRAW
	ALB
	ALH
	ALW
	ALD
	ALBU
	ALHU
	ALWU
	ASB
	ASH
	ASW
	ASD
	AFENCE
	AFENCEI
	AMUL
	AMULH
	AMULHSU
	AMULHU
	ADIV
	ADIVU
	AREM
	AREMU
	AMULW
	ADIVW
	ADIVUW
	AREMW
	AREMUW
	AAMOADDW
	AAMOXORW
	AAMOORW
	AAMOANDW
	AAMOMINW
	AAMOMAXW
	AAMOMINUW
	AAMOMAXUW
	AAMOSWAPW
	ALRW
	ASCW
	AAMOADDD
	AAMOXORD
	AAMOORD
	AAMOANDD
	AAMOMIND
	AAMOMAXD
	AAMOMINUD
	AAMOMAXUD
	AAMOSWAPD
	ALRD
	ASCD
	ASCALL
	ASBREAK
	ASRET
	ASFENCEVM
	AWFI
	AMRTH
	AMRTS
	AHRTS
	ACSRRW
	ACSRRS
	ACSRRC
	ACSRRWI
	ACSRRSI
	ACSRRCI
	AFADDS
	AFSUBS
	AFMULS
	AFDIVS
	AFNEGS
	AFSGNJS
	AFSGNJNS
	AFSGNJXS
	AFMINS
	AFMAXS
	AFSQRTS
	AFADDD
	AFSUBD
	AFMULD
	AFDIVD
	AFSGNJD
	AFSGNJND
	AFSGNJXD
	AFMIND
	AFMAXD
	AFCVTSD
	AFCVTDS
	AFSQRTD
	AFLES
	AFLTS
	AFEQS
	AFLED
	AFLTD
	AFEQD
	AFCVTWS
	AFCVTWUS
	AFCVTLS
	AFCVTLUS
	AFMVXS
	AFCLASSS
	AFCVTWD
	AFCVTWUD
	AFCVTLD
	AFCVTLUD
	AFMVXD
	AFCLASSD
	AFCVTSW
	AFCVTSWU
	AFCVTSL
	AFCVTSLU
	AFMVSX
	AFCVTDW
	AFCVTDWU
	AFCVTDL
	AFCVTDLU
	AFMVDX
	AFLW
	AFLD
	AFSW
	AFSD
	AFMADDS
	AFMSUBS
	AFNMSUBS
	AFNMADDS
	AFMADDD
	AFMSUBD
	AFNMSUBD
	AFNMADDD
	ASLLIRV32
	ASRLIRV32
	ASRAIRV32
	AFRFLAGS
	AFSFLAGS
	AFSFLAGSI
	AFRRM
	AFSRM
	AFSRMI
	AFSCSR
	AFRCSR
	ARDCYCLE
	ARDTIME
	ARDINSTRET
	ARDCYCLEH
	ARDTIMEH
	ARDINSTRETH
	AECALL
	AEBREAK
	AERET

	// Fake instructions.  These get translated by the assembler into other
	// instructions, based on their operands.
	AMOV
	AMOVB
	AMOVH
	AMOVW
	AMOVBU
	AMOVHU
	AMOVWU

	ASEQZ
	ASNEZ
)

// All unary instructions which write to their arguments (as opposed to reading
// from them) go here.  The assembly parser uses this information to populate
// its AST in a semantically reasonable way.
//
// Any instructions not listed here is assumed to either be non-unary or to read
// from its argument.
var unaryDst = map[obj.As]bool{
	ARDCYCLE:    true,
	ARDCYCLEH:   true,
	ARDTIME:     true,
	ARDTIMEH:    true,
	ARDINSTRET:  true,
	ARDINSTRETH: true,
}

// Operands
const (
	// No operand.  This constant goes in any operand slot which is unused
	// (e.g., the two source register slots in RDCYCLE).
	C_NONE = iota

	// An integer register, either numbered (e.g., R12) or with an ABI name
	// (e.g., T2).
	C_REGI

	// An integer immediate.
	C_IMMI

	// A relative address.
	C_RELADDR

	// A memory address contained in an integer register.
	C_MEM

	// The size of a TEXT section.
	C_TEXTSIZE
)

// Instruction encoding masks
const (
	// UTypeImmMask is a mask including only the immediate portion of
	// U-type instructions.
	UTypeImmMask = 0xfffff000

	// UJTypeImmMask is a mask including only the immediate portion of
	// UJ-type instructions.
	UJTypeImmMask = UTypeImmMask
)
