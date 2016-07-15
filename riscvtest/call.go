package main

// f is a large function to prevent inlining.
//func f() int {
//	x := new(int)
//	y := new(int)
//
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//	*x++
//	*y++
//
//	return *x
//}

func f() int {
	x := new(int)
	y := new(int)

	*x += 1
	*y = 1

	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	*x += 1
	*y = 1
	return *x;
}

func main() {
	x := new(int)
	y := new(int)

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	*x += f()
	*y = 1

	if f() != 15 {
		riscvexit(1)
	}

	riscvexit(0)
}
