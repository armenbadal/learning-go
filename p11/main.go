package main

import (
	"fmt"
	"sync"
)

type Result struct {
	Input int
	Value int
	Err   error
}

func process(n int) (int, error) {
	if n%7 != 0 {
		return n * n, nil
	}

	return 0, fmt.Errorf("%d-ը բաժանվում է 7-ի", n)
}

func main() {
	jobs := make(chan int)
	results := make(chan Result)
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
			for n := range jobs {
				sq, e := process(n)
				results <- Result{
					Input: n,
					Value: sq,
					Err:   e,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.Err == nil {
			fmt.Printf("%d² = %d\n", r.Input, r.Value)
		} else {
			fmt.Println(r.Err)
		}
	}
}
