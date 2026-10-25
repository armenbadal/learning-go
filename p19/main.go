package main

import (
	"ecol/collections"
	"fmt"
)

type Server struct {
	Name string
	CPU  int
	RAM  int
}

func main() {
	servers := []Server{
		{
			Name: "A",
			CPU:  1,
			RAM:  8,
		},
		{
			Name: "B",
			CPU:  2,
			RAM:  8,
		},
		{
			Name: "C",
			CPU:  4,
			RAM:  32,
		},
		{
			Name: "D",
			CPU:  32,
			RAM:  128,
		},
		{
			Name: "E",
			CPU:  1,
			RAM:  16,
		},
		{
			Name: "F",
			CPU:  2,
			RAM:  64,
		},
	}

	oneCpu := collections.Filter(servers, func(s Server) bool {
		return s.CPU == 1
	})
	collections.ForEach(oneCpu, func(s Server) {
		fmt.Println(s)
	})
}
