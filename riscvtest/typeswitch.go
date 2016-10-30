package main

type Color interface{}

type Blue int
type Red int

func main() {
	var c Color
	c = Blue(0)

	y := new(int)
	*y = 0

	switch c := c.(type) {
	case Blue:
		riscvexit(int(c))
	case Red:
		riscvexit(1)
	default:
		riscvexit(2)
	}
}
