package signals

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestComputed_Basic verifies basic computed signal functionality
func TestComputed_Basic(t *testing.T) {
	count := New(5)

	doubled := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	if got := doubled.Get(); got != 10 {
		t.Errorf("Computed() = %d, want 10", got)
	}

	// Change dependency
	count.Set(10)

	if got := doubled.Get(); got != 20 {
		t.Errorf("After Set(10), Computed() = %d, want 20", got)
	}
}

// TestComputed_MultipleDependencies verifies computed with multiple deps
func TestComputed_MultipleDependencies(t *testing.T) {
	firstName := New("John")
	lastName := New("Doe")

	fullName := Computed(
		func() string {
			return firstName.Get() + " " + lastName.Get()
		},
		firstName.AsReadonly(), lastName.AsReadonly(),
	)

	if got := fullName.Get(); got != "John Doe" {
		t.Errorf("fullName.Get() = %q, want %q", got, "John Doe")
	}

	firstName.Set("Jane")
	if got := fullName.Get(); got != "Jane Doe" {
		t.Errorf("After firstName change, fullName.Get() = %q, want %q", got, "Jane Doe")
	}

	lastName.Set("Smith")
	if got := fullName.Get(); got != "Jane Smith" {
		t.Errorf("After lastName change, fullName.Get() = %q, want %q", got, "Jane Smith")
	}
}

// TestComputed_Memoization verifies lazy evaluation and caching
func TestComputed_Memoization(t *testing.T) {
	count := New(5)
	var computeCount int32

	doubled := Computed(
		func() int {
			atomic.AddInt32(&computeCount, 1)
			return count.Get() * 2
		},
		count.AsReadonly(),
	)

	// First Get - should compute
	doubled.Get()
	if got := atomic.LoadInt32(&computeCount); got != 1 {
		t.Errorf("First Get: computed %d times, want 1", got)
	}

	// Second Get - should use cache (no recomputation)
	doubled.Get()
	doubled.Get()
	if got := atomic.LoadInt32(&computeCount); got != 1 {
		t.Errorf("After cache hits: computed %d times, want 1 (memoized)", got)
	}

	// Change dependency - mark dirty
	count.Set(10)

	// Next Get - should recompute
	doubled.Get()
	if got := atomic.LoadInt32(&computeCount); got != 2 {
		t.Errorf("After dependency change: computed %d times, want 2", got)
	}
}

// TestComputed_Subscribe verifies subscription to computed signals
func TestComputed_Subscribe(t *testing.T) {
	count := New(0)

	doubled := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	var calls []int
	var mu sync.Mutex

	unsub := doubled.SubscribeForever(func(v int) {
		mu.Lock()
		calls = append(calls, v)
		mu.Unlock()
	})
	defer unsub()

	count.Set(5)
	time.Sleep(10 * time.Millisecond)

	count.Set(10)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(calls) != 2 {
		t.Errorf("Expected 2 calls, got %d", len(calls))
	}

	expected := []int{10, 20}
	for i, want := range expected {
		if i >= len(calls) || calls[i] != want {
			t.Errorf("Call %d: got %v, want %d", i, calls, want)
			break
		}
	}
}

// TestComputed_Unsubscribe verifies unsubscribe stops notifications
func TestComputed_Unsubscribe(t *testing.T) {
	count := New(0)

	doubled := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	var called int32

	unsub := doubled.SubscribeForever(func(v int) {
		atomic.AddInt32(&called, 1)
	})

	count.Set(5)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After Set(5), called = %d, want 1", got)
	}

	// Unsubscribe
	unsub()

	count.Set(10)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After unsubscribe and Set(10), called = %d, want 1 (no new calls)", got)
	}
}

// TestComputed_ContextCancel verifies context-based auto-unsubscribe
func TestComputed_ContextCancel(t *testing.T) {
	count := New(0)

	doubled := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	var called int32

	doubled.Subscribe(ctx, func(v int) {
		atomic.AddInt32(&called, 1)
	})

	count.Set(5)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("Before cancel, called = %d, want 1", got)
	}

	// Cancel context
	cancel()
	time.Sleep(10 * time.Millisecond)

	count.Set(10)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After context cancel, called = %d, want 1 (no new calls)", got)
	}
}

