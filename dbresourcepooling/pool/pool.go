// package pool

// import (
// 	"fmt"
// 	"sync"
// 	"time"
// )

// type DBConnection struct {
// 	ID        string
// 	CreatedAt time.Time
// 	LastUsed  time.Time
// }

// type DBConnectionPool struct {
// 	connections chan *DBConnection
// 	maxSize     int
// 	timeout     time.Duration
// 	closed      bool
// 	mu          sync.Mutex
// }

// func NewDBConnectionPool(maxSize int, timeout time.Duration) *DBConnectionPool {
// 	pool := &DBConnectionPool{
// 		connections: make(chan *DBConnection, maxSize),
// 		maxSize:     maxSize,
// 		timeout:     timeout,
// 	}

// 	// Pre-populate the pool
// 	for i := 0; i < maxSize; i++ {
// 		conn := &DBConnection{
// 			ID:        fmt.Sprintf("db-conn-%d", i+1),
// 			CreatedAt: time.Now(),
// 			LastUsed:  time.Now(),
// 		}
// 		pool.connections <- conn
// 	}

// 	return pool
// }

// func (p *DBConnectionPool) Get() *DBConnection {
// 	select {
// 	case conn := <-p.connections:
// 		conn.LastUsed = time.Now()
// 		fmt.Printf("  Acquired connection: %s\n", conn.ID)
// 		return conn
// 	case <-time.After(p.timeout):
// 		fmt.Println("  Timeout waiting for connection")
// 		return nil
// 	}
// }

// func (p *DBConnectionPool) Put(conn *DBConnection) {
// 	if conn == nil {
// 		return
// 	}

// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	if p.closed {
// 		return
// 	}

// 	// Check if connection is still healthy
// 	if time.Since(conn.LastUsed) > p.timeout {
// 		fmt.Printf("  Discarding stale connection: %s\n", conn.ID)
// 		return
// 	}

// 	select {
// 	case p.connections <- conn:
// 		fmt.Printf("  Returned connection: %s\n", conn.ID)
// 	default:
// 		fmt.Printf("  Pool full, discarding connection: %s\n", conn.ID)
// 	}
// }

// func (p *DBConnectionPool) Close() {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	p.closed = true
// 	close(p.connections)
// }

package pool

import (
	"fmt"
	"sync"
	"time"
)

type DBConnection struct {
	ID        int
	CreatedAt time.Time
	LastUsed  time.Time
}

type DBConnectionPool struct {
	mu          sync.Mutex
	connections chan *DBConnection
	closed      bool
	timeout     time.Duration
	maxSize     int
}

func New(maxSize int, timeout time.Duration) *DBConnectionPool {
	p := DBConnectionPool{
		timeout:     timeout,
		connections: make(chan *DBConnection, maxSize),
		maxSize:     maxSize,
	}
	for i := 0; i < maxSize; i++ {
		dbConn := DBConnection{
			ID:        i,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}
		p.connections <- &dbConn
	}
	return &p
}

func (p *DBConnectionPool) Get() *DBConnection {
	select {
	case conn := <-p.connections:
		conn.LastUsed = time.Now()
		fmt.Printf("Acquired connection: %d\n", conn.ID)
		return conn
	// waits for p.timeout and sends cur
	// time to a channel
	case <-time.After(p.timeout):
		fmt.Println("timeout waiting for connection")
		return nil
	}
}

func (p *DBConnectionPool) Put(conn *DBConnection) {
	if conn == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	if time.Since(conn.LastUsed) > p.timeout {
		newConn := DBConnection{
			ID:        conn.ID,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}
		p.connections <- &newConn
		fmt.Printf("Discarding stale connection: %d\n", conn.ID)
		return
	}
	select {
	case p.connections <- conn:
		fmt.Printf("  Returned connection: %d\n", conn.ID)
	default:
		fmt.Printf("  Pool full, discarding connection: %d\n", conn.ID)
	}
}

func (p *DBConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	close(p.connections)
}
