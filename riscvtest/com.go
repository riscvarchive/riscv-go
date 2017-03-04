package main

import "os"

func main() {
	x := new(uint64)
	*x = 0x0102030405060708
	y := new(uint64) // used to force loads of x from memory
	*y = 0
	*x = ^*x
	*y = 0
	if *x != 0xfefdfcfbfaf9f8f7 {
		os.Exit(1)
	}
	os.Exit(0)
}
