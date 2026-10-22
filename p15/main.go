package main

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInvalidResources = errors.New("invalid resource request")
	ErrNoCapacity       = errors.New("insufficient resources")
	ErrDuplicateServer  = errors.New("duplicate server ID")
	ErrDuplicateJob     = errors.New("job already exists")
	ErrJobConflict      = errors.New("job ID used with different resources")
	ErrJobInProgress    = errors.New("job operation is in progress")
	ErrUnknownJob       = errors.New("unknown job")
)

type Resources struct {
	CPU int
	RAM int // GiB
}

type Job struct {
	ID       string
	Required Resources
}

type ServerConfig struct {
	ID       string
	Capacity Resources
}

type ServerSnapshot struct {
	ID       string
	Capacity Resources
	Used     Resources
}

type Allocation struct {
	ServerID string
	Required Resources
}

// server-ի վիճակը պաշտպանում է իր mutex-ը։
type server struct {
	mu       sync.Mutex
	id       string
	capacity Resources
	used     Resources
}

type ResourceManager struct {
	// Պաշտպանում է jobs և pending քարտեզները։
	jobsMu  sync.Mutex
	jobs    map[string]Allocation
	pending map[string]struct{}

	// Կոնֆիգուրացիան ստեղծումից հետո անփոփոխ է։
	servers []*server
	byID    map[string]*server
}

func NewResourceManager(configs ...ServerConfig) (*ResourceManager, error) {
	rm := &ResourceManager{
		jobs:    make(map[string]Allocation),
		pending: make(map[string]struct{}),
		byID:    make(map[string]*server),
	}

	for _, cfg := range configs {
		if cfg.ID == "" ||
			cfg.Capacity.CPU < 0 ||
			cfg.Capacity.RAM < 0 {
			return nil, fmt.Errorf(
				"invalid server configuration: %q", cfg.ID,
			)
		}

		if _, exists := rm.byID[cfg.ID]; exists {
			return nil, fmt.Errorf(
				"%w: %s", ErrDuplicateServer, cfg.ID,
			)
		}

		s := &server{
			id:       cfg.ID,
			capacity: cfg.Capacity,
		}

		rm.servers = append(rm.servers, s)
		rm.byID[cfg.ID] = s
	}

	return rm, nil
}

func validJob(job Job) bool {
	return job.ID != "" &&
		job.Required.CPU >= 0 &&
		job.Required.RAM >= 0 &&
		(job.Required.CPU > 0 || job.Required.RAM > 0)
}

// Reserve-ը փորձում է աշխատանքը տեղադրել մեկ սերվերի վրա։
func (rm *ResourceManager) Reserve(job Job) (string, error) {
	if !validJob(job) {
		return "", ErrInvalidResources
	}

	// Քայլ 1. Ամրագրում ենք job ID-ն։
	rm.jobsMu.Lock()

	if allocation, exists := rm.jobs[job.ID]; exists {
		rm.jobsMu.Unlock()

		// Նույն հարցման կրկնությունը նոր ռեսուրս չի վերցնում։
		if allocation.Required == job.Required {
			return allocation.ServerID, nil
		}

		return "", ErrJobConflict
	}

	if _, exists := rm.pending[job.ID]; exists {
		rm.jobsMu.Unlock()
		return "", ErrJobInProgress
	}

	rm.pending[job.ID] = struct{}{}
	rm.jobsMu.Unlock()

	// Քայլ 2. Փնտրում ենք հարմար սերվեր։
	for _, s := range rm.servers {
		s.mu.Lock()

		fits := s.capacity.CPU-s.used.CPU >= job.Required.CPU &&
			s.capacity.RAM-s.used.RAM >= job.Required.RAM

		if !fits {
			s.mu.Unlock()
			continue
		}

		// Քայլ 3. Ատոմային վերապահում տվյալ սերվերի վրա։
		s.used.CPU += job.Required.CPU
		s.used.RAM += job.Required.RAM
		s.mu.Unlock()

		// Քայլ 4. Գրանցում ենք allocation-ը։
		rm.jobsMu.Lock()
		delete(rm.pending, job.ID)
		rm.jobs[job.ID] = Allocation{
			ServerID: s.id,
			Required: job.Required,
		}
		rm.jobsMu.Unlock()

		return s.id, nil
	}

	// Ոչ մի սերվեր հարմար չէր։
	rm.jobsMu.Lock()
	delete(rm.pending, job.ID)
	rm.jobsMu.Unlock()

	return "", ErrNoCapacity
}

// Release-ը ազատում է allocation-ի ռեսուրսները։
func (rm *ResourceManager) Release(jobID string) error {
	rm.jobsMu.Lock()
	defer rm.jobsMu.Unlock()

	if _, exists := rm.pending[jobID]; exists {
		return ErrJobInProgress
	}

	allocation, exists := rm.jobs[jobID]
	if !exists {
		return ErrUnknownJob
	}

	s := rm.byID[allocation.ServerID]

	// Lock-երի հերթականությունը՝ jobsMu -> server.mu։
	s.mu.Lock()
	defer s.mu.Unlock()

	s.used.CPU -= allocation.Required.CPU
	s.used.RAM -= allocation.Required.RAM

	delete(rm.jobs, jobID)
	return nil
}

// Snapshot-ը վերադարձնում է սերվերների վիճակի պատճենները։
func (rm *ResourceManager) Snapshot() []ServerSnapshot {
	result := make([]ServerSnapshot, 0, len(rm.servers))

	for _, s := range rm.servers {
		s.mu.Lock()

		snapshot := ServerSnapshot{
			ID:       s.id,
			Capacity: s.capacity,
			Used:     s.used,
		}

		s.mu.Unlock()
		result = append(result, snapshot)
	}

	return result
}
