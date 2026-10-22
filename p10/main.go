package main

import (
	"fmt"
	"sync"
)

func main() {
	jobs := make(chan int)
	results := make(chan int)
	var wg sync.WaitGroup

	// job provider
	go func() {
		for i := range 20 {
			jobs <- (i + 1)
		}
		close(jobs)
	}()

	// workers
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range jobs {
				results <- (e * e)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		fmt.Println(r)
	}
}
