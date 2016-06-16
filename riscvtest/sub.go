package main

func main() {
	a := new(int)
	*a = 19
	b := new(int)
	*b = 7
	c := new(int)
	*c = *a - *b
	riscvexit(*c)
}
