package main

import (
	"fmt"
	"sync"
)

func main() {
	numbers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	var wg sync.WaitGroup
	ch := make(chan int)

	for _, n := range numbers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch <- n * n
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for v := range ch {
		fmt.Println(v)
	}

}
