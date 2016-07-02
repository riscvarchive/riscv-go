package main

func main() {
	y := new(int)

	b64 := new(uint64)
	b32 := new(uint32)
	b16 := new(uint16)
	b8 := new(uint8)

	*b64 = 1 << 63
	*b32 = 1 << 31
	*b16 = 1 << 15
	*b8 = 1 << 7

	s64 := new(uint64)
	s32 := new(uint32)
	s16 := new(uint16)
	s8 := new(uint8)
	*s64 = 1
	*s32 = 1
	*s16 = 1
	*s8 = 1

	*y = 1
	*b64 >>= *s64 // now 1 << 62
	*y = 1
	*b64 >>= *s32 // now 1 << 61
	*y = 1
	*b64 >>= *s16 // now 1 << 60
	*y = 1
	*b64 >>= *s8 // now 1 << 59
	if *b64 != 1<<59 {
		riscvexit(2)
	}

	*y = 1
	*b32 >>= *s64 // now 1 << 30
	*y = 1
	*b32 >>= *s32 // now 1 << 29
	*y = 1
	*b32 >>= *s16 // now 1 << 28
	*y = 1
	*b32 >>= *s8 // now 1 << 27
	// FIXME(prattmic): < 64-bit comparisons not supported
	if uint64(*b32) != 1<<27 {
		riscvexit(3)
	}

	*y = 1
	*b16 >>= *s64 // now 1 << 14
	*y = 1
	*b16 >>= *s32 // now 1 << 13
	*y = 1
	*b16 >>= *s16 // now 1 << 12
	*y = 1
	*b16 >>= *s8 // now 1 << 11
	if uint64(*b16) != 1<<11 {
		riscvexit(4)
	}

	*y = 1
	*b8 >>= *s64 // now 1 << 6
	*y = 1
	*b8 >>= *s32 // now 1 << 5
	*y = 1
	*b8 >>= *s16 // now 1 << 4
	*y = 1
	*b8 >>= *s8 // now 1 << 3
	if uint64(*b8) != 1<<3 {
		riscvexit(5)
	}

	riscvexit(0)
}
