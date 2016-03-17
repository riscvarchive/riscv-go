// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"log"

	"cmd/compile/internal/gc"
	"cmd/compile/internal/ssa"
)

// ssaRegToReg maps ssa register numbers to obj register numbers.
var ssaRegToReg = []int16{}

// markMoves marks any MOVXconst ops that need to avoid clobbering flags.
func ssaMarkMoves(s *gc.SSAGenState, b *ssa.Block) {
	log.Printf("ssaMarkMoves")
}

func ssaGenValue(s *gc.SSAGenState, v *ssa.Value) {
	log.Printf("ssaGenValues")
}

func ssaGenBlock(s *gc.SSAGenState, b, next *ssa.Block) {
	log.Printf("ssaGenBlock")
}
