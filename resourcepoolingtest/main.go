package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// ErrPoolClosed is returned when an action is attempted on a closed pool.
var ErrPoolClosed = errors.New("resource pool is closed")
var ErrStaleResource = errors.New("removing found stale resource")
var ErrFactoryCreation = errors.New("error creating new factory")
var ErrMaxResources = errors.New("maximum number of resources reached")

type pooledResource struct {
	ioCloser io.Closer
	lastUsed time.Time
}

// Pool manages a set of resources that can be shared.
// The resource type must implement the io.Closer interface.
type Pool struct {
	lock        sync.Mutex
	resources   chan *pooledResource // Changed to hold the wrapped resource
	factory     func() (io.Closer, error)
	closed      bool
	maxLifetime time.Duration         // New field for max resource lifetime
	healthCheck func(io.Closer) error // New: Function to check resource health
	maxOpen     int                   // New: Max number of total resources
	numOpen     int                   // New: Current number of total resources
}
type PoolStats struct {
	NumOpen int
	Idle    int
	MaxOpen int
	InUse   int
}

func (p *Pool) Stats() PoolStats {
	poolStats := PoolStats{}
	poolStats.NumOpen = p.numOpen
	poolStats.MaxOpen = p.maxOpen
	poolStats.Idle = p.maxOpen - p.numOpen
	poolStats.InUse = p.numOpen
	return poolStats
}

func (p *Pool) NewResource() (*pooledResource, error) {
	ioCloser, err := p.factory()
	if err != nil {
		return nil, ErrFactoryCreation
	}
	return &pooledResource{
		ioCloser: ioCloser,
		lastUsed: time.Now(),
	}, nil
}

func (p *Pool) inc() {
	p.lock.Lock()
	p.numOpen++
	p.lock.Unlock()
}
func (p *Pool) dec() {
	p.lock.Lock()
	p.numOpen--
	p.lock.Unlock()
}

func New(factory func() (io.Closer, error),
	size uint,
	maxLifetime time.Duration,
	maxOpen int,
	healthCheck func(io.Closer) error) (*Pool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("cannot create channel with size < 0")
	}
	p := Pool{
		resources:   make(chan *pooledResource, size),
		factory:     factory,
		maxLifetime: maxLifetime,
		healthCheck: healthCheck,
		maxOpen:     maxOpen,
	}
	for i := uint(0); i < size; i++ {
		res, err := p.NewResource()
		if err != nil {
			p.Shutdown()
		}
		p.resources <- res
		p.inc()
	}
	return &p, nil
}

func (p *Pool) Get(ctx context.Context) (io.Closer, error) {
	p.lock.Lock()
	if p.closed {
		p.lock.Unlock()
		fmt.Printf("%s", ErrPoolClosed)
		return nil, ErrPoolClosed
	}
	p.lock.Unlock()
	for {
		select {
		case res, ok := <-p.resources:
			if !ok {
				return nil, ErrPoolClosed
			}
			if time.Since(res.lastUsed) > p.maxLifetime {
				res.ioCloser.Close()
				p.dec()
				continue
			}
			if err := p.healthCheck(res.ioCloser); err != nil {
				res.ioCloser.Close()
				p.dec()
				continue
			}
			return res.ioCloser, nil
		default:
			fmt.Println("resource is not available, trying to create new one")
		}
		p.lock.Lock()
		if p.numOpen < p.maxOpen {
			p.lock.Unlock()
			res, err := p.NewResource()
			if err != nil {
				return nil, err
			}
			p.resources <- res
			p.inc()
			return res.ioCloser, nil
		}
		p.lock.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res, ok := <-p.resources:
			if !ok {
				return nil, ErrPoolClosed
			}
			if time.Since(res.lastUsed) > p.maxLifetime {
				res.ioCloser.Close()
				p.dec()
				continue
			}
			if err := p.healthCheck(res.ioCloser); err != nil {
				res.ioCloser.Close()
				p.dec()
				continue
			}
			return res.ioCloser, nil
		}
	}
}

func (p *Pool) Put(resource io.Closer) {
	p.lock.Lock()
	if p.closed {
		p.lock.Unlock()
		fmt.Printf("%s", ErrPoolClosed)
		p.dec()
		resource.Close()
		return
	}
	p.lock.Unlock()
	select {
	case p.resources <- &pooledResource{resource, time.Now()}:
		fmt.Println("successfully returned the resource")
	default:
		fmt.Println("pool is full, closing the resource")
		p.dec()
		resource.Close()
	}
}

func (p *Pool) Shutdown() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.closed = true
	close(p.resources)
	for res := range p.resources {
		res.ioCloser.Close()
	}
	p.numOpen = 0
	p.maxOpen = 0
}

func (p *Pool) Len() int {
	return len(p.resources)
}

// Example Usage (for your reference once you're done)
func main() {
	// This part is just to demonstrate how the pool would be used.
	// You don't need to implement this.
}
