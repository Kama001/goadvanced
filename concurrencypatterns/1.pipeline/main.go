package main

import (
	"fmt"
	"math/rand"
)

func generateRandom() <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; i < 5; i++ {
			ch <- rand.Intn(100)
		}
	}()
	return ch
}
func square(input <-chan int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for num := range input {
			ch <- (num * num)
		}
	}()
	return ch
}
func addSomething(input <-chan int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for num := range input {
			ch <- num + 10
		}
	}()
	return ch
}
func main() {
	stage1 := generateRandom()
	stage2 := square(stage1)
	stage3 := addSomething(stage2)
	for val := range stage3 {
		fmt.Println(val)
	}
}

// Logging (Envoy)
// Enable exact log level in envoy server
// Build required envoy docker image
// Integrate envoy logs with sflogs

// Log information between client to AWS VPC
// Dependency on network/cloud team

// Monitoring (envoy)
// Uptime status of envoy servers to be displayed in grafana
// Required metrics of envoy servers to be pushed to grafana

// Pipeline
// Necessary access to Jenkins or gitlab for pipeline creation
// Need ecus help to understand how to run astree analysis on required modules manually
// Automate the analysis, and build a pipeline
// Enable appropriate logging to know the issues
// Create appropriate reports to compare

// Test Scenarios
// Load Test Examples
// 50 concurrent analysis -> need to be completed within given time
// High throughput: send max analysis in 60 sec -> zero requests dropped, acceptable latency

// Dependencies:
// cloud infra team
// network team
// Developers from ECUs
