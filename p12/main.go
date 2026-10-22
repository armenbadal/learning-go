package main

import (
	"context"
	"fmt"
	"sync"
)

func process(n int) (int, error) {
	if n%7 != 0 {
		return n * n, nil
	}

	return 0, fmt.Errorf("%d-ը բաժանվում է 7-ի", n)
}

func provider(ctx context.Context, jobs chan<- int) {
	defer close(jobs)

	for i := range 20 {
		select {
		case <-ctx.Done():
			return
		case jobs <- (i + 1):
		}
	}
}

func worker(ctx context.Context, jobs <-chan int, results chan<- int, wg *sync.WaitGroup, ech chan<- error) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case n, ok := <-jobs:
			if !ok {
				return
			}
			sq, e := process(n)
			if e != nil {
				ech <- e
				return
			} else {
				select {
				case <-ctx.Done():
					return
				case results <- sq:
				}
			}
		}
	}
}

func printResults(results <-chan int) {
	for r := range results {
		fmt.Println(r)
	}
}

func main() {
	jobs := make(chan int)
	results := make(chan int)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ech := make(chan error, 1)

	// job provider
	go provider(ctx, jobs)

	// workers
	for range 3 {
		wg.Add(1)
		go worker(ctx, jobs, results, &wg, ech)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		err := <-ech
		if err != nil {
			fmt.Println(err)
			cancel()
		}
	}()

	printResults(results)
}
