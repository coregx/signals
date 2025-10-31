package signals

// EqualFunc is a function that compares two values for equality.
// It returns true if the values are considered equal, false otherwise.
//
// Use custom equality functions when you need:
//   - Value-based comparison for complex types
//   - Comparison by specific fields (e.g., ID only)
//   - Custom business logic for equality
//
// Example:
//
//	type User struct {
//	    ID   int
//	    Name string
//	}
//
//	// Compare users by ID only
//	userSignal := signals.NewWithOptions(&User{ID: 1, Name: "Alice"}, signals.Options[*User]{
//	    Equal: func(a, b *User) bool {
//	        if a == nil || b == nil {
//	            return a == b
//	        }
//	        return a.ID == b.ID
//	    },
//	})
type EqualFunc[T any] func(a, b T) bool

// Options configures the behavior of a Signal.
type Options[T any] struct {
	// Equal is an optional custom equality function.
	// If nil, signals will not perform equality checks and always notify on Set().
	//
	// Angular Signals use Object.is() by default (referential equality).
	// For Go, we allow optional equality checks since not all types are comparable.
	Equal EqualFunc[T]

	// OnPanic is an optional custom panic handler for subscriber callbacks.
	// If nil, panics are logged to stderr and execution continues.
	//
	// This handler is called when a subscriber panics, allowing custom logging,
	// metrics, or error recovery strategies.
	//
	// Example:
	//   OnPanic: func(err any, stack []byte) {
	//       log.Printf("Subscriber panic: %v\n%s", err, stack)
	//       metrics.IncrementPanicCounter()
	//   }
	OnPanic func(err any, stack []byte)
}
