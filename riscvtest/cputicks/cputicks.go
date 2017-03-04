package main

import "os"

func cputicks() int64

func main() {
	x := new(int)
	*x = 1
	*x = 2
	*x = 3
	*x = 4
	os.Exit(int(cputicks()) + *x)
}
