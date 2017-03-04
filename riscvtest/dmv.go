package main

import "os"

func main() {
	a := new(float64)
	*a = 3

	b := new(float32)
	*b = float32(*a) + 2

	c := new(float64)
	*c = float64(*b)

	os.Exit(int(*c))
}
