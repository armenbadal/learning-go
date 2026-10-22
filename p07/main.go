package main

import (
	"fmt"
	"sync"
)

func main() {
	numbers := []int{2, 4, 6, 8, 10}

	var wg sync.WaitGroup

	for _, n := range numbers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("%d * %d = %d\n", n, n, n*n)
		}()
	}

	wg.Wait()
}
