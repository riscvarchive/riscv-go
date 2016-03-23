// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv

import (
	"log"

	"cmd/compile/internal/gc"
	"cmd/internal/obj"
)

func gins(as obj.As, f *gc.Node, t *gc.Node) *obj.Prog {
	log.Printf("gins")

	return gc.Ctxt.NewProg() // dummy return
}
