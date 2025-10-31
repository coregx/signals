// Package signals provides a reactive state management library for Go,
// inspired by Angular Signals.
//
// Signals are reactive containers that notify subscribers when their values change.
// This package provides type-safe, thread-safe, and memory-safe primitives for
// building reactive applications.
//
// # Core Types
//
// Signal[T] - A writable reactive value that notifies subscribers on changes.
//
// ReadonlySignal[T] - A read-only view of a signal for encapsulation.
//
// Computed[T] - A derived reactive value that auto-updates when dependencies change.
//
// Effect - Side effects that run when dependencies change.
//
// # Example Usage
//
//	// Create a writable signal
//	count := signals.New(0)
//
//	// Subscribe to changes
//	unsub := count.SubscribeForever(func(v int) {
//	    fmt.Printf("Count changed: %d\n", v)
//	})
//	defer unsub()
//
//	// Update the signal
//	count.Set(5)                              // Prints: Count changed: 5
//	count.Update(func(v int) int { return v + 1 })  // Prints: Count changed: 6
//
// # Thread Safety
//
// All operations are thread-safe and protected by sync.RWMutex.
// The library is designed for concurrent use and includes comprehensive
// race condition testing.
//
// # Memory Safety
//
// All subscriptions return cleanup functions (Unsubscribe) that must be called
// to prevent memory leaks. Use context.Context for automatic cleanup:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	sig.Subscribe(ctx, func(v int) {
//	    fmt.Println(v)
//	})
//	// Automatically unsubscribes after 5 seconds
//
// # Encapsulation Pattern
//
// Use AsReadonly() to expose signals while keeping mutations controlled:
//
//	type CounterService struct {
//	    count Signal[int]  // private
//	}
//
//	func (s *CounterService) Count() ReadonlySignal[int] {
//	    return s.count.AsReadonly()  // expose read-only
//	}
//
//	func (s *CounterService) Increment() {
//	    s.count.Update(func(n int) int { return n + 1 })
//	}
//
// # Performance
//
// Signal operations are highly optimized:
//   - Get(): < 15ns/op (read-locked)
//   - Set(): < 200ns/op (with notification)
//   - Subscribe/Unsubscribe: O(1) using map-based storage
//
// # Design Principles
//
// 1. Type Safety - 100% generic, no interface{} or type assertions
// 2. Thread Safety - All operations protected with proper locking
// 3. Memory Safety - Explicit cleanup with Unsubscribe functions
// 4. Panic Safety - All callbacks execute with panic recovery
// 5. Context Awareness - Standard Go context.Context integration
//
// For detailed documentation, see: https://github.com/coregx/signals
package signals
