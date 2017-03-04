package main

import "os"

func main() {
	var a [8]byte
	a[1] = 3
	os.Exit(int(a[0] + a[1] + a[6] + a[7]))
}
