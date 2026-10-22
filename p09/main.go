package main

import "fmt"

func main() {
	jobs := make(chan int)
	done := make(chan struct{})

	go func() {
		for i := range 10 {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		for num := range jobs {
			fmt.Println(num * num)
		}
		close(done)
	}()

	<-done
}
