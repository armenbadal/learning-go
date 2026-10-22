package main

import (
	"fmt"
	"sync"
)

type job struct {
	id          int
	description string
}

func doJob(wid int, j *job) {
	fmt.Printf("Worker ID %d, Job ID: %d, Description: %s\n", wid, j.id, j.description)
}

type scheduler struct {
	capacity int
	jobs     chan *job
	wg       sync.WaitGroup
}

func (s *scheduler) start() {
	for i := range s.capacity {
		s.wg.Add(1)
		go func(wid int) {
			defer s.wg.Done()
			for j := range s.jobs {
				doJob(wid, j)
			}
		}(i)
	}
}

func (s *scheduler) add(j *job) {
	s.jobs <- j
}

func (s *scheduler) stop() {
	close(s.jobs)
	s.wg.Wait()
}

func main() {
	sc := &scheduler{
		capacity: 5,
		jobs:     make(chan *job, 4),
		wg:       sync.WaitGroup{},
	}
	sc.start()

	for i := range 16 {
		sc.add(&job{i, fmt.Sprintf("Job %d", i)})
	}

	sc.stop()
}
