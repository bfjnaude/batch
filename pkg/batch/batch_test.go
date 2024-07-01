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

	count := 0
	results := make([]JobResult, 0)
	go func() {
		for result := range resultChan {
			results = append(results, result)
			assert.NotEmpty(t, result)
			count = count + 1
		}
		wg.Done()
	}()

	time.Sleep(time.Second * 5)
	d.Shutdown()
	wg.Wait()

	// Then
	assert.Equal(t, len(results), batchSize*batches)
	assert.Equal(t, count, batchSize*batches)
}
