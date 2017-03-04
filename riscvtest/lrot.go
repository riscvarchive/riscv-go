package main

import "os"

func lrot8(x, c uint8) uint8 {
	return (x << c) | (x >> (8 - c))
}

func main() {
	x := new(uint8)
	*x = 0x81
	y := new(uint8)
	*y = 0
	os.Exit(int(lrot8(*x, 3) - 12))
}
