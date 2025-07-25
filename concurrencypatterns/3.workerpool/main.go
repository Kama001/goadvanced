package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(jobs <-chan string, workers int, results chan<- string, wg *sync.WaitGroup) {
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				time.Sleep(2 * time.Second)
				results <- fmt.Sprintf("%s completed at %v", job, time.Now().Format("2000-01-06"))
			}
		}()
	}
}
func main() {
	numWorkers := 3
	numJobs := 15
	var wg sync.WaitGroup
	jobsChannel := make(chan string)
	go func() {
		defer close(jobsChannel)
		for i := 0; i < numJobs; i++ {
			jobsChannel <- fmt.Sprintf("jobid: %d", i)
		}
	}()
	resultsChannel := make(chan string)
	worker(jobsChannel, numWorkers, resultsChannel, &wg)
	go func() {
		wg.Wait()
		close(resultsChannel)
	}()
	for results := range resultsChannel {
		fmt.Println(results)
	}
}
