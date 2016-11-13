// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

// TODO(prattmic): everythin
package atomic

import "unsafe"

//go:nosplit
//go:noinline
func Load(ptr *uint32) uint32 {
	return *ptr
}

//go:nosplit
//go:noinline
func Loadp(ptr unsafe.Pointer) unsafe.Pointer {
	return *(*unsafe.Pointer)(ptr)
}

//go:nosplit
func Xadd64(ptr *uint64, delta int64) uint64 {
	return 0
}

//go:nosplit
func Xadduintptr(ptr *uintptr, delta uintptr) uintptr {
	return 0
}

//go:nosplit
func Xchg64(ptr *uint64, new uint64) uint64 {
	return 0
}

//go:nosplit
func Xadd(ptr *uint32, delta int32) uint32 {
	return 0
}

//go:nosplit
func Xchg(ptr *uint32, new uint32) uint32 {
	return 0
}

//go:nosplit
func Xchguintptr(ptr *uintptr, new uintptr) uintptr {
	return 0
}

//go:nosplit
func Load64(ptr *uint64) uint64 {
	return 0
}

//go:nosplit
func And8(ptr *uint8, val uint8) {
}

//go:nosplit
func Or8(ptr *uint8, val uint8) {
}

//go:nosplit
func Cas64(ptr *uint64, old, new uint64) bool {
	return true
}

//go:nosplit
func Store(ptr *uint32, val uint32) {
}

//go:nosplit
func Store64(ptr *uint64, val uint64) {
}

// NO go:nosplit annotation; see atomic_pointer.go.
func StorepNoWB(ptr unsafe.Pointer, val unsafe.Pointer) {
}