// TestComputed_PanicRecovery verifies panic recovery in compute function
func TestComputed_PanicRecovery(t *testing.T) {
	count := New(0)

	var panicCount int32

	comp := ComputedWithOptions(
		func() int {
			if count.Get() == 5 {
				atomic.AddInt32(&panicCount, 1)
				panic("compute panic")
			}
			return count.Get() * 2
		},
		Options[int]{},
		count.AsReadonly(),
	)

	// First Get - no panic
	if got := comp.Get(); got != 0 {
		t.Errorf("Initial Get() = %d, want 0", got)
	}

	// Trigger panic
	count.Set(5)

	// Get should recover and return old value
	comp.Get()

	if got := atomic.LoadInt32(&panicCount); got != 1 {
		t.Errorf("Panic count = %d, want 1", got)
	}

	// Set to non-panic value
	count.Set(10)

	// Should compute normally again
	if got := comp.Get(); got != 20 {
		t.Errorf("After panic recovery, Get() = %d, want 20", got)
	}
}

// TestComputed_CustomPanicHandler verifies custom panic handling
func TestComputed_CustomPanicHandler(t *testing.T) {
	count := New(5)
	var panicHandlerCalled int32

	comp := ComputedWithOptions(
		func() int {
			panic("custom panic")
		},
		Options[int]{
			OnPanic: func(err any, stack []byte) {
				atomic.AddInt32(&panicHandlerCalled, 1)
				if err != "custom panic" {
					t.Errorf("OnPanic: got error %v, want 'custom panic'", err)
				}
			},
		},
		count.AsReadonly(),
	)

	comp.Get()

	if got := atomic.LoadInt32(&panicHandlerCalled); got != 1 {
		t.Errorf("Custom panic handler called %d times, want 1", got)
	}
}

// TestComputed_ChainedComputed verifies computed depending on computed
func TestComputed_ChainedComputed(t *testing.T) {
	count := New(5)

	doubled := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	quadrupled := Computed(
		func() int { return doubled.Get() * 2 },
		doubled,
	)

	if got := quadrupled.Get(); got != 20 {
		t.Errorf("quadrupled.Get() = %d, want 20 (5*2*2)", got)
	}

	count.Set(10)

	if got := quadrupled.Get(); got != 40 {
		t.Errorf("After Set(10), quadrupled.Get() = %d, want 40 (10*2*2)", got)
	}
}

// TestComputed_ConcurrentReads verifies safe concurrent reads
func TestComputed_ConcurrentReads(t *testing.T) {
	count := New(42)

	comp := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = comp.Get()
			}
		}()
	}

	wg.Wait()
}

// TestComputed_ConcurrentUpdates verifies safe concurrent dependency updates
func TestComputed_ConcurrentUpdates(t *testing.T) {
	count := New(0)

	comp := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	)

	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = comp.Get()
			}
		}()
	}

	// Concurrent writers (dependencies)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			count.Set(val)
		}(i)
	}

	wg.Wait()
}

// TestComputed_NoMemoryLeak verifies cleanup prevents memory leaks
func TestComputed_NoMemoryLeak(t *testing.T) {
	count := New(0)

	comp := Computed(
		func() int { return count.Get() * 2 },
		count.AsReadonly(),
	).(*computed[int])

	// Subscribe and unsubscribe 1000 times
	for i := 0; i < 1000; i++ {
		unsub := comp.SubscribeForever(func(v int) {})
		unsub()
	}

	// Check subscribers map is empty
	comp.mu.RLock()
	subCount := len(comp.subscribers)
	comp.mu.RUnlock()

	if subCount != 0 {
		t.Errorf("Memory leak: %d subscribers still registered, want 0", subCount)
	}
}

// TestComputed_RapidDependencyChanges tests rapid dep changes don't cause issues
func TestComputed_RapidDependencyChanges(t *testing.T) {
	count := New(0)
	var computeCount int32

	comp := Computed(
		func() int {
			atomic.AddInt32(&computeCount, 1)
			return count.Get() * 2
		},
		count.AsReadonly(),
	)

	// Rapid changes
	for i := 0; i < 100; i++ {
		count.Set(i)
	}

	// Final Get
	result := comp.Get()

	// Should compute at least once, but memoization should prevent excessive recomputation
	computations := atomic.LoadInt32(&computeCount)
	t.Logf("Computations: %d for 100 dependency changes", computations)

	if result != 198 { // 99 * 2
		t.Errorf("Final result = %d, want 198", result)
	}
}
