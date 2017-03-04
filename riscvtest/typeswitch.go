package main

import "os"

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
		os.Exit(int(c))
	case Red:
		os.Exit(1)
	default:
		os.Exit(2)
	}
}
