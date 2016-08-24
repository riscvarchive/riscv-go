package main

func main() {
	a := int(5)
	b := int(7)
	c := float32(a)
	d := float32(b)
	riscvexit(int(c + d))
}
