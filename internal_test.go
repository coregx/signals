package signals

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestTrackDependency_NonCommonTypes verifies reflection fallback for uncommon signal types.
// The type switch in trackDependencyHelper covers int, string, bool, float64, int64.
// These tests exercise the reflection-based subscribeAnyType fallback.
func TestTrackDependency_NonCommonTypes(t *testing.T) {
	t.Run("float32 signal", func(t *testing.T) {
		sig := New(float32(3.14))
		var called int32

		comp := Computed(
			func() string {
				if sig.Get() > 3.0 {
					return "big"
				}
				return "small"
			},
			sig.AsReadonly(), // ReadonlySignal[float32] — not in type switch
		)

		comp.SubscribeForever(func(v string) {
			atomic.AddInt32(&called, 1)
		})

		sig.Set(float32(1.0))
		time.Sleep(10 * time.Millisecond)

		if got := atomic.LoadInt32(&called); got != 1 {
			t.Errorf("float32 dep: subscriber called %d times, want 1", got)
		}
		if got := comp.Get(); got != "small" {
			t.Errorf("float32 dep: comp.Get() = %q, want %q", got, "small")
		}
	})

	t.Run("byte signal", func(t *testing.T) {
		sig := New(byte(0))

		comp := Computed(
			func() int {
				return int(sig.Get()) * 2
			},
			sig.AsReadonly(), // ReadonlySignal[byte] — not in type switch
		)

		if got := comp.Get(); got != 0 {
			t.Errorf("Initial byte comp.Get() = %d, want 0", got)
		}

		sig.Set(byte(5))
		time.Sleep(10 * time.Millisecond)

		if got := comp.Get(); got != 10 {
			t.Errorf("After Set(5), byte comp.Get() = %d, want 10", got)
		}
	})

	t.Run("struct signal", func(t *testing.T) {
		type Point struct {
			X, Y int
		}

		sig := New(Point{1, 2})

		comp := Computed(
			func() int {
				p := sig.Get()
				return p.X + p.Y
			},
			sig.AsReadonly(), // ReadonlySignal[Point] — not in type switch
		)

		if got := comp.Get(); got != 3 {
			t.Errorf("Initial struct comp.Get() = %d, want 3", got)
		}

		sig.Set(Point{10, 20})
		time.Sleep(10 * time.Millisecond)

		if got := comp.Get(); got != 30 {
			t.Errorf("After Set({10,20}), struct comp.Get() = %d, want 30", got)
		}
	})
}

// TestSubscribeAnyType_InvalidDependency verifies that non-signal inputs return no-op unsubscribes.
func TestSubscribeAnyType_InvalidDependency(t *testing.T) {
	t.Run("nil dependency", func(t *testing.T) {
		var called int32
		unsub := trackDependencyHelper(nil, func() {
			atomic.AddInt32(&called, 1)
		})

		// Should return a no-op unsubscribe
		unsub()

		if got := atomic.LoadInt32(&called); got != 0 {
			t.Errorf("nil dep: onChange called %d times, want 0", got)
		}
	})

	t.Run("int dependency", func(t *testing.T) {
		var called int32
		unsub := trackDependencyHelper(42, func() {
			atomic.AddInt32(&called, 1)
		})

		unsub()

		if got := atomic.LoadInt32(&called); got != 0 {
			t.Errorf("int dep: onChange called %d times, want 0", got)
		}
	})

	t.Run("string dependency", func(t *testing.T) {
		var called int32
		unsub := trackDependencyHelper("not a signal", func() {
			atomic.AddInt32(&called, 1)
		})

		unsub()

		if got := atomic.LoadInt32(&called); got != 0 {
			t.Errorf("string dep: onChange called %d times, want 0", got)
		}
	})
}

// TestTrackDependency_AnyTypeSignal verifies that Signal[any] matches the subscriber interface directly.
// Signal[any].SubscribeForever has signature func(func(any)) Unsubscribe, which matches
// the subscriber interface in trackDependencyHelper.
func TestTrackDependency_AnyTypeSignal(t *testing.T) {
	sig := New[any]("hello")

	var called int32

	comp := Computed(
		func() string {
			v := sig.Get()
			if s, ok := v.(string); ok {
				return s
			}
			return "unknown"
		},
		sig.AsReadonly(), // ReadonlySignal[any] — matches subscriber interface directly
	)

	comp.SubscribeForever(func(v string) {
		atomic.AddInt32(&called, 1)
	})

	if got := comp.Get(); got != "hello" {
		t.Errorf("Initial comp.Get() = %q, want %q", got, "hello")
	}

	sig.Set("world")
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("Signal[any] dep: subscriber called %d times, want 1", got)
	}
	if got := comp.Get(); got != "world" {
		t.Errorf("After Set, comp.Get() = %q, want %q", got, "world")
	}
}
