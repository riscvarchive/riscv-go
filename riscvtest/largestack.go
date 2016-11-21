package main

func main() {
	a := new([2000]int)
	(*a)[0] = 1
	if (*a)[0] != 1 {
		riscvexit(1)
	}

	(*a)[1999] = 2
	b := (*a)[1998]
	if b != 0 {
		riscvexit(2)
	}

	riscvexit(0)
}
