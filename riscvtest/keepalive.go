package main

import "os"

// OpKeepAlive is used to "keep input pointer args live until return".
// The compiler will emit OpKeepAlive in this function.
//go:noinline
func keepalive(a *int) {
	*a = 2
}

func main() {
	var a int
	keepalive(&a)
	if a != 2 {
		os.Exit(1)
	}

	os.Exit(0)
}
