package main

func main() {
	x := new(int)
	*x = -100

	u := new(uint)
	*u = 100

	// Use x to do fake memory writes to prevent optimizations
	// from knowing the value of *x or *u.
	y := new(int)

	var ok int
	*y = 1
	if *x == -100 {
		ok++
	}
	*y = 1
	if *x != -99 {
		ok++
	}
	*y = 1
	if *x <= -100 {
		ok++
	}
	*y = 1
	if *x <= -99 {
		ok++
	}
	*y = 1
	if *x < -99 {
		ok++
	}
	*y = 1
	if *x >= -100 {
		ok++
	}
	*y = 1
	if *x >= -101 {
		ok++
	}
	*y = 1
	if *x > -101 {
		ok++
	}
	*y = 1
	if ok != 8 {
		riscvexit(1)
	}

	ok = 0
	*y = 1
	if *u == 100 {
		ok++
	}
	*y = 1
	if *u != 99 {
		ok++
	}
	*y = 1
	if *u <= 100 {
		ok++
	}
	*y = 1
	if *u <= 101 {
		ok++
	}
	*y = 1
	if *u < 101 {
		ok++
	}
	*y = 1
	if *u >= 100 {
		ok++
	}
	*y = 1
	if *u >= 99 {
		ok++
	}
	*y = 1
	if *u > 99 {
		ok++
	}
	if ok != 8 {
		riscvexit(2)
	}

	*y = 1
	switch {
	case *x == 99:
		riscvexit(3)
	case *u == 99:
		riscvexit(4)
	}

	riscvexit(0)
}
