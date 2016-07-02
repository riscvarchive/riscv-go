package main

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
		riscvexit(2)
	}

	*y = 1
	*b32 <<= *s64 // now 0x2
	*y = 1
	*b32 <<= *s32 // now 0x4
	*y = 1
	*b32 <<= *s16 // now 0x8
	*y = 1
	*b32 <<= *s8 // now 0x10
	// FIXME(prattmic): < 64-bit comparisons not supported
	if int64(*b32) != 0x10 {
		riscvexit(3)
	}

	*y = 1
	*b16 <<= *s64 // now 0x2
	*y = 1
	*b16 <<= *s32 // now 0x4
	*y = 1
	*b16 <<= *s16 // now 0x8
	*y = 1
	*b16 <<= *s8 // now 0x10
	if int64(*b16) != 0x10 {
		riscvexit(4)
	}

	*y = 1
	*b8 <<= *s64 // now 0x2
	*y = 1
	*b8 <<= *s32 // now 0x4
	*y = 1
	*b8 <<= *s16 // now 0x8
	*y = 1
	*b8 <<= *s8 // now 0x10
	if int64(*b8) != 0x10 {
		riscvexit(5)
	}

	riscvexit(0)
}
