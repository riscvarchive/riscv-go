package main

import "os"

func main() {
	a := new(uint8)
	*a = 5
	b := new(uint8)
	*b = 7
	c := new(uint64)
	*c = uint64(*a) + uint64(*b)
	if *c != 12 {
		os.Exit(1)
	}

	d := new(int8)
	*d = -5
	e := new(int8)
	*e = -7
	f := new(int64)
	*f = int64(*d) + int64(*e)
	if *f != -12 {
		os.Exit(2)
	}

	os.Exit(0)
}
