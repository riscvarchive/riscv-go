package main

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
		riscvexit(1)
	}

	// Check multiple levels of CALL.
	if r := add20(30); r != 50 {
		riscvexit(2)
	}

	// Check function pointers
	if fn := ret1234; fn() != 0x1234 {
		riscvexit(3)
	}

	riscvexit(0)
}
