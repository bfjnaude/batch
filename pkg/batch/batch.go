package batch

// package batch provides micro batching by running submitted Jobs with
// the configured BatchSize and Interval using a Dispatcher and BatchProcessor

import (
	"errors"
	"sync"
	"time"
)

type Job interface {
	// GetID Used for keeping track which result maps to which job
	// This could also be changed to have jobs just be an empty interface and then leaving tracking
	// to the consumer
	GetID() string
}

type JobResult struct {
	JobID string
	// Leave this as an interface for now, the batch processor can have it's own ideas of
	// what goes in the result
	Result interface{}
}

// BatchProcessor provides an interface for Dispatcher to process jobs
type BatchProcessor interface {
	Process(jobs []Job) []JobResult
}

type Config struct {
	BatchSize int           // Number of jobs to run at a time
	Interval  time.Duration // Time between batch processings runs
}

// Dispatcher runs submitted jobs using a give BatchProcessor
type Dispatcher struct {
	processor BatchProcessor
	jobs      []Job
	conf      Config
	stop      chan bool
	results   chan JobResult
	busy      sync.Mutex
	done      bool
}

func NewDispatcher(bp BatchProcessor, conf Config) *Dispatcher {
	return &Dispatcher{
		conf:      conf,
		processor: bp,
		stop:      make(chan bool),
		// Use double the batch size for the returning channel for safety
		// This could be improved by making it configurable or relying on a different mechanism
		// Having results be too small can lead to jobs not being processed in a timely manner if the consumer
		// doesn't empty the channel regularly
		results: make(chan JobResult, conf.BatchSize*2),
	}
}

// Shutdown will stop the operation of sending Jobs to the BatchProcessor
// This function will block until remaining jobs are processed
func (d *Dispatcher) Shutdown() {
	// lock to avoid shutdown from duplicating Jobs
	d.busy.Lock()
	// Set flag to stop more jobs from being submitted
	d.done = true
	// Process any remaining jobs
	for len(d.jobs) > 0 {
		d.process()
	}
	d.busy.Unlock()
	close(d.stop)
	close(d.results)
}

// Run creates a routine that will try to run Jobs at the configured interval
// Consume the returned Channel to receive JobResults as the Jobs are processed
func (d *Dispatcher) Run() chan JobResult {
	tickTock := time.NewTicker(d.conf.Interval)
	// check if the ticker has reached its interval or if the stop signal
	// has been received
	go func() {
		for {
			select {
			case <-tickTock.C:
				d.busy.Lock()
				d.process()
				d.busy.Unlock()
			case <-d.stop:
				tickTock.Stop()
				return
			}
		}
	}()
	return d.results
}

// Submit adds a Job to be handled by the BatchProcessor
func (d *Dispatcher) Submit(j Job) error {
	if d.done {
		return errors.New("Dispatcher is shutting down")
	}

	d.jobs = append(d.jobs, j)
	return nil
}

// process sends Jobs to the BatchProcessor
func (d *Dispatcher) process() {
	l := len(d.jobs)
	if l == 0 {
		return
	}

	if l > d.conf.BatchSize {
		l = d.conf.BatchSize
	}

	// Process up to BatchSize number of jobs and return the results in a channel
	results := d.processor.Process(d.jobs[:l])
	for _, result := range results {
		d.results <- result
	}

	// Truncate the processes jobs
	d.jobs = d.jobs[l:]
}
