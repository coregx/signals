package signals

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestSignal_New verifies basic signal creation and initial value
func TestSignal_New(t *testing.T) {
	sig := New(42)

	if got := sig.Get(); got != 42 {
		t.Errorf("New(42).Get() = %d, want 42", got)
	}
}

// TestSignal_Get verifies reading signal values
func TestSignal_Get(t *testing.T) {
	tests := []struct {
		name  string
		value int
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := New(tt.value)
			if got := sig.Get(); got != tt.value {
				t.Errorf("Get() = %d, want %d", got, tt.value)
			}
		})
	}
}

// TestSignal_Set verifies setting signal values
func TestSignal_Set(t *testing.T) {
	sig := New(0)

	sig.Set(10)
	if got := sig.Get(); got != 10 {
		t.Errorf("After Set(10), Get() = %d, want 10", got)
	}

	sig.Set(20)
	if got := sig.Get(); got != 20 {
		t.Errorf("After Set(20), Get() = %d, want 20", got)
	}
}

// TestSignal_Update verifies transforming signal values
func TestSignal_Update(t *testing.T) {
	sig := New(5)

	sig.Update(func(v int) int { return v * 2 })
	if got := sig.Get(); got != 10 {
		t.Errorf("After Update(*2), Get() = %d, want 10", got)
	}

	sig.Update(func(v int) int { return v + 3 })
	if got := sig.Get(); got != 13 {
		t.Errorf("After Update(+3), Get() = %d, want 13", got)
	}
}

// TestSignal_SubscribeForever verifies basic subscription
func TestSignal_SubscribeForever(t *testing.T) {
	sig := New(0)

	var calls []int
	var mu sync.Mutex

	unsub := sig.SubscribeForever(func(v int) {
		mu.Lock()
		calls = append(calls, v)
		mu.Unlock()
	})
	defer unsub()

	sig.Set(1)
	sig.Set(2)
	sig.Set(3)

	// Give time for callbacks
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(calls) != 3 {
		t.Errorf("Expected 3 callbacks, got %d", len(calls))
	}

	expected := []int{1, 2, 3}
	for i, v := range expected {
		if i >= len(calls) || calls[i] != v {
			t.Errorf("Call %d: got %v, want %d", i, calls, v)
			break
		}
	}
}

// TestSignal_Unsubscribe verifies unsubscribe stops notifications
func TestSignal_Unsubscribe(t *testing.T) {
	sig := New(0)

	var called int32

	unsub := sig.SubscribeForever(func(v int) {
		atomic.AddInt32(&called, 1)
	})

	sig.Set(1)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After Set(1), called = %d, want 1", got)
	}

	// Unsubscribe
	unsub()

	sig.Set(2)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After unsubscribe and Set(2), called = %d, want 1 (no new calls)", got)
	}
}

// TestSignal_MultipleSubscribers verifies multiple subscribers work correctly
func TestSignal_MultipleSubscribers(t *testing.T) {
	sig := New(0)

	var calls1, calls2 int32

	unsub1 := sig.SubscribeForever(func(v int) {
		atomic.AddInt32(&calls1, 1)
	})
	defer unsub1()

	unsub2 := sig.SubscribeForever(func(v int) {
		atomic.AddInt32(&calls2, 1)
	})
	defer unsub2()

	sig.Set(1)
	sig.Set(2)

	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&calls1); got != 2 {
		t.Errorf("Subscriber 1: got %d calls, want 2", got)
	}

	if got := atomic.LoadInt32(&calls2); got != 2 {
		t.Errorf("Subscriber 2: got %d calls, want 2", got)
	}
}

// TestSignal_ContextCancel verifies context-based auto-unsubscribe
func TestSignal_ContextCancel(t *testing.T) {
	sig := New(0)

	ctx, cancel := context.WithCancel(context.Background())

	var called int32

	sig.Subscribe(ctx, func(v int) {
		atomic.AddInt32(&called, 1)
	})

	sig.Set(1)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("Before cancel, called = %d, want 1", got)
	}

	// Cancel context
	cancel()
	time.Sleep(10 * time.Millisecond) // Allow cleanup

	sig.Set(2)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After context cancel, called = %d, want 1 (no new calls)", got)
	}
}

// TestSignal_ContextTimeout verifies timeout-based auto-unsubscribe
func TestSignal_ContextTimeout(t *testing.T) {
	sig := New(0)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var called int32

	sig.Subscribe(ctx, func(v int) {
		atomic.AddInt32(&called, 1)
	})

	sig.Set(1)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("Before timeout, called = %d, want 1", got)
	}

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	sig.Set(2)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After timeout, called = %d, want 1 (no new calls)", got)
	}
}

