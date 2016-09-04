package main

var a int

//go:noinline
func checkA() {
	if a != 42 {
		riscvexit(2)
	}
}

func main() {
	if a != 0 {
		riscvexit(1)
	}

	a = 42

	checkA()

	riscvexit(0)
}
