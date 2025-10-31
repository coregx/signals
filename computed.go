package signals

import (
	"context"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// computed is the internal implementation of a computed signal.
// It lazily evaluates a computation function and caches the result
// until dependencies change.
//
// Uses atomic.Bool for lock-free dirty flag checks.
type computed[T any] struct {
	// compute is the function that derives the value
	compute func() T

	// cached is the memoized result
	cached T

	// dirty indicates if cached value needs recomputation
	// Using atomic.Bool for lock-free reads
	dirty atomic.Bool

	// unsubscribes are cleanup functions for dependency subscriptions
	// We don't store dependencies themselves, only their unsubscribe functions
	unsubscribes []Unsubscribe

	// subscribers for this computed signal
	subscribers map[uint64]func(T)
	nextID      uint64

	// mu protects cached, subscribers, and nextID
	mu sync.RWMutex

	// onPanic is optional custom panic handler
	onPanic func(any, []byte)
}

// Computed creates a read-only signal that derives its value from a computation function.
//
// IMPORTANT: The compute function MUST be pure - it should only read signals and compute
// a result without side effects (no logging, no mutations, no I/O).
//
// Dependencies must be explicitly passed as additional arguments. Each dependency can be
// a ReadonlySignal of any type. When any dependency changes, this computed signal is
// marked dirty and will recompute on the next Get().
//
// The computed signal uses lazy evaluation and memoization:
//   - Only computes when accessed (Get)
//   - Caches result until marked dirty
//   - Uses atomic operations for lock-free dirty checks
//
// Example:
//
//	firstName := signals.New("John")
//	lastName := signals.New("Doe")
//
//	// Dependencies are explicit - pass signals after compute function
//	fullName := signals.Computed(
//	    func() string {
//	        return firstName.Get() + " " + lastName.Get()
//	    },
//	    firstName.AsReadonly(),
//	    lastName.AsReadonly(),
//	)
//
//	fmt.Println(fullName.Get())  // "John Doe"
//	firstName.Set("Jane")
//	// Computed is marked dirty, will recompute on next Get()
//	fmt.Println(fullName.Get())  // "Jane Doe"
//
// Dependencies can be of different types:
//
//	count := signals.New(5)
//	name := signals.New("items")
//
//	message := signals.Computed(
//	    func() string {
//	        return fmt.Sprintf("%d %s", count.Get(), name.Get())
//	    },
//	    count.AsReadonly(),  // ReadonlySignal[int]
//	    name.AsReadonly(),   // ReadonlySignal[string]
//	)
func Computed[T any](compute func() T, deps ...any) ReadonlySignal[T] {
	return ComputedWithOptions(compute, Options[T]{}, deps...)
}

// ComputedWithOptions creates a computed signal with custom options.
//
// Use this when you need custom panic handling for the compute function or subscribers.
//
// Example:
//
//	count := signals.New(5)
//	comp := signals.ComputedWithOptions(
//	    func() int { return count.Get() * 2 },
//	    signals.Options[int]{
//	        OnPanic: func(err any, stack []byte) {
//	            metrics.IncrementComputedPanic()
//	        },
//	    },
//	    count.AsReadonly(),
//	)
func ComputedWithOptions[T any](compute func() T, opts Options[T], deps ...any) ReadonlySignal[T] {
	c := &computed[T]{
		compute:     compute,
		subscribers: make(map[uint64]func(T)),
		onPanic:     opts.OnPanic,
	}

	// Mark as dirty initially (needs first computation)
	c.dirty.Store(true)

	// Track dependencies using type erasure
	for _, dep := range deps {
		c.trackDependency(dep)
	}

	return c
}

// trackDependency registers a signal as a dependency using type erasure.
// Accepts any ReadonlySignal[X] where X is any type.
//
// This is an internal method used by Computed() and ComputedWithOptions().
func (c *computed[T]) trackDependency(dep any) {
	unsub := trackDependencyHelper(dep, c.markDirty)
	c.unsubscribes = append(c.unsubscribes, unsub)
}

// Get returns the current value of the computed signal.
//
// If the value is cached and not dirty, returns immediately (lock-free).
// If dirty, recomputes the value with proper locking.
//
// Uses double-check locking pattern to minimize lock contention.
func (c *computed[T]) Get() T {
	// Fast path: not dirty (lock-free!)
	if !c.dirty.Load() {
		c.mu.RLock()
		cached := c.cached
		c.mu.RUnlock()
		return cached
	}

	// Slow path: recompute with lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check locking: another goroutine might have recomputed
	if !c.dirty.Load() {
		return c.cached
	}

	// Recompute with panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				if c.onPanic != nil {
					c.onPanic(r, debug.Stack())
				} else {
					log.Printf("signals: panic in computed function: %v\n%s", r, debug.Stack())
				}
				// Don't update cached value on panic - keep old value
			}
		}()
		c.cached = c.compute()
	}()

	c.dirty.Store(false)
	return c.cached
}

// Subscribe registers a callback to be notified when the computed value changes.
//
// The computed signal notifies subscribers when:
//   - A dependency changes AND the computed value is recomputed
//
// Note: Unlike regular signals, computed signals only notify after recomputation,
// not on every dependency change (lazy evaluation).
func (c *computed[T]) Subscribe(ctx context.Context, fn func(T)) Unsubscribe {
	// Add subscriber
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.subscribers[id] = fn
	c.mu.Unlock()

	// Channel for cleanup coordination
	done := make(chan struct{})

	// Auto-cleanup on context cancellation
	go func() {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			delete(c.subscribers, id)
			c.mu.Unlock()
			close(done)
		case <-done:
			// Manual unsubscribe
		}
	}()

	// Return manual unsubscribe
	return func() {
		c.mu.Lock()
		delete(c.subscribers, id)
		c.mu.Unlock()

		select {
		case <-done:
			// Already closed
		default:
			close(done)
		}
	}
}

// SubscribeForever registers a callback that never auto-cancels.
func (c *computed[T]) SubscribeForever(fn func(T)) Unsubscribe {
	return c.Subscribe(context.Background(), fn)
}

// markDirty marks the computed value as stale and triggers recomputation.
//
// This is called when any dependency changes.
// Always triggers recomputation and notification to ensure subscribers are notified.
func (c *computed[T]) markDirty() {
	// Mark as dirty
	c.dirty.Store(true)

	// Always recompute and notify
	// This ensures that even if the signal was already dirty (e.g., initial state),
	// subscribers still get notified when dependencies change.
	newValue := c.Get()

	// Notify subscribers
	c.notifySubscribers(newValue)
}

// notifySubscribers calls all subscriber callbacks with panic recovery.
func (c *computed[T]) notifySubscribers(value T) {
	c.mu.RLock()
	callbacks := make([]func(T), 0, len(c.subscribers))
	for _, fn := range c.subscribers {
		callbacks = append(callbacks, fn)
	}
	c.mu.RUnlock()

	// Notify outside lock with panic recovery
	for _, fn := range callbacks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if c.onPanic != nil {
						c.onPanic(r, debug.Stack())
					} else {
						log.Printf("signals: panic in computed subscriber: %v\n%s", r, debug.Stack())
					}
				}
			}()
			fn(value)
		}()
	}
}

// Cleanup stops all dependency subscriptions.
// Call this to prevent memory leaks when the computed signal is no longer needed.
//
// Note: This is not part of the ReadonlySignal interface, but provided as
// a utility method on the concrete type.
func (c *computed[T]) Cleanup() {
	for _, unsub := range c.unsubscribes {
		unsub()
	}
	c.unsubscribes = nil
}
