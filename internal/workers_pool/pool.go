package workers_pool

import (
	"runtime"
	"sync"
)

// StartWorkers start N number of workers, if workers is 0 or negative start a worker for each CPU count.
func StartWorkers[T any, V any](jobQueue chan T, resultChannel chan V, worker func(id int, wg *sync.WaitGroup, jobQueue chan T, resultChannel chan V)) *sync.WaitGroup {
	count := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	for i := range count {
		wg.Add(1)
		go worker(i, wg, jobQueue, resultChannel)
	}

	return wg
}

// LoadJobs Iterate through a slice and every item into a job queue.  Optional validate function to filter out data
func LoadJobs[S ~[]E, E comparable](inputQueue chan E, inputData S, validate func(E) bool) {
	for _, item := range inputData {
		if validate != nil && validate(item) {
			inputQueue <- item
		}
	}
	close(inputQueue)

}

// GetResults waits fo all tasks to be done, closes results channel, and return the data from the channel
func GetResults[T any](wg *sync.WaitGroup, results chan T) []T {
	wg.Wait()
	close(results)

	var data []T
	for val := range results {
		data = append(data, val)
	}
	return data
}
