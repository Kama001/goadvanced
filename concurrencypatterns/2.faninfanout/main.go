package main

import (
	"fmt"
	"sync"
	"time"
)

type WorkItem struct {
	ID   int
	Data string
}

type Result struct {
	OriginalID int
	Processed  string
	WorkerID   int
}

func generateWorkItems(n int) <-chan WorkItem {
	workItems := make(chan WorkItem)
	go func() {
		defer close(workItems)
		for i := 0; i < n; i++ {

			workItems <- WorkItem{ID: i, Data: fmt.Sprintf("data-%d", i)}
		}
	}()
	return workItems
}

func fanOut(jobs <-chan WorkItem, workers int) []<-chan Result {
	var results []chan Result
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		workerResults := make(chan Result)
		results = append(results, workerResults)
		wg.Add(1)
		go worker(i+1, jobs, workerResults, &wg)
	}
	go func() {
		wg.Wait()
		for _, workerResult := range results {
			close(workerResult)
		}
	}()
	var resultsChannel []<-chan Result
	for _, result := range results {
		resultsChannel = append(resultsChannel, result)
	}
	return resultsChannel
}

func worker(workerID int, jobs <-chan WorkItem, workerResults chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		time.Sleep(2 * time.Second)
		workerResults <- Result{
			Processed:  fmt.Sprintf("processed-%s-by-worker-%d", job.Data, workerID),
			WorkerID:   workerID,
			OriginalID: job.ID,
		}
	}
}

func fanIn(resultsChannel []<-chan Result) <-chan Result {
	finalResults := make(chan Result)

	var wg sync.WaitGroup
	wg.Add(len(resultsChannel))
	for _, results := range resultsChannel {
		go func() {
			defer wg.Done()
			for result := range results {
				finalResults <- result
			}
		}()
	}
	go func() {
		wg.Wait()
		close(finalResults)
	}()
	return finalResults
}

func main() {
	workerItems := generateWorkItems(20)
	resultsChannel := fanOut(workerItems, 4)
	finalResults := fanIn(resultsChannel)
	var count int
	for result := range finalResults {
		fmt.Printf("Processed: Item %d -> %s (by Worker %d)\n", result.OriginalID, result.Processed, result.WorkerID)
		count++
	}
	fmt.Printf("\nFan-out/Fan-in completed! Processed %d items.\n", count)
}
