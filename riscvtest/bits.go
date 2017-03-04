package main

import "os"

func main() {
	x := new(int)
	*x = 0x10000000

	y := new(int)

	*y = 1
	*x |= 0x01100000 // now 0x1110000

	*y = 1
	*x &= 0x11011111 // now 0x1100000

	*y = 1
	*x ^= 0x11000000 // now 0

	*y = 1
	if *x != 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
