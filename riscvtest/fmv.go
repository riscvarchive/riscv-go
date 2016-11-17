package main

func main() {
	a := float32(5)
	if int(a) != 5 {
		riscvexit(1)
	}

	b := float32(1e6)
	if b < -1e7 || b > 1e7 {
		riscvexit(2)
	}

	riscvexit(0)
}
