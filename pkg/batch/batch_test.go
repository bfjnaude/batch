package batch

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_Dispatcher_LifeTime(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	batches := 2
	bp := NewMockBatchProcessor(ctrl)
	bp.EXPECT().Process(gomock.Any()).DoAndReturn(func(jobs []Job) []JobResult {
		results := make([]JobResult, 0)
		for _, j := range jobs {
			results = append(results, JobResult{
				JobID: j.GetID(),
			})
		}
		return results
	},
	).Times(batches)

	batchSize := 5
	interval := time.Second * 2
	d := NewDispatcher(bp, Config{
		BatchSize: batchSize,
		Interval:  interval,
	})

	for k := 0; k < batchSize*batches; k++ {
		job := NewMockJob(ctrl)
		job.EXPECT().GetID().Return(uuid.NewString())
		d.Submit(job)
	}

	// When
	resultChan := d.Run()
	var wg sync.WaitGroup
	wg.Add(1)

	results := make([]JobResult, 0)
	go func() {
		for result := range resultChan {
			results = append(results, result)
			assert.NotEmpty(t, result)
		}
		wg.Done()
	}()

	time.Sleep(time.Second * 5)
	d.Shutdown()
	wg.Wait()

	// Then
	assert.Equal(t, len(results), batchSize*batches)
}

// Test that Dispatcher correctly processes a single job
func Test_Dispatcher_ProcessSingleJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bp := NewMockBatchProcessor(ctrl)
	batchSize := 1
	interval := time.Millisecond * 100
	d := NewDispatcher(bp, Config{
		BatchSize: batchSize,
		Interval:  interval,
	})

	job := NewMockJob(ctrl)
	jobID := uuid.NewString()
	// Expect GetID to be called, but the exact number of times is determined by the bp.Process call
	job.EXPECT().GetID().Return(jobID).AnyTimes()

	// Use DoAndReturn to explicitly call job.GetID() when bp.Process is invoked
	bp.EXPECT().Process(gomock.Any()).DoAndReturn(func(jobs []Job) []JobResult {
		// Simulate processing logic where job.GetID() is called
		for _, job := range jobs {
			_ = job.GetID() // Simulate the call to GetID within the processing logic
		}
		return []JobResult{
			{JobID: jobID},
		}
	}).Times(1)

	resultChan := d.Run()
	d.Submit(job)

	select {
	case result := <-resultChan:
		assert.Equal(t, jobID, result.JobID)
	case <-time.After(time.Second * 1):
		t.Fatal("Expected result was not received in time")
	}

	d.Shutdown()
}

// Test Dispatcher Submit after Shutdown
func Test_Dispatcher_SubmitAfterShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bp := NewMockBatchProcessor(ctrl)
	d := NewDispatcher(bp, Config{
		BatchSize: 1,
		Interval:  time.Millisecond * 100,
	})

	d.Shutdown() // Shutdown before submitting jobs

	job := NewMockJob(ctrl)
	err := d.Submit(job)

	assert.NotNil(t, err, "Expected error when submitting job after shutdown")
}
