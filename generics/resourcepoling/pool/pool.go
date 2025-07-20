package pool

// type resource func(int) int

// func display(int) int {
// 	return 5
// }
// func main() {
// 	var x resource = display
// 	x(5)
// }
import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrPoolClosed is returned when an operation is attempted on a closed pool.
	ErrPoolClosed = errors.New("pool is closed")

	// ErrInvalidConfig is returned when the pool is configured with invalid parameters.
	ErrInvalidConfig = errors.New("invalid pool configuration")
)

type resource interface {
	Close() error
}

type Factory[T resource] func() (T, error)

type Pool[T resource] struct {
	mu        sync.Mutex
	resources chan T
	factory   Factory[T]
	closed    bool
}

func New[T resource](factory Factory[T], intial, max int) (*Pool[T], error) {
	if intial < 0 || max <= 0 || intial > max {
		return nil, ErrInvalidConfig
	}
	p := Pool[T]{
		resources: make(chan T, max),
		factory:   factory,
	}
	for i := 0; i < intial; i++ {
		res, err := factory()
		if err != nil {
			close(p.resources)
			for r := range p.resources {
				r.Close()
			}
			return nil, err
		}
		p.resources <- res
	}
	return &p, nil
}

func (p *Pool[T]) Get(ctx context.Context) (T, error) {
	select {
	case res, ok := <-p.resources:
		if !ok {
			var zero T
			return zero, ErrPoolClosed
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.closed {
			res.Close()
			var zero T
			return zero, ErrPoolClosed
		}
		if len(p.resources) < cap(p.resources) {
			go func() {
				res, err := p.factory()
				if err != nil {
					// Log the error, can't do much more here.
					return
				}
				// We need to check if the pool was closed while we were creating the resource.
				p.mu.Lock()
				defer p.mu.Unlock()
				if p.closed {
					res.Close()
					return
				}
				p.resources <- res
			}()
		}
		return res, nil
	case <-ctx.Done():
		// The context was cancelled (e.g., timeout).
		var zero T
		return zero, ctx.Err()
	}
}

func (p *Pool[T]) Len() int {
	return len(p.resources)
}

func (p *Pool[T]) Put(res T) {
	if any(res) == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		res.Close()
		return
	}
	select {
	case p.resources <- res:

	default:
		res.Close()
	}
}

func (p *Pool[T]) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.resources)
	p.mu.Unlock()
	for res := range p.resources {
		res.Close()
	}
}
