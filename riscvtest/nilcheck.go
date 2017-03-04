package main

import "os"

func main() {
	var x *int
	*x = 0
	os.Exit(0)
}
