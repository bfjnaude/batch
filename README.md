# batch
Micro Batch processor written in Go.

# Usage

Batch processing is performed by creating a `Dispatcher` with a configured `Interval` and `BatchSize`.

Once per `Interval`, a `BatchProcessor` will be called with at most `BatchSize` number of `Job`s

`JobResult`s are received via a channel, which is returned when starting the `Run` process of the `Dispatcher`. The channel should be read frequentlty.

Call `Shutdown` on the `Dispatcher` to terminate
the batch process. All previously submitted `Job`s
will be completed before the shutdown completes.

See [example.go](example.go)

```golang
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
```

## Commands

| command | description |
| - | - |
| `make example`     | Runs the example program |
| `make test`        | Runs unit tests |
| `make generate`    | Generates mocks use in testing |
| `make vendor`      | Vendors third party code |

