package main

import "os"

//go:noinline
func add10(i int) int {
	return i + 10
}

//go:noinline
func add20(i int) int {
	return add10(i) + 10
}

func ret1234() int {
	return 0x1234
}

func main() {
	// Check return value of leaf function.
	if r := add10(10); r != 20 {
		os.Exit(1)
	}

	// Check multiple levels of CALL.
	if r := add20(30); r != 50 {
		os.Exit(2)
	}

	// Check function pointers.
	if fn := ret1234; fn() != 0x1234 {
		os.Exit(3)
	}

	// Check closures.
	a := 0
	// Assigning the closure then calling it ensures that it uses the
	// closure pointer rather than just passing &a as an argument.
	//
	// Warning! This might get optimized away in the future, but it works
	// for now.
	fn := func() {
		a += 1
	}
	fn()
	if a != 1 {
		os.Exit(4)
	}

	os.Exit(0)
}
