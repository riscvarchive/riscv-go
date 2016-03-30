// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"log"

	"cmd/compile/internal/gc"
	"cmd/compile/internal/ssa"
	"cmd/internal/obj"
)

// ssaRegToReg maps ssa register numbers to obj register numbers.
var ssaRegToReg = []int16{}

// markMoves marks any MOVXconst ops that need to avoid clobbering flags.
func ssaMarkMoves(s *gc.SSAGenState, b *ssa.Block) {
	log.Printf("ssaMarkMoves")
}

func ssaGenValue(s *gc.SSAGenState, v *ssa.Value) {
	log.Printf("ssaGenValue")
	s.SetLineno(v.Line)

	switch v.Op {
	case ssa.OpRISCVADD:
		r := gc.SSARegNum(v)
		r1 := gc.SSARegNum(v.Args[0])
		r2 := gc.SSARegNum(v.Args[1])
		p := gc.Prog(v.Op.Asm())
		p.From.Type = obj.TYPE_REG
		p.From.Reg = r1
		p.From3 = &obj.Addr{}
		p.From3.Type = obj.TYPE_REG
		p.From3.Reg = r2
		p.To.Type = obj.TYPE_REG
		p.To.Reg = r
	}
}

func ssaGenBlock(s *gc.SSAGenState, b, next *ssa.Block) {
	log.Printf("ssaGenBlock")
}
