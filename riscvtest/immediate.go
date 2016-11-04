package main

//go:noinline
func maxInt32() uint32 {
	return 1<<31-1
}

func main() {
	x := maxInt32()
	if x != 1<<31-1 {
		riscvexit(1)
	}

	// Upper bits don't interfere in up-conversion.
	y := uint64(maxInt32()) + uint64(maxInt32())
	if y != 1<<32-2 {
		riscvexit(2)
	}

	riscvexit(0)
}
