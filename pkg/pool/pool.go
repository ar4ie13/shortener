// Package pool implements Pool struct that allows to reuse heavy objects.
package pool

import "sync"

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	mu    sync.Mutex
	items []T
	newFn func() T
}

func New[T Resetter](factory func() T) *Pool[T] {
	if factory == nil {
		panic("pool.New: factory cannot be nil")
	}
	return &Pool[T]{
		items: make([]T, 0, 16),
		newFn: factory,
	}
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.items) > 0 {
		last := len(p.items) - 1
		item := p.items[last]
		p.items = p.items[:last]
		return item
	}

	return p.newFn()
}

func (p *Pool[T]) Put(item T) {
	if any(item) == nil {
		return
	}
	item.Reset()

	p.mu.Lock()
	p.items = append(p.items, item)
	p.mu.Unlock()
}
