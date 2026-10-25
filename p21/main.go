package main

import (
	"fmt"
	"sync"
)

func odds(co, ci chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i <= 10; i += 2 {
		<-ci
		fmt.Printf(" %d\n", i)
		co <- true
	}
}

func evens(co, ci chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 2; i <= 10; i += 2 {
		<-co
		fmt.Printf("%d\n", i)
		if i != 10 {
			ci <- true
		}
	}
}

func main() {
	co := make(chan bool)
	ci := make(chan bool)
	var wg sync.WaitGroup

	wg.Add(1)
	go odds(co, ci, &wg)
	wg.Add(1)
	go evens(co, ci, &wg)

	ci <- true

	wg.Wait()
}
