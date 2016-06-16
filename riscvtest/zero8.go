package main

func main() {
	var a [8]byte
	a[1] = 3
	riscvexit(a[0] + a[1] + a[6] + a[7])
}
