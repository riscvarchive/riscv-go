package main

import "os"

func main() {
	x := new(uint64)
	*x = 739
	y := new(int)
	*y = 0
	os.Exit(int(*x/39 - 18)) // magic divide, uses avg64u; should yield zero
}
