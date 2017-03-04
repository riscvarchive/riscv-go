package main

import "os"

func main() {
	a := new(int)
	*a = 5
	b := new(int)
	*b = 7
	c := new(int)
	*c = *a + *b
	os.Exit(*c)
}
