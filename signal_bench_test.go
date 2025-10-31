package signals

import (
	"context"
	"testing"
)

// BenchmarkSignal_Get measures read performance
func BenchmarkSignal_Get(b *testing.B) {
	sig := New(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sig.Get()
	}
}

// BenchmarkSignal_Set measures write performance (no subscribers)
func BenchmarkSignal_Set(b *testing.B) {
	sig := New(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig.Set(i)
	}
}

// BenchmarkSignal_SetWithSubscribers measures write performance with subscribers
func BenchmarkSignal_SetWithSubscribers(b *testing.B) {
	sig := New(0)

	// Add 10 subscribers
	for i := 0; i < 10; i++ {
		sig.SubscribeForever(func(v int) {
			// Minimal work
			_ = v
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig.Set(i)
	}
}

// BenchmarkSignal_Update measures Update performance
func BenchmarkSignal_Update(b *testing.B) {
	sig := New(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig.Update(func(v int) int { return v + 1 })
	}
}

// BenchmarkSignal_Subscribe measures subscription performance
func BenchmarkSignal_Subscribe(b *testing.B) {
	sig := New(0)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsub := sig.Subscribe(ctx, func(v int) {})
		unsub()
	}
}

// BenchmarkSignal_Unsubscribe measures unsubscription performance
func BenchmarkSignal_Unsubscribe(b *testing.B) {
	sig := New(0)

	// Pre-create unsubscribe functions
	unsubs := make([]Unsubscribe, b.N)
	for i := 0; i < b.N; i++ {
		unsubs[i] = sig.SubscribeForever(func(v int) {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsubs[i]()
	}
}

// BenchmarkSignal_EqualCheck measures Set performance with equality checks
func BenchmarkSignal_EqualCheck(b *testing.B) {
	sig := NewWithOptions(42, Options[int]{
		Equal: func(a, b int) bool {
			return a == b
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig.Set(42) // Same value - should not notify
	}
}

// BenchmarkSignal_ParallelGet measures concurrent read performance
func BenchmarkSignal_ParallelGet(b *testing.B) {
	sig := New(42)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = sig.Get()
		}
	})
}

// BenchmarkSignal_ParallelSet measures concurrent write performance
func BenchmarkSignal_ParallelSet(b *testing.B) {
	sig := New(0)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			sig.Set(i)
			i++
		}
	})
}
