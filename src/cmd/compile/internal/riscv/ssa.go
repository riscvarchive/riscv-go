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
		panic("load float unsupported")
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
	}

	panic("bad load type")
}

// storeByType returns the store instruction of the given type.
func storeByType(t ssa.Type) obj.As {
	width := t.Size()
	if t.IsFloat() {
		panic("store float unsupported")
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
	}

	panic("bad store type")
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
		rs := gc.SSARegNum(v.Args[0])
		rd := gc.SSARegNum(v)
		if rs == rd {
			return
		}
		if v.Type.IsFloat() {
			v.Fatalf("OpCopy float not implemented")
		}
		p := gc.Prog(riscv.AMOV)
		p.From.Type = obj.TYPE_REG
		p.From.Reg = rs
		p.To.Type = obj.TYPE_REG
		p.To.Reg = rd
	case ssa.OpLoadReg:
		if v.Type.IsFlags() {
			v.Unimplementedf("load flags not implemented: %v", v.LongString())
			return
		}
		p := gc.Prog(loadByType(v.Type))
		n, off := gc.AutoVar(v.Args[0])
		p.From.Type = obj.TYPE_MEM
		p.From.Node = n
		p.From.Sym = gc.Linksym(n.Sym)
		p.From.Offset = off
		if n.Class == gc.PPARAM || n.Class == gc.PPARAMOUT {
			p.From.Name = obj.NAME_PARAM
			p.From.Offset += n.Xoffset
		} else {
			p.From.Name = obj.NAME_AUTO
		}
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)
	case ssa.OpStoreReg:
		if v.Type.IsFlags() {
			v.Unimplementedf("store flags not implemented: %v", v.LongString())
			return
		}
		p := gc.Prog(storeByType(v.Type))
		p.From.Type = obj.TYPE_REG
		p.From.Reg = gc.SSARegNum(v.Args[0])
		n, off := gc.AutoVar(v)
		p.To.Type = obj.TYPE_MEM
		p.To.Node = n
		p.To.Sym = gc.Linksym(n.Sym)
		p.To.Offset = off
		if n.Class == gc.PPARAM || n.Class == gc.PPARAMOUT {
			p.To.Name = obj.NAME_PARAM
			p.To.Offset += n.Xoffset
		} else {
			p.To.Name = obj.NAME_AUTO
		}
	case ssa.OpVarDef:
		gc.Gvardef(v.Aux.(*gc.Node))
	case ssa.OpVarKill:
		gc.Gvarkill(v.Aux.(*gc.Node))
	case ssa.OpVarLive:
		gc.Gvarlive(v.Aux.(*gc.Node))
	case ssa.OpKeepAlive:
		gc.KeepAlive(v)
	case ssa.OpSP, ssa.OpSB:
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
		r := gc.SSARegNum(v)
		r1 := gc.SSARegNum(v.Args[0])
		r2 := gc.SSARegNum(v.Args[1])
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
		p.From.Reg = gc.SSARegNum(v.Args[0])
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)
	case ssa.OpRISCVADDI, ssa.OpRISCVXORI, ssa.OpRISCVORI, ssa.OpRISCVANDI,
		ssa.OpRISCVSLLI, ssa.OpRISCVSRAI, ssa.OpRISCVSRLI, ssa.OpRISCVSLTI,
		ssa.OpRISCVSLTIU:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = v.AuxInt
		p.From3 = &obj.Addr{Type: obj.TYPE_REG, Reg: gc.SSARegNum(v.Args[0])}
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)
	case ssa.OpRISCVMOVBconst, ssa.OpRISCVMOVWconst, ssa.OpRISCVMOVLconst, ssa.OpRISCVMOVQconst:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = v.AuxInt
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)
	case ssa.OpRISCVMOVSconst:
		p := gc.Prog(v.Op.Asm())
		// Convert the float to the equivalent integer literal so we can
		// move it using existing infrastructure.
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(math.Float32bits(float32(math.Float64frombits(uint64(v.AuxInt)))))
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)
	case ssa.OpRISCVMOVmem:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_ADDR
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)

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
		if reg := gc.SSAReg(v.Args[0]); reg.Name() != wantreg {
			v.Fatalf("bad reg %s for symbol type %T, want %s", reg.Name(), v.Aux, wantreg)
		}
	case ssa.OpRISCVLB, ssa.OpRISCVLH, ssa.OpRISCVLW, ssa.OpRISCVLD, ssa.OpRISCVLBU, ssa.OpRISCVLHU, ssa.OpRISCVLWU,
		ssa.OpRISCVFLW, ssa.OpRISCVFLD:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = gc.SSARegNum(v.Args[0])
		gc.AddAux(&p.From, v)
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)
	case ssa.OpRISCVSB, ssa.OpRISCVSH, ssa.OpRISCVSW, ssa.OpRISCVSD,
		ssa.OpRISCVFSW, ssa.OpRISCVFSD:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = gc.SSARegNum(v.Args[1])
		p.To.Type = obj.TYPE_MEM
		p.To.Reg = gc.SSARegNum(v.Args[0])
		gc.AddAux(&p.To, v)
	case ssa.OpRISCVSEQZ, ssa.OpRISCVSNEZ:
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = gc.SSARegNum(v.Args[0])
		p.To.Type = obj.TYPE_REG
		p.To.Reg = gc.SSARegNum(v)
	case ssa.OpRISCVBEQ, ssa.OpRISCVBNE, ssa.OpRISCVBLT, ssa.OpRISCVBLTU, ssa.OpRISCVBGE, ssa.OpRISCVBGEU:
		// These are flag pseudo-ops used as control values for conditional branch blocks.
		// See the discussion in RISCVOps.
		// The actual conditional branch instruction will be issued in ssaGenBlock.
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
			p.To.Reg = gc.SSARegNum(v.Args[0])
		}
		if gc.Maxarg < v.AuxInt {
			gc.Maxarg = v.AuxInt
		}
	case ssa.OpRISCVLoweredNilCheck:
		// Issue a load which will fault if arg is nil.
		// TODO: optimizations. See arm and amd64 LoweredNilCheck.
		p := gc.Prog(riscv.AMOVB)
		p.From.Type = obj.TYPE_MEM
		p.From.Reg = gc.SSARegNum(v.Args[0])
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
		p.From.Reg = gc.SSARegNum(v.Args[0])
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
	case ssa.BlockPlain, ssa.BlockCall, ssa.BlockCheck:
		if b.Succs[0].Block() != next {
			p := gc.Prog(obj.AJMP)
			p.To.Type = obj.TYPE_BRANCH
			s.Branches = append(s.Branches, gc.Branch{P: p, B: b.Succs[0].Block()})
		}
	case ssa.BlockExit:
		gc.Prog(obj.AUNDEF)
	case ssa.BlockRet:
		gc.Prog(obj.ARET)
	case ssa.BlockRISCVBRANCH:
		// Conditional branch. The control value tells us what kind.
		v := b.Control
		// Double-check that the value is exactly where we expect it.
		if v.Block != b {
			gc.Fatalf("control value in the wrong block %v, want %v: %v", v.Block, b, v.LongString())
		}
		if v != b.Values[len(b.Values)-1] {
			gc.Fatalf("badly scheduled control value for block %v: %v", b, v.LongString())
		}
		p := gc.Prog(v.Op.Asm())
		p.To.Type = obj.TYPE_BRANCH
		p.Reg = gc.SSARegNum(v.Args[1])
		p.From.Type = obj.TYPE_REG
		p.From.Reg = gc.SSARegNum(v.Args[0])
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
