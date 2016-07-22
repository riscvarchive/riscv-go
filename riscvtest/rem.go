package main

func main() {
	a := new(int)
	*a = 32
	b := new(int)
	*b = 20
	c := new(int)
	*c = *a % *b
	riscvexit(*c)
}
