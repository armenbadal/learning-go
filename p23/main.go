package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int

	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 1000 {
				counter++
			}
		}()
	}

	wg.Wait()

	fmt.Println("Counter:", counter)
}

// Fix data race
