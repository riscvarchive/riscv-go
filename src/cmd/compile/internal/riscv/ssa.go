// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"math"

	"cmd/compile/internal/gc"
	"cmd/compile/internal/ssa"
	"cmd/internal/obj"
	"cmd/internal/obj/riscv"
)

// ssaRegToReg maps ssa register numbers to obj register numbers.
var ssaRegToReg = []int16{
	riscv.REG_X0,
	// X1 (RA): unused
	riscv.REG_X2,
	riscv.REG_X3,
	riscv.REG_X4,
	riscv.REG_X5,
	riscv.REG_X6,
	riscv.REG_X7,
	riscv.REG_X8,
	riscv.REG_X9,
	riscv.REG_X10,
	riscv.REG_X11,
	riscv.REG_X12,
	riscv.REG_X13,
	riscv.REG_X14,
	riscv.REG_X15,
	riscv.REG_X16,
	riscv.REG_X17,
	riscv.REG_X18,
	riscv.REG_X19,
	riscv.REG_X20,
	riscv.REG_X21,
	riscv.REG_X22,
	riscv.REG_X23,
	riscv.REG_X24,
	riscv.REG_X25,
	riscv.REG_X26,
	riscv.REG_X27,
	riscv.REG_X28,
	riscv.REG_X29,
	riscv.REG_X30,
	riscv.REG_X31,
	riscv.REG_F0,
	riscv.REG_F1,
	riscv.REG_F2,
	riscv.REG_F3,
	riscv.REG_F4,
	riscv.REG_F5,
	riscv.REG_F6,
	riscv.REG_F7,
	riscv.REG_F8,
	riscv.REG_F9,
	riscv.REG_F10,
	riscv.REG_F11,
	riscv.REG_F12,
	riscv.REG_F13,
	riscv.REG_F14,
	riscv.REG_F15,
	riscv.REG_F16,
	riscv.REG_F17,
	riscv.REG_F18,
	riscv.REG_F19,
	riscv.REG_F20,
	riscv.REG_F21,
	riscv.REG_F22,
	riscv.REG_F23,
	riscv.REG_F24,
	riscv.REG_F25,
	riscv.REG_F26,
	riscv.REG_F27,
	riscv.REG_F28,
	riscv.REG_F29,
	riscv.REG_F30,
	riscv.REG_F31,
	0, // SB isn't a real register.  We fill an Addr.Reg field with 0 in this case.
}

func loadByType(t ssa.Type) obj.As {
	width := t.Size()

	if t.IsFloat() {
		switch width {
		case 4:
			return riscv.AMOVF
		case 8:
			return riscv.AMOVD
		default:
			gc.Fatalf("unknown float width for load %d in type %v", width, t)
			return 0
		}
	}

	switch width {
	case 1:
		if t.IsSigned() {
			return riscv.AMOVB
		} else {
			return riscv.AMOVBU
		}
	case 2:
		if t.IsSigned() {
			return riscv.AMOVH
		} else {
			return riscv.AMOVHU
		}
	case 4:
		if t.IsSigned() {
			return riscv.AMOVW
		} else {
			return riscv.AMOVWU
		}
	case 8:
		return riscv.AMOV
	default:
		gc.Fatalf("unknown width for load %d in type %v", width, t)
		return 0
	}
}

// storeByType returns the store instruction of the given type.
func storeByType(t ssa.Type) obj.As {
	width := t.Size()

	if t.IsFloat() {
		switch width {
		case 4:
			return riscv.AMOVF
		case 8:
			return riscv.AMOVD
		default:
			gc.Fatalf("unknown float width for store %d in type %v", width, t)
			return 0
		}
	}

	switch width {
	case 1:
		return riscv.AMOVB
	case 2:
		return riscv.AMOVH
	case 4:
		return riscv.AMOVW
	case 8:
		return riscv.AMOV
	default:
		gc.Fatalf("unknown width for store %d in type %v", width, t)
		return 0
	}
}

