package main

func main() {
	a := new(float64)
	*a = 5.0
	b := new(float64)
	*b = 7.0
	c := new(float64)
	*c = *a + *b
	riscvexit(int(*c))
}
