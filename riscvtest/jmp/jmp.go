package main

import "os"

func ReturnZero() int

func returnZero() int {
	return 0
}

func main() {
	os.Exit(ReturnZero())
}
