package main

func main() {
	a := new(int)
	*a = 36
	b := new(int)
	*b = 3
	c := new(int)
	*c = *a / *b
	riscvexit(*c)
}
