package signals

import "testing"

// BenchmarkComputed_Get_Clean measures performance of cached reads
func BenchmarkComputed_Get_Clean(b *testing.B) {
	count := New(42)

	comp := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	// Prime the cache
	_ = comp.Get()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.Get() // Should be lock-free!
	}
}

// BenchmarkComputed_Get_Dirty measures performance when recomputation needed
func BenchmarkComputed_Get_Dirty(b *testing.B) {
	count := New(0)

	comp := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count.Set(i)   // Mark dirty
		_ = comp.Get() // Recompute
	}
}

// BenchmarkComputed_MultipleDeps measures performance with multiple dependencies
func BenchmarkComputed_MultipleDeps(b *testing.B) {
	a := New(1)
	b1 := New(2)
	c := New(3)

	comp := Computed(
		func() int {
			return a.Get() + b1.Get() + c.Get()
		},
		a.AsReadonly(), b1.AsReadonly(), c.AsReadonly(),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.Get()
	}
}

// BenchmarkComputed_Chained measures performance of chained computed signals
func BenchmarkComputed_Chained(b *testing.B) {
	count := New(5)

	doubled := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	quadrupled := Computed(
		func() int { return doubled.Get() * 2 },
		doubled,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = quadrupled.Get()
	}
}

// BenchmarkComputed_Subscribe measures subscription performance
func BenchmarkComputed_Subscribe(b *testing.B) {
	count := New(0)

	comp := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsub := comp.SubscribeForever(func(v int) {})
		unsub()
	}
}

// BenchmarkComputed_ParallelGet_Clean measures concurrent cached reads
func BenchmarkComputed_ParallelGet_Clean(b *testing.B) {
	count := New(42)

	comp := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	// Prime cache
	_ = comp.Get()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = comp.Get()
		}
	})
}

// BenchmarkComputed_ComplexComputation measures expensive computation
func BenchmarkComputed_ComplexComputation(b *testing.B) {
	count := New(100)

	comp := Computed(
		func() int {
			// Simulate expensive computation
			result := 0
			n := count.Get()
			for i := 0; i < n; i++ {
				result += i
			}
			return result
		},
		count.AsReadonly(),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.Get() // Should be cached!
	}
}
