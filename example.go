package main

import (
	"batch/pkg/batch"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type exampleBatchProcessor struct {
}

func (e *exampleBatchProcessor) Process(jobs []batch.Job) []batch.JobResult {
	results := make([]batch.JobResult, 0)

	for _, job := range jobs {
		results = append(results, batch.JobResult{
			JobID:  job.GetID(),
			Result: fmt.Sprintf("I am the result for job with id %s", job.GetID()),
		})
	}

	return results
}

type exampleJob struct {
	id string
}

func (e *exampleJob) GetID() string {
	return e.id
}

func main() {
	bp := exampleBatchProcessor{}
	dispatcher := batch.NewDispatcher(&bp, batch.Config{
		BatchSize: 10,
		Interval:  time.Millisecond * 200,
	})

	results := dispatcher.Run()

	for k := 0; k < 100; k++ {
		dispatcher.Submit(&exampleJob{id: uuid.NewString()})
	}

	go func() {
		time.Sleep(time.Second * 5)
		dispatcher.Shutdown()
	}()

	for result := range results {
		fmt.Printf("%+v\n", result)
	}
}
