package main

import "os"

func main() {
	a := new(float32)
	*a = 5.0
	b := new(float32)
	*b = 7.0
	c := new(float32)
	*c = *a + *b
	os.Exit(int(*c))
}