// TestSignal_EqualFunc verifies custom equality checks
func TestSignal_EqualFunc(t *testing.T) {
	// Create signal with custom equality (compare by length)
	sig := NewWithOptions([]int{1, 2, 3}, Options[[]int]{
		Equal: func(a, b []int) bool {
			return len(a) == len(b)
		},
	})

	var called int32

	sig.SubscribeForever(func(v []int) {
		atomic.AddInt32(&called, 1)
	})

	// Same length - should NOT notify
	sig.Set([]int{4, 5, 6})
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 0 {
		t.Errorf("After Set with same length, called = %d, want 0", got)
	}

	// Different length - should notify
	sig.Set([]int{1, 2})
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("After Set with different length, called = %d, want 1", got)
	}
}

// TestSignal_PanicRecovery verifies panic recovery in subscribers
func TestSignal_PanicRecovery(t *testing.T) {
	sig := New(0)

	var panicCalls, goodCalls int32

	// Panicking subscriber
	sig.SubscribeForever(func(v int) {
		atomic.AddInt32(&panicCalls, 1)
		panic("test panic")
	})

	// Good subscriber (should still be called despite panic)
	sig.SubscribeForever(func(v int) {
		atomic.AddInt32(&goodCalls, 1)
	})

	sig.Set(1)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&panicCalls); got != 1 {
		t.Errorf("Panicking subscriber: got %d calls, want 1", got)
	}

	if got := atomic.LoadInt32(&goodCalls); got != 1 {
		t.Errorf("Good subscriber: got %d calls, want 1 (should be called despite panic)", got)
	}
}

// TestSignal_CustomPanicHandler verifies custom panic handling
func TestSignal_CustomPanicHandler(t *testing.T) {
	var panicHandlerCalled int32

	sig := NewWithOptions(0, Options[int]{
		OnPanic: func(err any, stack []byte) {
			atomic.AddInt32(&panicHandlerCalled, 1)
			if err != "custom panic" {
				t.Errorf("OnPanic: got error %v, want 'custom panic'", err)
			}
		},
	})

	sig.SubscribeForever(func(v int) {
		panic("custom panic")
	})

	sig.Set(1)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&panicHandlerCalled); got != 1 {
		t.Errorf("Custom panic handler called %d times, want 1", got)
	}
}

// TestSignal_AsReadonly verifies read-only wrapper
func TestSignal_AsReadonly(t *testing.T) {
	sig := New(42)
	readonly := sig.AsReadonly()

	// Read should work
	if got := readonly.Get(); got != 42 {
		t.Errorf("readonly.Get() = %d, want 42", got)
	}

	// Subscribe should work
	var called int32
	unsub := readonly.SubscribeForever(func(v int) {
		atomic.AddInt32(&called, 1)
	})
	defer unsub()

	sig.Set(100)
	time.Sleep(10 * time.Millisecond)

	if got := atomic.LoadInt32(&called); got != 1 {
		t.Errorf("readonly subscriber called %d times, want 1", got)
	}

	if got := readonly.Get(); got != 100 {
		t.Errorf("After Set(100), readonly.Get() = %d, want 100", got)
	}
}

// TestSignal_NoMemoryLeak verifies unsubscribe prevents memory leaks
func TestSignal_NoMemoryLeak(t *testing.T) {
	sig := New(0).(*signal[int])

	// Subscribe and unsubscribe 1000 times
	for i := 0; i < 1000; i++ {
		unsub := sig.SubscribeForever(func(v int) {})
		unsub()
	}

	// Check subscribers map is empty
	sig.mu.RLock()
	count := len(sig.subscribers)
	sig.mu.RUnlock()

	if count != 0 {
		t.Errorf("Memory leak: %d subscribers still registered, want 0", count)
	}
}

// TestSignal_ConcurrentReads verifies safe concurrent reads
func TestSignal_ConcurrentReads(t *testing.T) {
	sig := New(42)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = sig.Get()
			}
		}()
	}

	wg.Wait()
}

// TestSignal_ConcurrentWrites verifies safe concurrent writes
func TestSignal_ConcurrentWrites(t *testing.T) {
	sig := New(0)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sig.Update(func(v int) int { return v + 1 })
		}()
	}

	wg.Wait()

	if got := sig.Get(); got != 100 {
		t.Errorf("After 100 concurrent increments, Get() = %d, want 100", got)
	}
}

// TestSignal_ConcurrentSubscribe verifies safe concurrent subscribe/unsubscribe
func TestSignal_ConcurrentSubscribe(t *testing.T) {
	sig := New(0)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsub := sig.SubscribeForever(func(v int) {})
			unsub()
		}()
	}

	wg.Wait()

	// Verify no memory leak
	s := sig.(*signal[int])
	s.mu.RLock()
	count := len(s.subscribers)
	s.mu.RUnlock()

	if count != 0 {
		t.Errorf("After concurrent subscribe/unsubscribe, %d subscribers remain, want 0", count)
	}
}
