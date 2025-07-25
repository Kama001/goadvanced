// https://compositecode.blog/2025/06/25/go-concurrency-patternsproducer-consumer-pattern/

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	bufferSize := 5
	numProducers := 2
	numConsumers := 3
	numItems := 10

	buffer := make(chan int, bufferSize)

	// start producers
	var wg sync.WaitGroup
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < numItems; i++ {
				item := rand.Intn(100)
				buffer <- item
				fmt.Printf("Producer %d produced: %d\n", id, item)
				time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)
			}
		}(i)
	}

	// start consumers
	var consumerWg sync.WaitGroup
	for i := 0; i < numConsumers; i++ {
		consumerWg.Add(1)
		go func(id int) {
			defer consumerWg.Done()
			for item := range buffer {
				fmt.Printf("Consumer %d consumed: %d\n", id, item)
				time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)
			}
		}(i)
	}
	// Wait for all producers to finish, then close the buffer
	wg.Wait()
	close(buffer)
	// Wait for all consumers to finish
	consumerWg.Wait()
}
