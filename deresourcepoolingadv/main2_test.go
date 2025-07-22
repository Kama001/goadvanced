package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	dbFactory := &DBFactory{}
	pool, err := New[*DBConnection](dbFactory, 2, 3*time.Second, 5, 20*time.Second, 5*time.Second)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("pool size is", pool.Len())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count := 0
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res, err := pool.Get(ctx)
			if err != nil {
				count++
			}
			fmt.Println("pool size is", pool.Len())
			fmt.Printf("go routine %d got the connectiond %d\n", i, res.GetID())
			time.Sleep(10 * time.Second)
			pool.Put(res)
			fmt.Printf("go routine %d put the connectiond %d\n", i, res.GetID())
		}(i)
	}
	wg.Wait()
	fmt.Println("pool size is", pool.Len())
	if count == pool.Len() {
		t.Logf("pool sizes matched")
	} else {
		t.Fatalf("pool sizes didn't match")
	}
}
