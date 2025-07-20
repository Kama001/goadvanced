package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Resource interface {
	Close() error
	GetID() string
	IsNil() bool
}

type Factory[T Resource] interface {
	Create() (T, error)
	Destroy(T) error
}

type Pool[T Resource] struct {
	mu           sync.Mutex
	resources    chan T
	closed       bool
	factory      Factory[T]
	min          int
	max          int
	currentCount int
	timeout      time.Duration
}

type DBConnection struct {
	ID string
}

func (conn *DBConnection) Close() error {
	fmt.Printf("connection is closed for %s\n", conn.ID)
	return nil
}

func (conn *DBConnection) GetID() string {
	return conn.ID
}

func (conn *DBConnection) IsNil() bool {
	if conn == nil {
		return true
	} else {
		return false
	}
}

type DBFactory struct {
	counter int
	mu      sync.Mutex
}

func (f *DBFactory) Create() (*DBConnection, error) {
	f.mu.Lock()
	f.counter++
	f.mu.Unlock()
	fmt.Printf("Creating DB connection %d\n", f.counter)
	conn := &DBConnection{
		ID: fmt.Sprintf("%d", f.counter),
	}
	return conn, nil
}

func (f *DBFactory) Destroy(conn *DBConnection) error {
	fmt.Println("Destroying db connection")
	conn.Close()
	return nil
}

func New[T Resource](min, max int, factory Factory[T], timeout time.Duration) (*Pool[T], error) {
	p := &Pool[T]{
		resources: make(chan T, max),
		factory:   factory,
		min:       min,
		timeout:   timeout,
	}
	for i := 0; i < p.min; i++ {
		res, err := p.factory.Create()
		if err != nil {
			fmt.Println("cannot add resource!!!")
			close(p.resources)
			return nil, fmt.Errorf("error creating pool")
		}

		p.resources <- res
	}
	return p, nil
}

func (p *Pool[T]) Get(ctx context.Context) (T, error) {
	var zero T
	select {
	case res, ok := <-p.resources:
		if !ok {
			return zero, fmt.Errorf("Pool closed")
		}
		return res, nil
	case <-ctx.Done():
		return zero, ctx.Err()
	default:
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return zero, fmt.Errorf("Pool closed")
		}
		p.mu.Unlock()
		if p.currentCount < p.max {
			p.inc()
			fmt.Println("Pool: No resource available, creating new one...")
			res, err := p.factory.Create()
			if err != nil {
				p.dec()
				return zero, fmt.Errorf("cannot create resource")
			}
			return res, nil
		}
		select {
		case res, ok := <-p.resources:
			if !ok {
				return zero, fmt.Errorf("pool: pool is closed while waiting")
			}
			return res, nil
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(p.timeout):
			return zero, fmt.Errorf("timed out waiting for resources")
		}
	}
}

func (p *Pool[T]) inc() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentCount++

}

func (p *Pool[T]) dec() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentCount--
}

func (p *Pool[T]) Put(res T) {
	if res.IsNil() {
		fmt.Println("nil resource received!!!")
		p.factory.Destroy(res)
		p.dec()
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		p.mu.Unlock()
		p.factory.Destroy(res)
		fmt.Printf("pool closed, destroying resource %s\n", res.GetID())
		p.dec()
		return
	}
	select {
	case p.resources <- res:
		// fmt.Printf("resource %s returned\n", res.GetID())
	default:
		fmt.Printf("pool full, destroying resource %s\n", res.GetID())
		p.factory.Destroy(res)
		p.dec()
		return
	}
}

func (p *Pool[T]) Close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	close(p.resources)
	for res := range p.resources {
		fmt.Printf("destroying resource %s\n", res.GetID())
		p.factory.Destroy(res)
		p.dec()
	}
}

func main() {
	dbFactory := &DBFactory{}
	dbPool, err := New[*DBConnection](2, 5, dbFactory, 3*time.Second)
	if err != nil {
		fmt.Printf("error creating pool %s", err.Error())
	}
	defer dbPool.Close()
	ctx := context.Background()
	for i := 0; i < 7; i++ { // Try to get more than max
		go func(i int) {
			conn, err := dbPool.Get(ctx)
			if err != nil {
				fmt.Printf("Goroutine %d: Failed to get connection: %v\n", i, err)
				return
			}
			fmt.Printf("Goroutine %d: Got connection %s\n", i, conn.GetID())
			time.Sleep(3 * time.Second) // Simulate work
			dbPool.Put(conn)
			fmt.Printf("Goroutine %d: Put connection %s back\n", i, conn.GetID())
		}(i)
	}
	time.Sleep(10 * time.Second)
}

// // only one new resource is created
// // when Get is called, how to create based load

// =============================================================================
// =============================================================================
// =============================================================================
// =============================================================================

// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"
// )

// // Resource interface defines the contract for pool-managed objects.
// type Resource interface {
// 	Close() error
// 	GetID() string
// 	IsNil() bool // Useful for validating resources, especially generic types
// 	// Optional: Add a HealthCheck() error method for robustness
// 	// HealthCheck() error
// }

