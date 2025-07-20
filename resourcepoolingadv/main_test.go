package main

import (
	"context"
	"testing"
	"time"
)

func TestPool_BasicUsage(t *testing.T) {
	factory := &DBFactory{}
	pool, err := New[*DBConnection](2, 5, factory, 3*time.Second)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	ctx := context.Background()

	// Get two resources (initially created)
	res1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get resource 1: %v", err)
	}
	t.Logf("Got resource: %s", res1.GetID())

	res2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get resource 2: %v", err)
	}
	t.Logf("Got resource: %s", res2.GetID())

	// Put back one resource
	pool.Put(res1)
	t.Logf("Put back resource: %s", res1.GetID())

	// Get again (should reuse)
	res3, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get resource 3: %v", err)
	}
	t.Logf("Got resource: %s", res3.GetID())

	// Create new ones until limit
	for i := 0; i < 3; i++ {
		go func(i int) {
			res, err := pool.Get(ctx)
			if err != nil {
				t.Fatalf("Failed to get resource %d: %v", i+4, err)
			}
			t.Logf("Got new resource: %s", res.GetID())
		}(i)
	}

	// This call should block or fail due to reaching max, so use timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	_, err = pool.Get(ctxTimeout)
	if err == nil {
		t.Error("Expected error or timeout due to max pool size")
	} else {
		t.Logf("Correctly failed to get resource due to max limit: %v", err)
	}

	// Clean up
	pool.Put(res2)
	pool.Put(res3)
	pool.Close()
}
