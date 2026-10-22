package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func newTestManager(t *testing.T) *ResourceManager {
	t.Helper()

	rm, err := NewResourceManager(
		ServerConfig{
			ID:       "server-1",
			Capacity: Resources{CPU: 10, RAM: 10},
		},
		ServerConfig{
			ID:       "server-2",
			Capacity: Resources{CPU: 10, RAM: 10},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return rm
}

func TestConcurrentReservations(t *testing.T) {
	rm := newTestManager(t)

	var wg sync.WaitGroup
	var succeeded atomic.Int32

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			_, err := rm.Reserve(Job{
				ID:       fmt.Sprintf("job-%d", i),
				Required: Resources{CPU: 1, RAM: 1},
			})

			if err == nil {
				succeeded.Add(1)
				return
			}

			if !errors.Is(err, ErrNoCapacity) {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Երկու սերվերի ընդհանուր տարողությունը 20 է։
	if got := succeeded.Load(); got != 20 {
		t.Fatalf("got %d successful jobs, want 20", got)
	}

	var total Resources
	for _, s := range rm.Snapshot() {
		if s.Used.CPU > s.Capacity.CPU ||
			s.Used.RAM > s.Capacity.RAM {
			t.Fatalf("capacity exceeded: %+v", s)
		}

		total.CPU += s.Used.CPU
		total.RAM += s.Used.RAM
	}

	if total != (Resources{CPU: 20, RAM: 20}) {
		t.Fatalf("unexpected total usage: %+v", total)
	}
}

func TestReserveIsIdempotent(t *testing.T) {
	rm := newTestManager(t)

	job := Job{
		ID:       "job-A",
		Required: Resources{CPU: 4, RAM: 4},
	}

	first, err := rm.Reserve(job)
	if err != nil {
		t.Fatal(err)
	}

	second, err := rm.Reserve(job)
	if err != nil {
		t.Fatal(err)
	}

	if first != second {
		t.Fatalf("server changed: %s != %s", first, second)
	}

	var total Resources
	for _, s := range rm.Snapshot() {
		total.CPU += s.Used.CPU
		total.RAM += s.Used.RAM
	}

	if total != job.Required {
		t.Fatalf("resources reserved more than once: %+v", total)
	}
}

func TestReleaseIsIdempotentInEffect(t *testing.T) {
	rm := newTestManager(t)

	job := Job{
		ID:       "job-A",
		Required: Resources{CPU: 4, RAM: 4},
	}

	if _, err := rm.Reserve(job); err != nil {
		t.Fatal(err)
	}

	if err := rm.Release(job.ID); err != nil {
		t.Fatal(err)
	}

	if err := rm.Release(job.ID); !errors.Is(err, ErrUnknownJob) {
		t.Fatalf("expected ErrUnknownJob, got %v", err)
	}

	for _, s := range rm.Snapshot() {
		if s.Used != (Resources{}) {
			t.Fatalf("resources leaked: %+v", s)
		}
	}
}
