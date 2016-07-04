package main

func main() {
	x := new(uint64)
	*x = 0x0102030405060708
	y := new(uint64) // used to force loads of x from memory
	*y = 0
	*x = ^*x
	*y = 0
	if *x != 0xfefdfcfbfaf9f8f7 {
		riscvexit(1)
	}
	riscvexit(0)
}
