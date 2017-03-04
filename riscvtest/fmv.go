package main

import "os"

func main() {
	a := float32(5)
	if int(a) != 5 {
		os.Exit(1)
	}

	b := float32(1e6)
	if b < -1e7 || b > 1e7 {
		os.Exit(2)
	}

	os.Exit(0)
}
