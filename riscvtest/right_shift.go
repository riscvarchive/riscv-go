package main

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
		riscvexit(2)
	}

	*y = 1
	*b32 >>= *s64 // now -8
	*y = 1
	*b32 >>= *s32 // now -4
	*y = 1
	*b32 >>= *s16 // now -2
	*y = 1
	*b32 >>= *s8 // now -1
	// FIXME(prattmic): < 64-bit comparisons not supported
	if int64(*b32) != -1 {
		riscvexit(3)
	}

	*y = 1
	*b16 >>= *s64 // now -8
	*y = 1
	*b16 >>= *s32 // now -4
	*y = 1
	*b16 >>= *s16 // now -2
	*y = 1
	*b16 >>= *s8 // now -1
	if int64(*b16) != -1 {
		riscvexit(4)
	}

	*y = 1
	*b8 >>= *s64 // now -8
	*y = 1
	*b8 >>= *s32 // now -4
	*y = 1
	*b8 >>= *s16 // now -2
	*y = 1
	*b8 >>= *s8 // now -1
	if int64(*b8) != -1 {
		riscvexit(5)
	}

	riscvexit(0)
}
