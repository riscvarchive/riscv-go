package main

import "os"

func main() {
	a := new(int)
	*a = 4
	b := new(int)
	*b = 3
	c := new(int)
	*c = *a * *b
	os.Exit(*c)
}
