package main

import "os"

var a int

//go:noinline
func checkA() {
	if a != 42 {
		os.Exit(2)
	}
}

func main() {
	if a != 0 {
		os.Exit(1)
	}

	a = 42

	checkA()

	os.Exit(0)
}
