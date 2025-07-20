package main

import (
	"context"
	"log"
	"math/rand"
	"resourcepooling/pool"
	"sync"
	"time"
)

type DBConnection struct {
	ID int
}

func (c *DBConnection) Close() error {
	log.Printf("Closing connection %d\n", c.ID)
	return nil
}

func factory() (*DBConnection, error) {
	conn := &DBConnection{ID: rand.Intn(1000)}
	log.Printf("Creating new connection %d\n", conn.ID)
	return conn, nil
}
func main() {
	p, err := pool.New(factory, 5, 20)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	log.Printf("✅ Pool created with initial size: %d, capacity: 20", p.Len())
	var wg sync.WaitGroup
	numworkers := 30
	for i := 0; i < numworkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			log.Printf("Worker %d trying to get a connection...\n", workerID)
			conn, err := p.Get(ctx)
			if err != nil {
				log.Printf("❗️ Worker %d failed to get connection: %v\n", workerID, err)
				return
			}
			log.Printf("✅ Worker %d acquired connection %d. Pool size: %d", workerID, conn.ID, p.Len())

			// Simulate doing work with the connection
			time.Sleep(time.Duration(500+rand.Intn(500)) * time.Millisecond)

			// Return the connection to the pool
			log.Printf("Worker %d returning connection %d...\n", workerID, conn.ID)
			p.Put(conn)
			log.Printf("➡️ Worker %d returned connection. Pool size: %d", workerID, p.Len())
		}(i)
	}
	wg.Wait()

	// 3. Close the pool
	log.Println("All workers finished. Closing the pool...")
	p.Close()
	log.Println("✅ Pool closed.")

	// 4. Demonstrate that operations on a closed pool fail
	_, err = p.Get(context.Background())
	if err != nil {
		log.Printf("Attempted Get on closed pool, got expected error: %v\n", err)
	}

	log.Println("Demonstration complete. 🎉")
}
