package signals

import (
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// EffectRef represents a running side effect that can be stopped.
//
// Effects run immediately upon creation (Angular pattern) and re-run when dependencies change.
// They support cleanup functions that run before the next effect and when stopped.
//
// Use the Stop() method to clean up the effect when no longer needed.
type EffectRef interface {
	// Stop stops the effect and runs final cleanup.
	// After calling Stop, the effect will no longer run.
	// Safe to call multiple times.
	Stop()
}

// effect is the internal implementation of Effect.
type effect struct {
	// fn is the effect function that returns a cleanup function (EffectWithCleanup path)
	fn func() func()

	// fnSimple is the effect function without cleanup (Effect path).
	// When set, fn is nil — avoids a wrapper closure allocation.
	fnSimple func()

	// cleanup is the current cleanup function from the last run
	cleanup func()

	// unsubscribes are cleanup functions for dependency subscriptions
	unsubscribes []Unsubscribe

	// mu protects cleanup field
	mu sync.Mutex

	// stopped prevents effect from running after Stop()
	stopped atomic.Bool

	// onPanic is optional custom panic handler
	onPanic func(any, []byte)
}

// Effect creates an effect that runs immediately and on dependency changes.
//
// CRITICAL: The effect function runs IMMEDIATELY upon creation, then again
// whenever any dependency changes. This matches Angular's effect() behavior.
//
// Dependencies must be explicitly passed as additional arguments. Each dependency
// can be a ReadonlySignal of any type.
//
// Example:
//
//	count := signals.New(0)
//	name := signals.New("Alice")
//
//	// Effect runs immediately (prints "Alice: 0")
//	// Then runs again when count or name changes
//	eff := signals.Effect(
//	    func() {
//	        fmt.Printf("%s: %d\n", name.Get(), count.Get())
//	    },
//	    count.AsReadonly(),
//	    name.AsReadonly(),
//	)
//	defer eff.Stop()
//
//	count.Set(5)  // Effect runs again (prints "Alice: 5")
//	name.Set("Bob")  // Effect runs again (prints "Bob: 5")
//
// For effects that need cleanup, use EffectWithCleanup instead.
func Effect(fn func(), deps ...any) EffectRef {
	return effectCreate(fn, nil, EffectOptions{}, deps...)
}

// EffectWithCleanup creates an effect with cleanup callback support.
//
// The effect function returns a cleanup function that will be called:
//   - Before the next effect execution
//   - When Stop() is called
//
// This is useful for:
//   - Canceling timers or intervals
//   - Closing connections or file handles
//   - Removing event listeners
//   - Aborting pending operations
//
// Example:
//
//	count := signals.New(0)
//
//	eff := signals.EffectWithCleanup(
//	    func() func() {
//	        // Effect: start timer
//	        ticker := time.NewTicker(time.Second)
//	        go func() {
//	            for range ticker.C {
//	                fmt.Println("Tick:", count.Get())
//	            }
//	        }()
//
//	        // Cleanup: stop timer
//	        return func() {
//	            ticker.Stop()
//	        }
//	    },
//	    count.AsReadonly(),
//	)
//	defer eff.Stop()  // Runs cleanup
//
//	count.Set(5)  // Old cleanup runs, new effect starts, new cleanup registered
func EffectWithCleanup(fn func() func(), deps ...any) EffectRef {
	return EffectWithOptions(fn, EffectOptions{}, deps...)
}

// EffectOptions configures effect behavior.
type EffectOptions struct {
	// OnPanic is called when the effect or cleanup function panics.
	// If nil, panics are logged to stderr.
	OnPanic func(err any, stack []byte)
}

// EffectWithOptions creates an effect with custom options.
//
// Use this when you need custom panic handling for effects or cleanup functions.
//
// Example:
//
//	count := signals.New(0)
//	eff := signals.EffectWithOptions(
//	    func() func() {
//	        fmt.Println("Effect:", count.Get())
//	        return nil
//	    },
//	    signals.EffectOptions{
//	        OnPanic: func(err any, stack []byte) {
//	            metrics.IncrementEffectPanic()
//	        },
//	    },
//	    count.AsReadonly(),
//	)
func EffectWithOptions(fn func() func(), opts EffectOptions, deps ...any) EffectRef {
	return effectCreate(nil, fn, opts, deps...)
}

// effectCreate is the shared constructor for both Effect and EffectWithCleanup paths.
// Exactly one of fnSimple or fn must be non-nil.
func effectCreate(fnSimple func(), fn func() func(), opts EffectOptions, deps ...any) EffectRef {
	e := &effect{
		fn:           fn,
		fnSimple:     fnSimple,
		onPanic:      opts.OnPanic,
		unsubscribes: make([]Unsubscribe, 0, len(deps)),
	}

	runFn := e.run
	for _, dep := range deps {
		unsub := trackDependencyHelper(dep, runFn)
		e.unsubscribes = append(e.unsubscribes, unsub)
	}

	// CRITICAL: Run effect IMMEDIATELY (Angular pattern)
	e.run()

	return e
}

// run executes the effect function with proper cleanup handling.
//
// Cleanup sequence:
//  1. Check if stopped (early return if true)
//  2. Run old cleanup (if exists)
//  3. Execute effect function
//  4. Store new cleanup (if returned)
//
// All steps have panic recovery to prevent one bad effect from breaking others.
func (e *effect) run() {
	// Don't run if stopped
	if e.stopped.Load() {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check after acquiring lock
	if e.stopped.Load() {
		return
	}

	// Step 1: Run old cleanup (if exists)
	if e.cleanup != nil {
		oldCleanup := e.cleanup
		e.cleanup = nil

		func() {
			defer func() {
				if r := recover(); r != nil {
					if e.onPanic != nil {
						e.onPanic(r, debug.Stack())
					} else {
						log.Printf("signals: panic in effect cleanup: %v\n%s", r, debug.Stack())
					}
				}
			}()
			oldCleanup()
		}()
	}

	// Step 2: Execute effect function
	if e.fnSimple != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if e.onPanic != nil {
						e.onPanic(r, debug.Stack())
					} else {
						log.Printf("signals: panic in effect function: %v\n%s", r, debug.Stack())
					}
				}
			}()
			e.fnSimple()
		}()
	} else {
		var newCleanup func()
		func() {
			defer func() {
				if r := recover(); r != nil {
					if e.onPanic != nil {
						e.onPanic(r, debug.Stack())
					} else {
						log.Printf("signals: panic in effect function: %v\n%s", r, debug.Stack())
					}
				}
			}()
			newCleanup = e.fn()
		}()
		e.cleanup = newCleanup
	}
}

// Stop stops the effect and runs final cleanup.
//
// After calling Stop:
//   - The effect will no longer run when dependencies change
//   - All dependency subscriptions are canceled
//   - The final cleanup function is executed (if any)
//
// Safe to call multiple times. Subsequent calls are no-ops.
//
// Example:
//
//	eff := signals.Effect(func() {
//	    fmt.Println("Running")
//	}, dep.AsReadonly())
//
//	// Later...
//	eff.Stop()  // Runs cleanup, unsubscribes from deps
//	eff.Stop()  // Safe, does nothing
func (e *effect) Stop() {
	// Set stopped flag (prevents future runs)
	if e.stopped.Swap(true) {
		// Already stopped
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Run final cleanup
	if e.cleanup != nil {
		cleanup := e.cleanup
		e.cleanup = nil

		func() {
			defer func() {
				if r := recover(); r != nil {
					if e.onPanic != nil {
						e.onPanic(r, debug.Stack())
					} else {
						log.Printf("signals: panic in final effect cleanup: %v\n%s", r, debug.Stack())
					}
				}
			}()
			cleanup()
		}()
	}

	// Unsubscribe from all dependencies
	for _, unsub := range e.unsubscribes {
		unsub()
	}
	e.unsubscribes = nil
}