// // Factory interface defines how to create and destroy resources.
// type Factory[T Resource] interface {
// 	Create() (T, error)
// 	Destroy(T) error
// }

// // Pool manages a collection of resources.
// type Pool[T Resource] struct {
// 	mu        sync.Mutex
// 	resources chan T // Channel for available resources
// 	factory   Factory[T]
// 	min       int
// 	max       int
// 	idleTimeout time.Duration // New: Time after which idle resources are destroyed

// 	// Internal state
// 	currentTotal int       // New: Tracks total resources (in-use + available)
// 	closed       bool
// 	done         chan struct{} // New: Signal for background goroutines to stop
// }

// // DBConnection implements the Resource interface.
// type DBConnection struct {
// 	ID string
// }

// func (conn *DBConnection) Close() error {
// 	fmt.Printf("DBConnection %s: closed.\n", conn.ID)
// 	return nil
// }

// func (conn *DBConnection) GetID() string {
// 	return conn.ID
// }

// func (conn *DBConnection) IsNil() bool {
// 	return conn == nil // Simplified
// }

// // DBFactory implements the Factory interface for DBConnections.
// type DBFactory struct {
// 	counter int
// 	mu      sync.Mutex // Protect counter for concurrent Create calls
// }

// func (f *DBFactory) Create() (*DBConnection, error) {
// 	f.mu.Lock()
// 	f.counter++
// 	id := f.counter
// 	f.mu.Unlock()

// 	fmt.Printf("DBFactory: creating DB connection %d...\n", id)
// 	time.Sleep(100 * time.Millisecond) // Simulate work
// 	return &DBConnection{ID: fmt.Sprintf("%d", id)}, nil
// }

// func (f *DBFactory) Destroy(conn *DBConnection) error {
// 	if conn == nil || conn.IsNil() { // Handle nil resources safely
// 		return nil
// 	}
// 	fmt.Printf("DBFactory: destroying DB connection %s...\n", conn.ID)
// 	return conn.Close()
// }

// // New creates a new resource pool.
// func New[T Resource](min, max int, idleTimeout time.Duration, factory Factory[T]) (*Pool[T], error) {
// 	if min < 0 || max <= 0 || min > max {
// 		return nil, fmt.Errorf("pool: invalid min/max parameters: min=%d, max=%d", min, max)
// 	}
// 	if idleTimeout <= 0 && min != max { // If min != max, we need timeout to shrink
// 		return nil, fmt.Errorf("pool: idleTimeout must be positive if min != max")
// 	}

// 	p := &Pool[T]{
// 		resources:    make(chan T, max),
// 		factory:      factory,
// 		min:          min,
// 		max:          max,
// 		idleTimeout:  idleTimeout,
// 		currentTotal: 0,
// 		closed:       false,
// 		done:         make(chan struct{}),
// 	}

// 	// Initialize with min resources
// 	for i := 0; i < min; i++ {
// 		res, err := p.factory.Create()
// 		if err != nil {
// 			// Clean up any resources already created if initialization fails
// 			p.Close() // Call Close to drain and destroy already created resources
// 			return nil, fmt.Errorf("pool: failed to create initial resource %d: %w", i+1, err)
// 		}
// 		p.resources <- res
// 		p.currentTotal++ // Increment total count after successful creation
// 	}

// 	// Start background goroutine for idle resource management
// 	if p.min < p.max && p.idleTimeout > 0 {
// 		go p.reaper()
// 	}

// 	return p, nil
// }

// // Get retrieves a resource from the pool.
// func (p *Pool[T]) Get(ctx context.Context) (T, error) {
// 	var zero T // Zero value for T

// 	select {
// 	case res, ok := <-p.resources:
// 		if !ok {
// 			return zero, fmt.Errorf("pool: pool is closed")
// 		}
// 		// Optional: Health check on resource before returning
// 		// if healthCheckable, isHealthCheckable := interface{}(res).(interface{ HealthCheck() error }); isHealthCheckable {
// 		//     if err := healthCheckable.HealthCheck(); err != nil {
// 		//         p.mu.Lock()
// 		//         p.factory.Destroy(res) // Destroy unhealthy resource
// 		//         p.currentTotal--
// 		//         p.mu.Unlock()
// 		//         fmt.Printf("Pool: Destroyed unhealthy resource %s, trying to get another.\n", res.GetID())
// 		//         return p.Get(ctx) // Try getting another one
// 		//     }
// 		// }
// 		return res, nil
// 	case <-ctx.Done():
// 		return zero, ctx.Err() // Return context error, don't close pool
// 	default:
// 		// No resource immediately available, try to create a new one if below max
// 		p.mu.Lock()
// 		if p.closed {
// 			p.mu.Unlock()
// 			return zero, fmt.Errorf("pool: pool is closed")
// 		}
// 		if p.currentTotal < p.max {
// 			p.currentTotal++ // Optimistically increment total before creation
// 			p.mu.Unlock()

