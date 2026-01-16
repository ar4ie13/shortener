// Package pool implements Pool struct that allows to reuse heavy objects.
package pool

import "sync"

const (
	capacity = 16 // pre-allocated capacity for Pool.items
)

// Resetter defines the interface for objects that can be reset to their initial state.
// This is required for objects in the pool to be safely reused.
type Resetter interface {
	Reset()
}

// Pool is a generic thread-safe pool for managing reusable objects of type T.
// T must implement the Resetter interface to ensure objects can be properly
// cleaned before being returned to the pool.
type Pool[T Resetter] struct {
	mu    sync.Mutex // Protects concurrent access to the items slice
	items []T        // Slice holding available pooled objects
	newFn func() T   // Factory function to create new objects when pool is empty
}

// New creates and returns a new Pool instance.
// The factory function is required and will be called to create new objects
// when no reusable objects are available in the pool.
// Panics if factory is nil.
func New[T Resetter](factory func() T) *Pool[T] {
	if factory == nil {
		panic("pool.New: factory cannot be nil") // Ensure valid factory function
	}
	return &Pool[T]{
		items: make([]T, 0, capacity),
		newFn: factory,
	}
}

// Get returns an object from the pool. If the pool is empty, a new object
// is created using the factory function. This method is thread-safe.
func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if there are available objects in the pool
	if len(p.items) > 0 {
		// Retrieve the last object (O(1) operation)
		last := len(p.items) - 1
		item := p.items[last]
		// Remove the object from the pool
		p.items = p.items[:last]
		return item
	}

	// Pool is empty, create a new object
	return p.newFn()
}

// Put returns an object to the pool for reuse. The object is reset before
// being added back to ensure clean state. Nil objects are safely ignored.
// This method is thread-safe.
func (p *Pool[T]) Put(item T) {
	// Check for nil to prevent panics (handles interface with nil value)
	if any(item) == nil {
		return // Ignore nil objects
	}

	// Reset the object to its initial state before pooling
	item.Reset()

	p.mu.Lock()
	// Add the object back to the pool for future reuse
	p.items = append(p.items, item)
	p.mu.Unlock()
}
