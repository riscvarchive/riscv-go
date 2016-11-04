package main

func ReturnZero() int

func returnZero() int {
	return 0
}

func main() {
	riscvexit(ReturnZero())
}
