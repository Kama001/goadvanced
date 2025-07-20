package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	ch := make(chan struct{})
	go func() {
		time.Sleep(2 * time.Second)
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		fmt.Println("got value before context exceeded!!")
	case <-ctxWithTimeout.Done():
		fmt.Println("oops context execeeded", ctxWithTimeout.Err())
	}
}
