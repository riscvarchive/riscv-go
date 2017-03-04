package main

import "os"

func main() {
	b64 := new(int64)
	b32 := new(int32)
	b16 := new(int16)
	b8 := new(int8)
	*b64 = -16
	*b32 = -16
	*b16 = -16
	*b8 = -16

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
	*b64 >>= *s64 // now -8
	*y = 1
	*b64 >>= *s32 // now -4
	*y = 1
	*b64 >>= *s16 // now -2
	*y = 1
	*b64 >>= *s8 // now -1
	if *b64 != -1 {
		os.Exit(2)
	}

	*y = 1
	*b32 >>= *s64 // now -8
	*y = 1
	*b32 >>= *s32 // now -4
	*y = 1
	*b32 >>= *s16 // now -2
	*y = 1
	*b32 >>= *s8 // now -1
	if *b32 != -1 {
		os.Exit(3)
	}

	*y = 1
	*b16 >>= *s64 // now -8
	*y = 1
	*b16 >>= *s32 // now -4
	*y = 1
	*b16 >>= *s16 // now -2
	*y = 1
	*b16 >>= *s8 // now -1
	if *b16 != -1 {
		os.Exit(4)
	}

	*y = 1
	*b8 >>= *s64 // now -8
	*y = 1
	*b8 >>= *s32 // now -4
	*y = 1
	*b8 >>= *s16 // now -2
	*y = 1
	*b8 >>= *s8 // now -1
	if *b8 != -1 {
		os.Exit(5)
	}

	// Large shift sanity test
	*b8 = -1 << 7
	*s8 = 8
	*y = 1
	*b8 >>= *s8
	if *b8 != -1 {
		os.Exit(6)
	}

	*b8 = 1 << 6
	*s8 = 8
	*y = 1
	*b8 >>= *s8
	if *b8 != 0 {
		os.Exit(7)
	}

	*b16 = -1 << 15
	*s16 = 16
	*y = 1
	*b16 >>= *s16
	if *b16 != -1 {
		os.Exit(8)
	}

	*b16 = 1 << 14
	*s16 = 16
	*y = 1
	*b16 >>= *s16
	if *b16 != 0 {
		os.Exit(9)
	}

	*b32 = -1 << 31
	*s32 = 32
	*y = 1
	*b32 >>= *s32
	if *b32 != -1 {
		os.Exit(10)
	}

	*b32 = 1 << 30
	*s32 = 32
	*y = 1
	*b32 >>= *s32
	if *b32 != 0 {
		os.Exit(11)
	}

	*b64 = -1 << 63
	*s64 = 64
	*y = 1
	*b64 >>= *s64
	if *b64 != -1 {
		os.Exit(12)
	}

	*b64 = 1 << 62
	*s64 = 64
	*y = 1
	*b64 >>= *s64
	if *b64 != 0 {
		os.Exit(13)
	}

	os.Exit(0)
}
