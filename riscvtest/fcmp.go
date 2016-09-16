package main

func main() {
	x := new(float32)
	*x = 100.0

	// Use z to do fake memory writes to prevent optimizations from knowing
	// the value of *x
	z := new(float32)

	var ok int
	*z = 1
	if *x == 100 {
		ok++
	}
	*z = 1
	if *x != 101 {
		ok++
	}
	*z = 1
	if *x <= 100 {
		ok++
	}
	*z = 1
	if *x <= 101 {
		ok++
	}
	*z = 1
	if *x < 101 {
		ok++
	}
	*z = 1
	if *x >= 100 {
		ok++
	}
	*z = 1
	if *x >= 99 {
		ok++
	}
	*z = 1
	if *x > 99 {
		ok++
	}
	*z = 1
	if ok != 8 {
		riscvexit(1)
	}

	*z = 1
	if *x != 100 {
		riscvexit(2)
	}

	riscvexit(0)
}
