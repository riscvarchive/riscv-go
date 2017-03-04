package main

import "os"

func main() {
	b64 := new(int64)
	b32 := new(int32)
	b16 := new(int16)
	b8 := new(int8)
	*b64 = 1
	*b32 = 1
	*b16 = 1
	*b8 = 1

	s64 := new(uint64)
	s32 := new(uint32)
	s16 := new(uint16)
	s8 := new(uint8)
	*s64 = 1
	*s32 = 1
	*s16 = 1
	*s8 = 1

	y := new(int)

	*y = 1
	*b64 <<= *s64 // now 0x2
	*y = 1
	*b64 <<= *s32 // now 0x4
	*y = 1
	*b64 <<= *s16 // now 0x8
	*y = 1
	*b64 <<= *s8 // now 0x10
	if *b64 != 0x10 {
		os.Exit(2)
	}

	*y = 1
	*b32 <<= *s64 // now 0x2
	*y = 1
	*b32 <<= *s32 // now 0x4
	*y = 1
	*b32 <<= *s16 // now 0x8
	*y = 1
	*b32 <<= *s8 // now 0x10
	if *b32 != 0x10 {
		os.Exit(3)
	}

	*y = 1
	*b16 <<= *s64 // now 0x2
	*y = 1
	*b16 <<= *s32 // now 0x4
	*y = 1
	*b16 <<= *s16 // now 0x8
	*y = 1
	*b16 <<= *s8 // now 0x10
	if *b16 != 0x10 {
		os.Exit(4)
	}

	*y = 1
	*b8 <<= *s64 // now 0x2
	*y = 1
	*b8 <<= *s32 // now 0x4
	*y = 1
	*b8 <<= *s16 // now 0x8
	*y = 1
	*b8 <<= *s8 // now 0x10
	if *b8 != 0x10 {
		os.Exit(5)
	}

	// Large shift sanity test
	*b8 = 1
	*s8 = 8
	*y = 1
	*b8 <<= *s8
	if *b8 != 0 {
		os.Exit(6)
	}

	*b16 = 1
	*s16 = 16
	*y = 1
	*b16 <<= *s16
	if *b16 != 0 {
		os.Exit(7)
	}

	*b32 = 1
	*s32 = 32
	*y = 1
	*b32 <<= *s32
	if *b32 != 0 {
		os.Exit(8)
	}

	*b64 = 1
	*s64 = 64
	*y = 1
	*b64 <<= *s64
	if *b64 != 0 {
		os.Exit(9)
	}

	os.Exit(0)
}
