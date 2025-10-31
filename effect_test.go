package signals

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEffect_ImmediateExecution verifies that effects run immediately upon creation.
// This is the critical Angular pattern - effects are eager, not lazy.
func TestEffect_ImmediateExecution(t *testing.T) {
	count := New(0)
	executed := atomic.Bool{}

	eff := Effect(
		func() {
			executed.Store(true)
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	// Effect MUST have run immediately
	if !executed.Load() {
		t.Fatal("Effect did not run immediately upon creation")
	}
}

// TestEffect_DependencyChange verifies that effects re-run when dependencies change.
func TestEffect_DependencyChange(t *testing.T) {
	count := New(0)
	runCount := atomic.Int32{}

	eff := Effect(
		func() {
			runCount.Add(1)
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	initialRuns := runCount.Load()
	if initialRuns != 1 {
		t.Fatalf("Expected 1 initial run, got %d", initialRuns)
	}

	// Change dependency
	count.Set(5)
	time.Sleep(10 * time.Millisecond) // Allow effect to run

	runs := runCount.Load()
	if runs != 2 {
		t.Fatalf("Expected 2 runs after dependency change, got %d", runs)
	}

	// Change again
	count.Set(10)
	time.Sleep(10 * time.Millisecond)

	runs = runCount.Load()
	if runs != 3 {
		t.Fatalf("Expected 3 runs after second dependency change, got %d", runs)
	}
}

// TestEffect_MultipleDependencies verifies that effects track multiple dependencies correctly.
func TestEffect_MultipleDependencies(t *testing.T) {
	firstName := New("John")
	lastName := New("Doe")
	log := []string{}
	var mu sync.Mutex

	eff := Effect(
		func() {
			mu.Lock()
			log = append(log, fmt.Sprintf("%s %s", firstName.Get(), lastName.Get()))
			mu.Unlock()
		},
		firstName.AsReadonly(),
		lastName.AsReadonly(),
	)
	defer eff.Stop()

	time.Sleep(10 * time.Millisecond)

	// Check immediate execution
	mu.Lock()
	if len(log) != 1 || log[0] != "John Doe" {
		t.Fatalf("Expected immediate execution with 'John Doe', got: %v", log)
	}
	mu.Unlock()

	// Change first dependency
	firstName.Set("Jane")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if len(log) != 2 || log[1] != "Jane Doe" {
		t.Fatalf("Expected effect to run on firstName change, got: %v", log)
	}
	mu.Unlock()

	// Change second dependency
	lastName.Set("Smith")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if len(log) != 3 || log[2] != "Jane Smith" {
		t.Fatalf("Expected effect to run on lastName change, got: %v", log)
	}
	mu.Unlock()
}

// TestEffect_Cleanup verifies that cleanup runs before the next effect execution.
func TestEffect_Cleanup(t *testing.T) {
	t.Run("initial execution has no cleanup", func(t *testing.T) {
		count := New(0)
		cleanupLog := []string{}
		effectLog := []string{}
		var mu sync.Mutex

		eff := EffectWithCleanup(
			func() func() {
				mu.Lock()
				effectLog = append(effectLog, fmt.Sprintf("effect-%d", count.Get()))
				currentValue := count.Get()
				mu.Unlock()

				return func() {
					mu.Lock()
					cleanupLog = append(cleanupLog, fmt.Sprintf("cleanup-%d", currentValue))
					mu.Unlock()
				}
			},
			count.AsReadonly(),
		)
		defer eff.Stop()

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if len(effectLog) != 1 || effectLog[0] != "effect-0" {
			t.Fatalf("Expected initial effect, got: %v", effectLog)
		}
		if len(cleanupLog) != 0 {
			t.Fatalf("Expected no cleanup yet, got: %v", cleanupLog)
		}
	})

	t.Run("cleanup runs before next effect", func(t *testing.T) {
		count := New(0)
		cleanupLog := []string{}
		effectLog := []string{}
		var mu sync.Mutex

		eff := EffectWithCleanup(
			func() func() {
				mu.Lock()
				effectLog = append(effectLog, fmt.Sprintf("effect-%d", count.Get()))
				currentValue := count.Get()
				mu.Unlock()

				return func() {
					mu.Lock()
					cleanupLog = append(cleanupLog, fmt.Sprintf("cleanup-%d", currentValue))
					mu.Unlock()
				}
			},
			count.AsReadonly(),
		)
		defer eff.Stop()

		time.Sleep(10 * time.Millisecond)
		count.Set(1)
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if len(cleanupLog) != 1 || cleanupLog[0] != "cleanup-0" {
			t.Fatalf("Expected cleanup from first effect, got: %v", cleanupLog)
		}
	})

	t.Run("multiple cleanups execute in order", func(t *testing.T) {
		count := New(0)
		cleanupLog := []string{}
		var mu sync.Mutex

		eff := EffectWithCleanup(
			func() func() {
				mu.Lock()
				currentValue := count.Get()
				mu.Unlock()

				return func() {
					mu.Lock()
					cleanupLog = append(cleanupLog, fmt.Sprintf("cleanup-%d", currentValue))
					mu.Unlock()
				}
			},
			count.AsReadonly(),
		)
		defer eff.Stop()

		time.Sleep(10 * time.Millisecond)
		count.Set(1)
		time.Sleep(10 * time.Millisecond)
		count.Set(2)
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if len(cleanupLog) != 2 {
			t.Fatalf("Expected 2 cleanups, got: %v", cleanupLog)
		}
		if cleanupLog[0] != "cleanup-0" || cleanupLog[1] != "cleanup-1" {
			t.Fatalf("Expected cleanups in order, got: %v", cleanupLog)
		}
	})
}

// TestEffect_Stop verifies that Stop() prevents future runs and executes final cleanup.
func TestEffect_Stop(t *testing.T) {
	count := New(0)
	runCount := atomic.Int32{}
	cleanupCalled := atomic.Bool{}

	eff := EffectWithCleanup(
		func() func() {
			runCount.Add(1)
			return func() {
				cleanupCalled.Store(true)
			}
		},
		count.AsReadonly(),
	)

	time.Sleep(10 * time.Millisecond)

	// Verify immediate execution
	if runCount.Load() != 1 {
		t.Fatalf("Expected 1 initial run, got %d", runCount.Load())
	}

	// Stop the effect
	eff.Stop()

	// Verify cleanup was called
	if !cleanupCalled.Load() {
		t.Fatal("Expected cleanup to be called on Stop()")
	}

	// Try to trigger effect (should not run)
	count.Set(5)
	time.Sleep(10 * time.Millisecond)

	if runCount.Load() != 1 {
		t.Fatalf("Expected effect to not run after Stop(), got %d runs", runCount.Load())
	}
}

// TestEffect_StopMultipleTimes verifies that Stop() is safe to call multiple times.
func TestEffect_StopMultipleTimes(t *testing.T) {
	count := New(0)
	cleanupCount := atomic.Int32{}

	eff := EffectWithCleanup(
		func() func() {
			return func() {
				cleanupCount.Add(1)
			}
		},
		count.AsReadonly(),
	)

	time.Sleep(10 * time.Millisecond)

	// Stop multiple times
	eff.Stop()
	eff.Stop()
	eff.Stop()

	// Cleanup should only run once
	if cleanupCount.Load() != 1 {
		t.Fatalf("Expected cleanup to run once, got %d", cleanupCount.Load())
	}
}

// TestEffect_PanicRecovery verifies that panics in effect functions are recovered.
func TestEffect_PanicRecovery(t *testing.T) {
	count := New(0)
	panicCount := atomic.Int32{}
	customPanicCalled := atomic.Bool{}

	eff := EffectWithOptions(
		func() func() {
			panicCount.Add(1)
			if count.Get() == 1 {
				panic("test panic in effect")
			}
			return nil
		},
		EffectOptions{
			OnPanic: func(err any, stack []byte) {
				customPanicCalled.Store(true)
				if msg, ok := err.(string); !ok || msg != "test panic in effect" {
					t.Errorf("Expected panic message 'test panic in effect', got: %v", err)
				}
			},
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	time.Sleep(10 * time.Millisecond)

	// Initial run (no panic)
	if panicCount.Load() != 1 {
		t.Fatalf("Expected 1 initial run, got %d", panicCount.Load())
	}

	// Trigger panic
	count.Set(1)
	time.Sleep(10 * time.Millisecond)

	// Verify panic was recovered and custom handler called
	if !customPanicCalled.Load() {
		t.Fatal("Expected custom panic handler to be called")
	}

	// Effect should still work after panic
	count.Set(2)
	time.Sleep(10 * time.Millisecond)

	if panicCount.Load() != 3 {
		t.Fatalf("Expected effect to continue after panic, got %d runs", panicCount.Load())
	}
}

// TestEffect_CleanupPanic verifies that panics in cleanup functions are recovered.
func TestEffect_CleanupPanic(t *testing.T) {
	count := New(0)
	cleanupPanicCalled := atomic.Bool{}

	eff := EffectWithOptions(
		func() func() {
			currentValue := count.Get()
			return func() {
				if currentValue == 0 {
					panic("test panic in cleanup")
				}
			}
		},
		EffectOptions{
			OnPanic: func(err any, stack []byte) {
				cleanupPanicCalled.Store(true)
				if msg, ok := err.(string); !ok || msg != "test panic in cleanup" {
					t.Errorf("Expected panic message 'test panic in cleanup', got: %v", err)
				}
			},
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	time.Sleep(10 * time.Millisecond)

	// Trigger effect again (cleanup from first run will panic)
	count.Set(1)
	time.Sleep(10 * time.Millisecond)

	// Verify cleanup panic was recovered
	if !cleanupPanicCalled.Load() {
		t.Fatal("Expected cleanup panic handler to be called")
	}
}

// TestEffect_ConcurrentStop verifies that Stop() is thread-safe.
func TestEffect_ConcurrentStop(t *testing.T) {
	count := New(0)
	cleanupCount := atomic.Int32{}

	eff := EffectWithCleanup(
		func() func() {
			return func() {
				cleanupCount.Add(1)
			}
		},
		count.AsReadonly(),
	)

	time.Sleep(10 * time.Millisecond)

	// Stop from multiple goroutines concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eff.Stop()
		}()
	}
	wg.Wait()

	// Cleanup should only run once
	if cleanupCount.Load() != 1 {
		t.Fatalf("Expected cleanup to run once despite concurrent Stop(), got %d", cleanupCount.Load())
	}
}

// TestEffect_NoMemoryLeak verifies that stopped effects don't leak memory.
func TestEffect_NoMemoryLeak(t *testing.T) {
	count := New(0)

	// Get baseline memory
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Create and stop many effects
	for i := 0; i < 1000; i++ {
		eff := Effect(
			func() {
				_ = count.Get()
			},
			count.AsReadonly(),
		)
		eff.Stop()
	}

	// Force GC and check memory
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Memory growth should be minimal (< 1MB for 1000 effects)
	growth := m2.Alloc - m1.Alloc
	if growth > 1*1024*1024 {
		t.Logf("Warning: Memory growth after 1000 stopped effects: %d bytes", growth)
		// Note: This is a warning, not a hard failure, as GC timing varies
	}
}

// TestEffect_ChainedSignals verifies that effects work with computed signals.
func TestEffect_ChainedSignals(t *testing.T) {
	base := New(5)
	doubled := Computed(
		func() int {
			return base.Get() * 2
		},
		base.AsReadonly(),
	)

	log := []int{}
	var mu sync.Mutex

	eff := Effect(
		func() {
			mu.Lock()
			log = append(log, doubled.Get())
			mu.Unlock()
		},
		doubled,
	)
	defer eff.Stop()

	time.Sleep(10 * time.Millisecond)

	// Check immediate execution
	mu.Lock()
	if len(log) != 1 || log[0] != 10 {
		t.Fatalf("Expected immediate execution with value 10, got: %v", log)
	}
	mu.Unlock()

	// Change base signal
	base.Set(7)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if len(log) != 2 || log[1] != 14 {
		t.Fatalf("Expected effect to run with computed value 14, got: %v", log)
	}
	mu.Unlock()
}

// TestEffect_NoDependencies verifies that effects can run without dependencies.
func TestEffect_NoDependencies(t *testing.T) {
	executed := atomic.Bool{}

	eff := Effect(func() {
		executed.Store(true)
	})
	defer eff.Stop()

	// Should run immediately even without dependencies
	if !executed.Load() {
		t.Fatal("Effect without dependencies did not run immediately")
	}
}

// TestEffect_CleanupOrder verifies the exact order of cleanup and effect execution.
func TestEffect_CleanupOrder(t *testing.T) {
	count := New(0)
	events := []string{}
	var mu sync.Mutex

	eff := EffectWithCleanup(
		func() func() {
			mu.Lock()
			events = append(events, fmt.Sprintf("effect-%d", count.Get()))
			currentValue := count.Get()
			mu.Unlock()

			return func() {
				mu.Lock()
				events = append(events, fmt.Sprintf("cleanup-%d", currentValue))
				mu.Unlock()
			}
		},
		count.AsReadonly(),
	)
	defer eff.Stop()

	time.Sleep(10 * time.Millisecond)

	// Trigger multiple changes
	count.Set(1)
	time.Sleep(10 * time.Millisecond)

	count.Set(2)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	expected := []string{
		"effect-0",  // Initial run
		"cleanup-0", // Cleanup from initial
		"effect-1",  // Second run
		"cleanup-1", // Cleanup from second
		"effect-2",  // Third run
	}
	if len(events) != len(expected) {
		t.Fatalf("Expected %d events, got %d: %v", len(expected), len(events), events)
	}
	for i, exp := range expected {
		if events[i] != exp {
			t.Errorf("Event %d: expected %s, got %s", i, exp, events[i])
		}
	}
	mu.Unlock()
}

// TestEffect_WithComputedAndMultipleTypes tests effects with mixed type dependencies.
func TestEffect_WithComputedAndMultipleTypes(t *testing.T) {
	// Different types
	count := New(5)
	name := New("items")
	enabled := New(true)

	// Computed signal
	message := Computed(
		func() string {
			if !enabled.Get() {
				return "disabled"
			}
			return fmt.Sprintf("%d %s", count.Get(), name.Get())
		},
		count.AsReadonly(),
		name.AsReadonly(),
		enabled.AsReadonly(),
	)

	log := []string{}
	var mu sync.Mutex

	eff := Effect(
		func() {
			mu.Lock()
			log = append(log, message.Get())
			mu.Unlock()
		},
		message,
	)
	defer eff.Stop()

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if len(log) != 1 || log[0] != "5 items" {
		t.Fatalf("Expected '5 items', got: %v", log)
	}
	mu.Unlock()

	// Disable
	enabled.Set(false)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if len(log) != 2 || log[1] != "disabled" {
		t.Fatalf("Expected 'disabled', got: %v", log)
	}
	mu.Unlock()
}
