package signals

import "context"

// Unsubscribe is a function that removes a subscription.
// Call it to stop receiving notifications and prevent memory leaks.
//
// Example:
//
//	unsub := signal.Subscribe(ctx, func(v int) {
//	    fmt.Println(v)
//	})
//	defer unsub()  // Cleanup
type Unsubscribe func()

// Signal is a writable reactive container for a value of type T.
//
// Signals notify subscribers when their value changes (via Set or Update).
// All operations are thread-safe and can be called from multiple goroutines.
//
// Example:
//
//	count := signals.New(0)
//	count.Set(5)
//	value := count.Get()  // 5
//	count.Update(func(v int) int { return v + 1 })  // Now 6
type Signal[T any] interface {
	// Get returns the current value of the signal.
	// This operation is thread-safe and uses a read lock.
	Get() T

	// Set replaces the signal's value with a new value.
	// If a custom Equal function is provided, the signal will only notify
	// subscribers if the new value is different from the old value.
	//
	// All subscribers are notified after the value is updated.
	Set(value T)

	// Update transforms the signal's value using the provided function.
	// The function receives the current value and returns the new value.
	//
	// This operation locks the signal for the duration of the transform function,
	// so keep the function fast. After the transform, Set() is called with the
	// new value (triggering equality checks and notifications).
	//
	// Example:
	//   count.Update(func(v int) int { return v + 1 })
	Update(fn func(T) T)

	// AsReadonly returns a read-only view of this signal.
	// Use this for encapsulation - keep the Signal private, expose ReadonlySignal.
	//
	// This follows the Angular Signals pattern of controlled mutations.
	//
	// Example:
	//   type Service struct {
	//       count Signal[int]  // private
	//   }
	//
	//   func (s *Service) Count() ReadonlySignal[int] {
	//       return s.count.AsReadonly()
	//   }
	AsReadonly() ReadonlySignal[T]

	// Subscribe registers a callback to be notified when the signal's value changes.
	// The callback receives the new value.
	//
	// The subscription is automatically canceled when the context is done.
	// This allows automatic cleanup on timeout, cancellation, or deadline.
	//
	// Returns an Unsubscribe function for manual cleanup.
	//
	// Example:
	//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//   defer cancel()
	//
	//   unsub := sig.Subscribe(ctx, func(v int) {
	//       fmt.Println("Value:", v)
	//   })
	//   defer unsub()  // Manual cleanup (happens before context timeout)
	Subscribe(ctx context.Context, fn func(T)) Unsubscribe

	// SubscribeForever registers a callback that will never be automatically canceled.
	// Equivalent to Subscribe(context.Background(), fn).
	//
	// IMPORTANT: You MUST call the returned Unsubscribe function to prevent memory leaks.
	//
	// Example:
	//   unsub := sig.SubscribeForever(func(v int) {
	//       fmt.Println(v)
	//   })
	//   defer unsub()  // REQUIRED for cleanup
	SubscribeForever(fn func(T)) Unsubscribe
}

// ReadonlySignal is a read-only view of a Signal.
//
// Use this type for encapsulation - expose ReadonlySignal while keeping
// the writable Signal private. This prevents external code from modifying
// the signal's value.
//
// This pattern is inspired by Angular Signals' asReadonly() method.
//
// Example:
//
//	type CounterService struct {
//	    counter Signal[int]  // private writable signal
//	}
//
//	func (s *CounterService) Counter() ReadonlySignal[int] {
//	    return s.counter.AsReadonly()  // expose as read-only
//	}
//
//	func (s *CounterService) Increment() {
//	    s.counter.Update(func(n int) int { return n + 1 })
//	}
type ReadonlySignal[T any] interface {
	// Get returns the current value of the signal.
	Get() T

	// Subscribe registers a callback to be notified when the signal's value changes.
	Subscribe(ctx context.Context, fn func(T)) Unsubscribe

	// SubscribeForever registers a callback that will never be automatically canceled.
	SubscribeForever(fn func(T)) Unsubscribe
}