// largestMove returns the largest move instruction possible and its size,
// given the alignment of the total size of the move.
//
// e.g., a 16-byte move may use MOV, but an 11-byte move must use MOVB.
//
// Note that the moves may not be on naturally aligned addresses depending on
// the source and destination.
//
// This matches the calculation in ssa.moveSize.
func largestMove(alignment int64) (obj.As, int64) {
	switch {
	case alignment%8 == 0:
		return riscv.AMOV, 8
	case alignment%4 == 0:
		return riscv.AMOVW, 4
	case alignment%2 == 0:
		return riscv.AMOVH, 2
	default:
		return riscv.AMOVB, 1
	}
}

// markMoves marks any MOVXconst ops that need to avoid clobbering flags.
// RISC-V has no flags, so this is a no-op.
func ssaMarkMoves(s *gc.SSAGenState, b *ssa.Block) {}

func ssaGenValue(s *gc.SSAGenState, v *ssa.Value) {
	s.SetLineno(v.Line)

	switch v.Op {
	case ssa.OpInitMem:
		// memory arg needs no code
	case ssa.OpArg:
		// input args need no code
	case ssa.OpPhi:
		gc.CheckLoweredPhi(v)
	case ssa.OpCopy, ssa.OpRISCVMOVconvert:
		if v.Type.IsMemory() {
			return
		}
		rs := v.Args[0].Reg()
		rd := v.Reg()
		if rs == rd {
			return
		}
		as := riscv.AMOV
		if v.Type.IsFloat() {
			as = riscv.AMOVD
		}
		p := gc.Prog(as)
		p.From.Type = obj.TYPE_REG
		p.From.Reg = rs
		p.To.Type = obj.TYPE_REG
		p.To.Reg = rd
	case ssa.OpLoadReg:
		if v.Type.IsFlags() {
			v.Fatalf("load flags not implemented: %v", v.LongString())
			return
		}
		p := gc.Prog(loadByType(v.Type))
		gc.AddrAuto(&p.From, v.Args[0])
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()
	case ssa.OpStoreReg:
		if v.Type.IsFlags() {
			v.Fatalf("store flags not implemented: %v", v.LongString())
			return
		}
		p := gc.Prog(storeByType(v.Type))
		p.From.Type = obj.TYPE_REG
		p.From.Reg = v.Args[0].Reg()
		gc.AddrAuto(&p.To, v)
	case ssa.OpVarDef:
		gc.Gvardef(v.Aux.(*gc.Node))
	case ssa.OpVarKill:
		gc.Gvarkill(v.Aux.(*gc.Node))
	case ssa.OpVarLive:
		gc.Gvarlive(v.Aux.(*gc.Node))
	case ssa.OpKeepAlive:
		gc.KeepAlive(v)
	case ssa.OpSP, ssa.OpSB, ssa.OpGetG:
		// nothing to do
	case ssa.OpRISCVADD, ssa.OpRISCVSUB, ssa.OpRISCVXOR, ssa.OpRISCVOR, ssa.OpRISCVAND,
		ssa.OpRISCVSLL, ssa.OpRISCVSRA, ssa.OpRISCVSRL,
		ssa.OpRISCVSLT, ssa.OpRISCVSLTU, ssa.OpRISCVMUL, ssa.OpRISCVMULW, ssa.OpRISCVMULH,
		ssa.OpRISCVMULHU, ssa.OpRISCVDIV, ssa.OpRISCVDIVU, ssa.OpRISCVDIVW,
		ssa.OpRISCVDIVUW, ssa.OpRISCVREM, ssa.OpRISCVREMU, ssa.OpRISCVREMW,
		ssa.OpRISCVREMUW,
		ssa.OpRISCVFADDS, ssa.OpRISCVFSUBS, ssa.OpRISCVFMULS, ssa.OpRISCVFDIVS,
		ssa.OpRISCVFEQS, ssa.OpRISCVFNES, ssa.OpRISCVFLTS, ssa.OpRISCVFLES,
		ssa.OpRISCVFADDD, ssa.OpRISCVFSUBD, ssa.OpRISCVFMULD, ssa.OpRISCVFDIVD,
		ssa.OpRISCVFEQD, ssa.OpRISCVFNED, ssa.OpRISCVFLTD, ssa.OpRISCVFLED:
		r := v.Reg()
		r1 := v.Args[0].Reg()
		r2 := v.Args[1].Reg()
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = r2
		p.From3 = &obj.Addr{
			Type: obj.TYPE_REG,
			Reg:  r1,
		}
		p.To.Type = obj.TYPE_REG
		p.To.Reg = r
	case ssa.OpRISCVFSQRTS, ssa.OpRISCVFNEGS, ssa.OpRISCVFSQRTD, ssa.OpRISCVFNEGD,
		ssa.OpRISCVFMVSX, ssa.OpRISCVFMVDX,
		ssa.OpRISCVFCVTSW, ssa.OpRISCVFCVTSL, ssa.OpRISCVFCVTWS, ssa.OpRISCVFCVTLS,
		ssa.OpRISCVFCVTDW, ssa.OpRISCVFCVTDL, ssa.OpRISCVFCVTWD, ssa.OpRISCVFCVTLD, ssa.OpRISCVFCVTDS, ssa.OpRISCVFCVTSD:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = v.Args[0].Reg()
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()
	case ssa.OpRISCVADDI, ssa.OpRISCVXORI, ssa.OpRISCVORI, ssa.OpRISCVANDI,
		ssa.OpRISCVSLLI, ssa.OpRISCVSRAI, ssa.OpRISCVSRLI, ssa.OpRISCVSLTI,
		ssa.OpRISCVSLTIU:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = v.AuxInt
		p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[0].Reg()}
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()
	case ssa.OpRISCVMOVBconst, ssa.OpRISCVMOVHconst, ssa.OpRISCVMOVWconst, ssa.OpRISCVMOVDconst:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = v.AuxInt
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()
	case ssa.OpRISCVMOVSconst:
		p := gc.Prog(v.Op.Asm())
		// Convert the float to the equivalent integer literal so we can
		// move it using existing infrastructure.
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(int32(math.Float32bits(float32(math.Float64frombits(uint64(v.AuxInt))))))
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()
	case ssa.OpRISCVMOVaddr:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_ADDR
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()

		var wantreg string
		// MOVW $sym+off(base), R
		switch v.Aux.(type) {
		default:
			v.Fatalf("aux is of unknown type %T", v.Aux)
		case *ssa.ExternSymbol:
			wantreg = "SB"
			gc.AddAux(&p.From, v)
		case *ssa.ArgSymbol, *ssa.AutoSymbol:
			wantreg = "SP"
			gc.AddAux(&p.From, v)
		case nil:
			// No sym, just MOVW $off(SP), R
			wantreg = "SP"
			p.From.Reg = riscv.REG_SP
			p.From.Offset = v.AuxInt
		}
		if reg := v.Args[0].RegName(); reg != wantreg {
			v.Fatalf("bad reg %s for symbol type %T, want %s", reg, v.Aux, wantreg)
		}
	case ssa.OpRISCVMOVBload, ssa.OpRISCVMOVHload, ssa.OpRISCVMOVWload, ssa.OpRISCVMOVDload,
		ssa.OpRISCVMOVBUload, ssa.OpRISCVMOVHUload, ssa.OpRISCVMOVWUload,
		ssa.OpRISCVFMOVWload, ssa.OpRISCVFMOVDload:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = v.Args[0].Reg()
		gc.AddAux(&p.From, v)
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()
	case ssa.OpRISCVMOVBstore, ssa.OpRISCVMOVHstore, ssa.OpRISCVMOVWstore, ssa.OpRISCVMOVDstore,
		ssa.OpRISCVFMOVWstore, ssa.OpRISCVFMOVDstore:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = v.Args[1].Reg()
		p.To.Type = obj.TYPE_MEM
		p.To.Reg = v.Args[0].Reg()
		gc.AddAux(&p.To, v)
	case ssa.OpRISCVSEQZ, ssa.OpRISCVSNEZ:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = v.Args[0].Reg()
		p.To.Type = obj.TYPE_REG
		p.To.Reg = v.Reg()
	case ssa.OpRISCVCALLstatic, ssa.OpRISCVCALLclosure, ssa.OpRISCVCALLdefer, ssa.OpRISCVCALLgo, ssa.OpRISCVCALLinter:
		if v.Op == ssa.OpRISCVCALLstatic && v.Aux.(*gc.Sym) == gc.Deferreturn.Sym {
			// Deferred calls will appear to be returning to
			// the CALL deferreturn(SB) that we are about to emit.
			// However, the stack trace code will show the line
			// of the instruction byte before the return PC.
			// To avoid that being an unrelated instruction,
			// insert an actual hardware NOP that will have the right line number.
			// This is different from obj.ANOP, which is a virtual no-op
			// that doesn't make it into the instruction stream.
			ginsnop()
		}
		p := gc.Prog(obj.ACALL)
		p.To.Type = obj.TYPE_MEM
		switch v.Op {
		case ssa.OpRISCVCALLstatic:
			p.To.Name = obj.NAME_EXTERN
			p.To.Sym = gc.Linksym(v.Aux.(*gc.Sym))
		case ssa.OpRISCVCALLdefer:
			p.To.Name = obj.NAME_EXTERN
			p.To.Sym = gc.Linksym(gc.Deferproc.Sym)
		case ssa.OpRISCVCALLgo:
			p.To.Name = obj.NAME_EXTERN
			p.To.Sym = gc.Linksym(gc.Newproc.Sym)
		case ssa.OpRISCVCALLclosure, ssa.OpRISCVCALLinter:
			p.To.Type = obj.TYPE_REG
			p.To.Reg = v.Args[0].Reg()
		}
		if gc.Maxarg < v.AuxInt {
			gc.Maxarg = v.AuxInt
		}

	case ssa.OpRISCVLoweredZero:
		mov, sz := largestMove(v.AuxInt)

		//	mov	ZERO, (Rarg0)
		//	ADD	$sz, Rarg0
		//	BNE	Rarg1, Rarg0, -2(PC)

		p := gc.Prog(mov)
		p.From.Type = obj.TYPE_REG
		p.From.Reg = riscv.REG_ZERO
		p.To.Type = obj.TYPE_MEM
		p.To.Reg = v.Args[0].Reg()

		p2 := gc.Prog(riscv.AADD)
		p2.From.Type = obj.TYPE_CONST
		p2.From.Offset = sz
		p2.To.Type = obj.TYPE_REG
		p2.To.Reg = v.Args[0].Reg()

		p3 := gc.Prog(riscv.ABNE)
		p3.To.Type = obj.TYPE_BRANCH
		p3.Reg = v.Args[1].Reg()
		p3.From.Type = obj.TYPE_REG
		p3.From.Reg = v.Args[0].Reg()
		gc.Patch(p3, p)

	case ssa.OpRISCVLoweredMove:
		mov, sz := largestMove(v.AuxInt)

		//	mov	(Rarg1), T2
		//	mov	T2, (Rarg0)
		//	ADD	$sz, Rarg0
		//	ADD	$sz, Rarg1
		//	BNE	Rarg2, Rarg0, -4(PC)

		p := gc.Prog(mov)
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = v.Args[1].Reg()
		p.To.Type = obj.TYPE_REG
		p.To.Reg = riscv.REG_T2

		p2 := gc.Prog(mov)
		p2.From.Type = obj.TYPE_REG
		p2.From.Reg = riscv.REG_T2
		p2.To.Type = obj.TYPE_MEM
		p2.To.Reg = v.Args[0].Reg()

		p3 := gc.Prog(riscv.AADD)
		p3.From.Type = obj.TYPE_CONST
		p3.From.Offset = sz
		p3.To.Type = obj.TYPE_REG
		p3.To.Reg = v.Args[0].Reg()

		p4 := gc.Prog(riscv.AADD)
		p4.From.Type = obj.TYPE_CONST
		p4.From.Offset = sz
		p4.To.Type = obj.TYPE_REG
		p4.To.Reg = v.Args[1].Reg()

		p5 := gc.Prog(riscv.ABNE)
		p5.To.Type = obj.TYPE_BRANCH
		p5.Reg = v.Args[2].Reg()
		p5.From.Type = obj.TYPE_REG
		p5.From.Reg = v.Args[1].Reg()
		gc.Patch(p5, p)

	case ssa.OpRISCVLoweredNilCheck:
		// Issue a load which will fault if arg is nil.
		// TODO: optimizations. See arm and amd64 LoweredNilCheck.
		p := gc.Prog(riscv.AMOVB)
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = v.Args[0].Reg()
		gc.AddAux(&p.From, v)
		p.To.Type = obj.TYPE_REG
		p.To.Reg = riscv.REG_ZERO
		if gc.Debug_checknil != 0 && v.Line > 1 { // v.Line == 1 in generated wrappers
			gc.Warnl(v.Line, "generated nil check")
		}
	case ssa.OpRISCVLoweredGetClosurePtr:
		// Closure pointer is S4 (riscv.REG_CTXT).
		// TODO: replace this inline check with gc.CheckLoweredGetClosurePtr(v) once upstream dev.ssa is merged
		if entry := v.Block.Func.Entry; entry != v.Block || entry.Values[0] != v {
			gc.Fatalf("badly placed LoweredGetClosurePtr: %v %v", v.Block, v)
		}
	case ssa.OpRISCVLoweredExitProc:
		// MOV rc, A0
		p := gc.Prog(riscv.AMOV)
		p.From.Type = obj.TYPE_REG
		p.From.Reg = v.Args[0].Reg()
		p.To.Type = obj.TYPE_REG
		p.To.Reg = riscv.REG_A0
		// MOV $SYS_EXIT_GROUP, A7
		p = gc.Prog(riscv.AMOV)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = 94 // SYS_EXIT_GROUP
		p.To.Type = obj.TYPE_REG
		p.To.Reg = riscv.REG_A7
		// SCALL
		p = gc.Prog(riscv.AECALL)
	default:
		v.Fatalf("Unhandled op %v", v.Op)
	}
}

