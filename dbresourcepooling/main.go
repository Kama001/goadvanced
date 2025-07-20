package main

import (
	"fmt"
	"sync"
	"test/pool"
	"time"
)

func main() {
	// Example 1: Database connection pool
	fmt.Println("1. Database Connection Pool Example:")
	dbPool := pool.New(3, 5*time.Second)
	defer dbPool.Close()

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn := dbPool.Get()
			if conn != nil {
				defer dbPool.Put(conn)
				fmt.Printf("Worker %d: Using connection %d\n", id, conn.ID)
				time.Sleep(4 * time.Second)
			}
		}(i)
	}
	wg.Wait()

}

// package main

// import (
// 	"fmt"
// 	"time"
// )

// func main() {
// 	fmt.Println(time.After(5 * time.Second))
// }
