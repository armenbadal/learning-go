package main

import (
	"edset/set"
	"fmt"
)

func main() {
	a := set.NewSet[int]()
	a.Add(10)
	a.Add(20)
	a.Add(10)

	fmt.Println(a.Len())
	fmt.Println(a.Contains(10))

}
