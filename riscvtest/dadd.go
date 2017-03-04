package main

import "os"

func main() {
	a := new(float64)
	*a = 5.0
	b := new(float64)
	*b = 7.0
	c := new(float64)
	*c = *a + *b
	os.Exit(int(*c))
}
