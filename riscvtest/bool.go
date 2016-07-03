package main

func main() {
	x := new(int)
	*x = 0

	y := new(int)
	*y = 0

	z := new(bool)
	*z = *x == *y

	*x = 0 // distract compiler

	if !*z {
		riscvexit(1)
	}

	*z = x == nil
	*x = 0

	if *z {
		riscvexit(2)
	}

	riscvexit(0)
}
