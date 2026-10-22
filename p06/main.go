package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Main started")

	go func() {
		for i := range 5 {
			fmt.Println(i)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		for _, c := range "ABCDE" {
			fmt.Printf(" %c\n", c)
			time.Sleep(150 * time.Millisecond)
		}
	}()

	time.Sleep(1 * time.Second)
	fmt.Println("Main finished")
}