// 			fmt.Println("Pool: No resource available, creating new one...")
// 			res, err := p.factory.Create()
// 			if err != nil {
// 				p.mu.Lock()
// 				p.currentTotal-- // Decrement if creation failed
// 				p.mu.Unlock()
// 				return zero, fmt.Errorf("pool: failed to create new resource: %w", err)
// 			}
// 			return res, nil // Return the newly created resource
// 		}
// 		p.mu.Unlock()

// 		// If at max capacity and no resource available, wait on the channel
// 		select {
// 		case res, ok := <-p.resources:
// 			if !ok {
// 				return zero, fmt.Errorf("pool: pool is closed while waiting")
// 			}
// 			return res, nil
// 		case <-ctx.Done():
// 			return zero, ctx.Err()
// 		}
// 	}
// }

// // Put returns a resource to the pool.
// func (p *Pool[T]) Put(res T) {
// 	if res.IsNil() {
// 		fmt.Println("Pool: received empty resource, destroying it.")
// 		// We still destroy it if it's nil, as it might have been an error-state resource.
// 		p.factory.Destroy(res)
// 		p.mu.Lock()
// 		p.currentTotal-- // Decrement if we just destroyed a nil resource
// 		p.mu.Unlock()
// 		return
// 	}

// 	p.mu.Lock()
// 	if p.closed {
// 		p.mu.Unlock()
// 		fmt.Printf("Pool: pool is closed, destroying resource %s.\n", res.GetID())
// 		p.factory.Destroy(res)
// 		p.mu.Lock()
// 		p.currentTotal-- // Decrement if we just destroyed a resource because pool is closed
// 		p.mu.Unlock()
// 		return
// 	}
// 	p.mu.Unlock()

// 	select {
// 	case p.resources <- res:
// 		fmt.Printf("Pool: resource %s returned.\n", res.GetID())
// 	default: // Pool channel is full
// 		fmt.Printf("Pool: channel full, destroying resource %s.\n", res.GetID())
// 		p.mu.Lock()
// 		p.currentTotal-- // Decrement total before destroying
// 		p.mu.Unlock()
// 		p.factory.Destroy(res) // Destroy if pool is full (exceeds max available)
// 	}
// }

// // Close closes the pool and destroys all resources.
// func (p *Pool[T]) Close() {
// 	p.mu.Lock()
// 	if p.closed {
// 		p.mu.Unlock()
// 		return // Already closed
// 	}
// 	p.closed = true
// 	close(p.done) // Signal reaper to stop
// 	p.mu.Unlock()

// 	// Drain and destroy all resources
// 	for {
// 		select {
// 		case res, ok := <-p.resources:
// 			if !ok { // Channel is empty and closed
// 				fmt.Println("Pool: All resources drained and destroyed.")
// 				return
// 			}
// 			p.factory.Destroy(res)
// 			p.mu.Lock()
// 			p.currentTotal-- // Decrement count as resources are destroyed
// 			p.mu.Unlock()
// 		default: // Channel is empty (no more resources to drain)
// 			fmt.Println("Pool: All resources drained (or none left).")
// 			return
// 		}
// 	}
// }

// // reaper periodically destroys idle resources if currentTotal > min
// func (p *Pool[T]) reaper() {
// 	ticker := time.NewTicker(p.idleTimeout / 2) // Check periodically
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			p.mu.Lock()
// 			if p.closed {
// 				p.mu.Unlock()
// 				return
// 			}

// 			// Shrink if currentTotal > min and there are more resources available than min
// 			for len(p.resources) > p.min && p.currentTotal > p.min {
// 				select {
// 				case res := <-p.resources: // Try to get an idle resource
// 					fmt.Printf("Pool Reaper: Destroying idle resource %s (pool size %d, min %d)\n", res.GetID(), p.currentTotal, p.min)
// 					p.factory.Destroy(res)
// 					p.currentTotal--
// 				default:
// 					// No more idle resources to take from channel right now
// 					break
// 				}
// 			}
// 			p.mu.Unlock()
// 		case <-p.done:
// 			fmt.Println("Pool Reaper: Stopping.")
// 			return
// 		}
// 	}
// }

// func main() {
// 	factory := &DBFactory{}
// 	// Create a pool with min 2, max 5 connections, and an idle timeout of 5 seconds
// 	pool, err := New[*DBConnection](2, 5, 5*time.Second, factory)
// 	if err != nil {
// 		log.Fatalf("Failed to create pool: %v", err)
// 	}
// 	defer pool.Close() // Ensure pool is closed on exit

// 	// Example usage
// 	ctx := context.Background()

// 	// Get a few resources
// 	for i := 0; i < 7; i++ { // Try to get more than max
// 		go func(i int) {
// 			conn, err := pool.Get(ctx)
// 			if err != nil {
// 				fmt.Printf("Goroutine %d: Failed to get connection: %v\n", i, err)
// 				return
// 			}
// 			fmt.Printf("Goroutine %d: Got connection %s\n", i, conn.GetID())
// 			time.Sleep(2 * time.Second) // Simulate work
// 			pool.Put(conn)
// 			fmt.Printf("Goroutine %d: Put connection %s back\n", i, conn.GetID())
// 		}(i)
// 	}

// 	time.Sleep(15 * time.Second) // Let operations and reaper run
// }
