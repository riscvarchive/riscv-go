package main

import "os"

func main() {
	x := new(int)
	*x = 0

	y := new(int)
	*y = 0

	z := new(bool)
	*z = *x == *y

	*x = 0 // distract compiler

	if !*z {
		os.Exit(1)
	}

	*z = x == nil
	*x = 0

	if *z {
		os.Exit(2)
	}

	os.Exit(0)
}
