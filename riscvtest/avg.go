package main

func main() {
	x := new(uint64)
	*x = 739
	y := new(int)
	*y = 0
	riscvexit(*x/39 - 18) // magic divide, uses avg64u; should yield zero
}
