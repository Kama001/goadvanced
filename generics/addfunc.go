package main

import "fmt"

type generic interface {
	float64 | int32 | int
}

func add[T generic](a T, b T) {
	fmt.Println(a + b)
}

func main() {
	add(5, 4)
	add(2423423424234234234234234, 6.4)
}
