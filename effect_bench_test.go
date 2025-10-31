package signals

import (
	"testing"
)

// BenchmarkEffect_Create measures the overhead of creating an effect.
// This includes dependency tracking and immediate execution.
func BenchmarkEffect_Create(b *testing.B) {
	count := New(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eff := Effect(
			func() {
				_ = count.Get()
			},
			count.AsReadonly(),
		)
		eff.Stop()
	}
}

// BenchmarkEffect_CreateMultipleDeps measures creation with multiple dependencies.
func BenchmarkEffect_CreateMultipleDeps(b *testing.B) {
	s1 := New(0)
	s2 := New("test")
	s3 := New(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eff := Effect(
			func() {
				_ = s1.Get()
				_ = s2.Get()
				_ = s3.Get()
			},
			s1.AsReadonly(),
			s2.AsReadonly(),
			s3.AsReadonly(),
		)
		eff.Stop()
	}
}

// BenchmarkEffect_Execute measures the time to execute an effect.
// This simulates the overhead of dependency changes triggering effects.
func BenchmarkEffect_Execute(b *testing.B) {
	count := New(0)

	eff := Effect(
		func() {
			_ = count.Get()
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count.Set(i)
	}
}

// BenchmarkEffect_ExecuteWithComputation measures effect execution with non-trivial work.
func BenchmarkEffect_ExecuteWithComputation(b *testing.B) {
	count := New(0)

	var result int
	eff := Effect(
		func() {
			// Simulate some computation
			val := count.Get()
			result = val * val
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count.Set(i)
	}
	_ = result
}

// BenchmarkEffect_Stop measures the overhead of stopping an effect.
func BenchmarkEffect_Stop(b *testing.B) {
	effects := make([]EffectRef, b.N)
	count := New(0)

	// Pre-create effects
	for i := 0; i < b.N; i++ {
		effects[i] = Effect(
			func() {
				_ = count.Get()
			},
			count.AsReadonly(),
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		effects[i].Stop()
	}
}

// BenchmarkEffect_WithCleanup measures effect creation with cleanup overhead.
func BenchmarkEffect_WithCleanup(b *testing.B) {
	count := New(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eff := EffectWithCleanup(
			func() func() {
				_ = count.Get()
				return func() {
					// Cleanup
				}
			},
			count.AsReadonly(),
		)
		eff.Stop()
	}
}

// BenchmarkEffect_CleanupExecution measures cleanup function execution overhead.
func BenchmarkEffect_CleanupExecution(b *testing.B) {
	count := New(0)

	eff := EffectWithCleanup(
		func() func() {
			_ = count.Get()
			return func() {
				// Cleanup on each run
			}
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count.Set(i)
	}
}

// BenchmarkEffect_ManyEffectsOneSignal measures memory overhead with many effects.
func BenchmarkEffect_ManyEffectsOneSignal(b *testing.B) {
	count := New(0)
	effects := make([]EffectRef, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create 100 effects on same signal
		for j := 0; j < 100; j++ {
			effects[j] = Effect(
				func() {
					_ = count.Get()
				},
				count.AsReadonly(),
			)
		}

		// Trigger them all
		count.Set(i)

		// Clean up
		for j := 0; j < 100; j++ {
			effects[j].Stop()
		}
	}
}

// BenchmarkEffect_ChainedComputed measures effect with computed signal dependency.
func BenchmarkEffect_ChainedComputed(b *testing.B) {
	base := New(0)
	computed := Computed(
		func() int {
			return base.Get() * 2
		},
		base.AsReadonly(),
	)

	var result int
	eff := Effect(
		func() {
			result = computed.Get()
		},
		computed,
	)
	defer eff.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.Set(i)
	}
	_ = result
}

// BenchmarkEffect_ConcurrentStops measures concurrent Stop() performance.
func BenchmarkEffect_ConcurrentStops(b *testing.B) {
	count := New(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eff := Effect(
			func() {
				_ = count.Get()
			},
			count.AsReadonly(),
		)

		// Concurrent stops (simulates race conditions)
		go eff.Stop()
		go eff.Stop()
		go eff.Stop()

		eff.Stop() // Ensure cleanup
	}
}

// BenchmarkEffect_NoCleanup measures effects without cleanup overhead.
func BenchmarkEffect_NoCleanup(b *testing.B) {
	count := New(0)

	eff := Effect(
		func() {
			_ = count.Get()
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count.Set(i)
	}
}

// BenchmarkEffect_WithCleanupExecution measures effects with actual cleanup work.
func BenchmarkEffect_WithCleanupExecution(b *testing.B) {
	count := New(0)

	cleanupCounter := 0
	eff := EffectWithCleanup(
		func() func() {
			_ = count.Get()
			return func() {
				cleanupCounter++
			}
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count.Set(i)
	}
}
