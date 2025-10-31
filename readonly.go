package signals

import "context"

// readonlySignal is a read-only wrapper around a Signal.
// It implements ReadonlySignal by delegating to the source Signal,
// but does not expose Set/Update methods.
type readonlySignal[T any] struct {
	source Signal[T]
}

// Get returns the current value from the source signal.
func (r *readonlySignal[T]) Get() T {
	return r.source.Get()
}

// Subscribe registers a callback with the source signal.
func (r *readonlySignal[T]) Subscribe(ctx context.Context, fn func(T)) Unsubscribe {
	return r.source.Subscribe(ctx, fn)
}

// SubscribeForever registers a callback with the source signal that never auto-cancels.
func (r *readonlySignal[T]) SubscribeForever(fn func(T)) Unsubscribe {
	return r.source.SubscribeForever(fn)
}
