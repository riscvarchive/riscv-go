package main

//go:noinline
func exit(rc int) {
	riscvexit(rc)
}

func main() {
	exit(99)
}