func ssaGenBlock(s *gc.SSAGenState, b, next *ssa.Block) {
	s.SetLineno(b.Line)

	switch b.Kind {
	case ssa.BlockDefer:
		// defer returns in A0:
		// 0 if we should continue executing
		// 1 if we should jump to deferreturn call
		p := gc.Prog(riscv.ABNE)
		p.To.Type = obj.TYPE_BRANCH
		p.From.Type = obj.TYPE_REG
		p.From.Reg = riscv.REG_ZERO
		p.Reg = riscv.REG_A0
		s.Branches = append(s.Branches, gc.Branch{P: p, B: b.Succs[1].Block()})
		if b.Succs[0].Block() != next {
			p := gc.Prog(obj.AJMP)
			p.To.Type = obj.TYPE_BRANCH
			s.Branches = append(s.Branches, gc.Branch{P: p, B: b.Succs[0].Block()})
		}
	case ssa.BlockPlain:
		if b.Succs[0].Block() != next {
			p := gc.Prog(obj.AJMP)
			p.To.Type = obj.TYPE_BRANCH
			s.Branches = append(s.Branches, gc.Branch{P: p, B: b.Succs[0].Block()})
		}
	case ssa.BlockExit:
		gc.Prog(obj.AUNDEF)
	case ssa.BlockRet:
		gc.Prog(obj.ARET)
	case ssa.BlockRetJmp:
		p := gc.Prog(obj.AJMP)
		p.To.Type = obj.TYPE_MEM
		p.To.Name = obj.NAME_EXTERN
		p.To.Sym = gc.Linksym(b.Aux.(*gc.Sym))
	case ssa.BlockRISCVBNE:
		// Conditional branch if Control != 0.
		p := gc.Prog(riscv.ABNE)
		p.To.Type = obj.TYPE_BRANCH
		p.Reg = b.Control.Reg()
		p.From.Type = obj.TYPE_REG
		p.From.Reg = riscv.REG_ZERO
		switch next {
		case b.Succs[0].Block():
			p.As = riscv.InvertBranch(p.As)
			s.Branches = append(s.Branches, gc.Branch{P: p, B: b.Succs[1].Block()})
		case b.Succs[1].Block():
			s.Branches = append(s.Branches, gc.Branch{P: p, B: b.Succs[0].Block()})
		default:
			s.Branches = append(s.Branches, gc.Branch{P: p, B: b.Succs[0].Block()})
			q := gc.Prog(obj.AJMP)
			q.To.Type = obj.TYPE_BRANCH
			s.Branches = append(s.Branches, gc.Branch{P: q, B: b.Succs[1].Block()})
		}

	default:
		b.Fatalf("Unhandled kind %v", b.Kind)
	}
}
