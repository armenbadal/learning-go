package main

import "fmt"

func main() {
	a := []int{1, 2, 3, 4}
	b := a[0:2:2]

	b = append(b, 10)

	fmt.Println(a)
	fmt.Println(b)
}
