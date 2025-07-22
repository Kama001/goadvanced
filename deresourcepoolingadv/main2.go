package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// type IDType interface {
// 	int | float32 | float64 | string
// }

type Resource interface {
	Close() error
	GetID() int
	IsNil() bool
	SetLastused(time.Time)
	GetLastused() time.Time
}

type Factory[T Resource] interface {
	Create() (T, error)
	Destroy(T) error
}

type DBConnection struct {
	ID       int
	lastUsed time.Time
}

func (dbc *DBConnection) GetID() int {
	return dbc.ID
}

func (dbc *DBConnection) IsNil() bool {
	if dbc == nil {
		return true
	}
	return false
}

func (dbc *DBConnection) GetLastused() time.Time {
	return dbc.lastUsed
}

func (dbc *DBConnection) SetLastused(t time.Time) {
	dbc.lastUsed = t
}

func (dbc *DBConnection) Close() error {
	fmt.Printf("db connection %d closed\n", dbc.GetID())
	return nil
}

type DBFactory struct {
	counter int
	mu      sync.Mutex
}

func (dbf *DBFactory) Create() (*DBConnection, error) {
	dbf.mu.Lock()
	dbf.counter++
	dbf.mu.Unlock()
	return &DBConnection{
		ID:       dbf.counter,
		lastUsed: time.Now(),
	}, nil
}

func (dbf *DBFactory) Destroy(dbc *DBConnection) error {
	return dbc.Close()
}

type Pool[T Resource] struct {
	closed         bool
	mu             sync.Mutex
	resources      chan T
	factory        Factory[T]
	maxLifetime    time.Duration
	maxSize        int
	curSize        int
	idleTimeout    time.Duration
	reaperInterval time.Duration
	reaperChan     chan struct{}
}

func (p *Pool[T]) NewResource() (T, error) {
	var zero T
	res, err := p.factory.Create()
	if err != nil {
		return zero, fmt.Errorf("resources creation error")
	}
	return res, nil
}

func (p *Pool[T]) incSize(res T) {
	p.mu.Lock()
	p.curSize++
	p.resources <- res
	p.mu.Unlock()
}

func (p *Pool[T]) decSize() {
	p.mu.Lock()
	p.curSize--
	p.mu.Unlock()
}

func New[T Resource](factory Factory[T], size int, maxLifetime time.Duration, maxSize int, reaperInt, idleTimeout time.Duration) (*Pool[T], error) {
	if maxSize < size || size == 0 || maxSize < 0 {
		return nil, fmt.Errorf("invalid pool sizes")
	}
	p := &Pool[T]{
		factory:        factory,
		resources:      make(chan T, size),
		maxLifetime:    maxLifetime,
		maxSize:        maxSize,
		reaperInterval: reaperInt,
		idleTimeout:    idleTimeout,
	}
	for i := 0; i < size; i++ {
		res, err := p.NewResource()
		if err != nil {
			p.Shutdown()
			return nil, fmt.Errorf("pool creation error, %s", err)
		}
		p.incSize(res)
	}
	go p.Reaper(p.reaperChan)
	return p, nil
}

func (p *Pool[T]) Get(ctx context.Context) (T, error) {
	var zero T
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return zero, fmt.Errorf("Error pool closed")
	}
	p.mu.Unlock()
	for {
		select {
		case res, ok := <-p.resources:
			if !ok {
				return zero, fmt.Errorf("Error pool closed")
			}
			if time.Since(res.GetLastused()) > p.maxLifetime {
				p.factory.Destroy(res)
				p.decSize()
				fmt.Printf("resource %d became stale. creating new resource", res.GetID())
				continue
			}
			return res, nil
		default:
		}
		p.mu.Lock()
		if p.curSize < p.maxSize {
			p.mu.Unlock()
			res, err := p.NewResource()
			if err != nil {
				return zero, fmt.Errorf("%s", err)
			}
			p.incSize(res)
			return res, nil
		}
		p.mu.Unlock()
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case res, ok := <-p.resources:
			if !ok {
				return zero, fmt.Errorf("Error pool closed")
			}
			if time.Since(res.GetLastused()) > p.maxLifetime {
				p.factory.Destroy(res)
				p.decSize()
				fmt.Printf("resource %d became stale. creating new resource", res.GetID())
				continue
			}
			return res, nil
		}
	}
	// return zero, fmt.Errorf("error getting the resource")
}

func (p *Pool[T]) Put(res T) string {
	if res.IsNil() {
		p.mu.Lock()
		p.factory.Destroy(res)
		p.curSize--
		p.mu.Unlock()
		return "nil resource received!!!"
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		p.factory.Destroy(res)
		return fmt.Sprintf("pool closed while returning resource: %d", res.GetID())
	}
	p.mu.Unlock()
	res.SetLastused(time.Now())
	select {
	case p.resources <- res:
		return fmt.Sprintf("successfully returned resource: %d", res.GetID())
	default:
		return fmt.Sprintf("discarded resource: %d", res.GetID())
	}
}

func (p *Pool[T]) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	close(p.resources)
	for res := range p.resources {
		p.factory.Destroy(res)
	}
	p.reaperChan <- struct{}{}
}

func (p *Pool[T]) Len() int {
	return len(p.resources)
}

func (p *Pool[T]) Reaper(reaperChan <-chan struct{}) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		fmt.Printf("pool closed while reaping idle resources\n")
	}
	p.mu.Unlock()
	ticker := time.NewTicker(p.reaperInterval)
	defer ticker.Stop()
	for {
		select {
		case <-reaperChan:
			fmt.Println("closing reaper")
			return
		case <-ticker.C:
			var idleResources []T
		drainLoop:
			for {
				select {
				case res := <-p.resources:
					idleResources = append(idleResources, res)
				default:
					break drainLoop
				}
			}
			for _, res := range idleResources {
				if time.Since(res.GetLastused()) > p.idleTimeout {
					p.mu.Lock()
					p.factory.Destroy(res)
					p.curSize--
					p.mu.Unlock()
				} else {
					p.resources <- res
				}
			}
		}
	}
}

func (p *Pool[T]) Do(ctx context.Context, fn func(conn T) error) error {
	conn, err := p.Get(ctx)
	if err != nil {
		return err
	}
	defer p.Put(conn)
	if err := fn(conn); err != nil {
		return err
	}
	return nil
}

func main() {
	dbFactory := &DBFactory{}
	pool, err := New[*DBConnection](dbFactory, 2, 3*time.Second, 5, 10*time.Second, 5*time.Second)
	if err != nil {
		fmt.Println(err)
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("trying to get resource for Goroutine %d:\n", i)
			res, err := pool.Get(ctx)
			if err != nil {
				fmt.Printf("cannot create the resource for go routine %d, %s\n", i, err)
				return
			}
			fmt.Printf("Goroutine %d: Got resource %d\n", i, res.GetID())
			time.Sleep(2 * time.Second)
			pool.Put(res)
			fmt.Printf("Goroutine %d: Put resource %d back\n", i, res.GetID())
		}(i)
	}
	wg.Wait()
}
