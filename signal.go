package signals

import (
	"context"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// signal is the internal implementation of Signal[T].
// It uses map-based subscriber storage for O(1) unsubscribe operations.
type signal[T any] struct {
	// value is the current value of the signal
	value T

	// equal is an optional custom equality function
	equal EqualFunc[T]

	// subscribers maps unique IDs to callback functions
	// Using map instead of slice provides O(1) delete without index corruption
	subscribers map[uint64]func(T)

	// nextID is the incrementing unique ID for subscribers
	nextID uint64

	// mu protects value, subscribers, and nextID
	mu sync.RWMutex

	// onPanic is an optional custom panic handler
	onPanic func(any, []byte)

	// metrics for observability (lock-free counters)
	reads  atomic.Int64
	writes atomic.Int64
}

// New creates a new writable signal with the given initial value.
//
// The signal uses default behavior:
//   - No equality checks (always notifies on Set)
//   - Default panic handling (log and continue)
//
// Example:
//
//	count := signals.New(0)
//	count.Set(5)
//	fmt.Println(count.Get())  // 5
func New[T any](initial T) Signal[T] {
	return NewWithOptions(initial, Options[T]{})
}

// NewWithOptions creates a new writable signal with custom options.
//
// Use this when you need:
//   - Custom equality checks (opts.Equal)
//   - Custom panic handling (opts.OnPanic)
//
// Example:
//
//	// Compare slices by content, not by pointer
//	data := signals.NewWithOptions([]int{1, 2, 3}, signals.Options[[]int]{
//	    Equal: func(a, b []int) bool {
//	        return slices.Equal(a, b)
//	    },
//	})
func NewWithOptions[T any](initial T, opts Options[T]) Signal[T] {
	return &signal[T]{
		value:       initial,
		equal:       opts.Equal,
		subscribers: make(map[uint64]func(T)),
		onPanic:     opts.OnPanic,
	}
}

// Get returns the current value of the signal.
// This operation is thread-safe and uses a read lock (RLock).
func (s *signal[T]) Get() T {
	s.reads.Add(1) // Lock-free metric

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set replaces the signal's value with a new value.
//
// If a custom Equal function is provided, Set will check equality
// and only notify subscribers if the value has changed.
//
// All subscriber callbacks are executed with panic recovery.
// One panicking subscriber does not affect others.
func (s *signal[T]) Set(newValue T) {
	// Fast path: check equality without write lock
	if s.equal != nil {
		s.mu.RLock()
		if s.equal(s.value, newValue) {
			s.mu.RUnlock()
			return // Value hasn't changed, don't notify
		}
		s.mu.RUnlock()
	}

	s.writes.Add(1) // Lock-free metric

	// Update value and copy subscribers inside lock
	s.mu.Lock()
	s.value = newValue

	// Copy subscribers to slice for safe iteration outside lock
	callbacks := make([]func(T), 0, len(s.subscribers))
	for _, fn := range s.subscribers {
		callbacks = append(callbacks, fn)
	}
	s.mu.Unlock()

	// Notify subscribers outside lock (prevents deadlock)
	s.notifySubscribers(callbacks, newValue)
}

// Update transforms the signal's value using the provided function.
//
// The transform function receives the current value and returns the new value.
// The entire read-transform-write operation is atomic.
//
// Example:
//
//	count.Update(func(v int) int { return v + 1 })
func (s *signal[T]) Update(fn func(T) T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Atomic read-transform-write
	oldValue := s.value
	newValue := fn(oldValue)

	// Check equality if custom function provided
	if s.equal != nil && s.equal(oldValue, newValue) {
		return
	}

	// Update value
	s.value = newValue

	// Copy subscribers before unlock
	callbacks := make([]func(T), 0, len(s.subscribers))
	for _, fn := range s.subscribers {
		callbacks = append(callbacks, fn)
	}
	s.mu.Unlock()

	// Notify outside lock
	s.notifySubscribers(callbacks, newValue)

	// Re-acquire lock for defer
	s.mu.Lock()
}

// Subscribe registers a callback to be notified when the signal's value changes.
//
// The subscription is automatically canceled when the context is done.
// Returns an Unsubscribe function for manual cleanup.
//
// IMPORTANT: Unsubscribe MUST be called to prevent memory leaks, even when
// using context cancellation.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	unsub := sig.Subscribe(ctx, func(v int) {
//	    fmt.Println("Value:", v)
//	})
//	defer unsub()  // Cleanup (before context timeout)
func (s *signal[T]) Subscribe(ctx context.Context, fn func(T)) Unsubscribe {
	// Add subscriber with unique ID
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.subscribers[id] = fn
	s.mu.Unlock()

	// Channel to signal cleanup completion
	done := make(chan struct{})

	// Goroutine for context-based cleanup
	go func() {
		select {
		case <-ctx.Done():
			// Context canceled - auto cleanup
			s.mu.Lock()
			delete(s.subscribers, id)
			s.mu.Unlock()
			close(done)
		case <-done:
			// Manual unsubscribe happened
		}
	}()

	// Return manual unsubscribe function
	return func() {
		s.mu.Lock()
		delete(s.subscribers, id)
		s.mu.Unlock()

		// Signal goroutine to stop
		select {
		case <-done:
			// Already closed by context
		default:
			close(done)
		}
	}
}

// SubscribeForever registers a callback that will never be automatically canceled.
// Equivalent to Subscribe(context.Background(), fn).
//
// IMPORTANT: You MUST call the returned Unsubscribe function to prevent memory leaks.
//
// Example:
//
//	unsub := sig.SubscribeForever(func(v int) {
//	    fmt.Println(v)
//	})
//	defer unsub()  // REQUIRED
func (s *signal[T]) SubscribeForever(fn func(T)) Unsubscribe {
	return s.Subscribe(context.Background(), fn)
}

// AsReadonly returns a read-only view of this signal.
// Use for encapsulation - keep Signal private, expose ReadonlySignal.
func (s *signal[T]) AsReadonly() ReadonlySignal[T] {
	return &readonlySignal[T]{source: s}
}

// notifySubscribers calls all subscriber callbacks with panic recovery.
// One panicking subscriber does not affect others.
func (s *signal[T]) notifySubscribers(callbacks []func(T), value T) {
	for _, fn := range callbacks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if s.onPanic != nil {
						s.onPanic(r, debug.Stack())
					} else {
						// Default: log and continue
						log.Printf("signals: panic in subscriber: %v\n%s", r, debug.Stack())
					}
				}
			}()
			fn(value)
		}()
	}
}
