package main

import (
	"fmt"

	"github.com/coregx/signals"
)

func main() {
	demoBasicSignals()
	demoComputedSignals()
	demoEffects()
	fmt.Println("\n=== Demo Complete ===")
}

func demoBasicSignals() {
	fmt.Println("=== Phase 1: Basic Signals ===")

	// Create a signal
	s := signals.New("test")

	// Subscribe to changes
	unsub := s.SubscribeForever(func(v string) {
		fmt.Println("Signal changed:", v)
	})
	defer unsub()

	// Get current value
	value := s.Get()
	fmt.Println("Current value:", value)

	// Update value
	s.Set("test1")

	// Use Update to transform
	s.Update(func(v string) string {
		return v + "_updated"
	})
}

func demoComputedSignals() {
	fmt.Println("\n=== Phase 2: Computed Signals ===")

	// Example 1: Basic computed signal
	count := signals.New(5)
	doubled := signals.Computed(
		func() int {
			return count.Get() * 2
		},
		count.AsReadonly(),
	)

	fmt.Printf("count = %d, doubled = %d\n", count.Get(), doubled.Get())

	count.Set(10)
	fmt.Printf("After count.Set(10): doubled = %d\n", doubled.Get())

	// Example 2: Computed signal with multiple dependencies
	firstName := signals.New("John")
	lastName := signals.New("Doe")

	fullName := signals.Computed(
		func() string {
			return firstName.Get() + " " + lastName.Get()
		},
		firstName.AsReadonly(),
		lastName.AsReadonly(),
	)

	fmt.Printf("\nFull name: %s\n", fullName.Get())

	firstName.Set("Jane")
	fmt.Printf("After firstName.Set('Jane'): %s\n", fullName.Get())

	// Example 3: Chained computed signals
	quadrupled := signals.Computed(
		func() int {
			return doubled.Get() * 2
		},
		doubled,
	)

	fmt.Printf("\ncount = %d, quadrupled = %d\n", count.Get(), quadrupled.Get())

	count.Set(5)
	fmt.Printf("After count.Set(5): quadrupled = %d\n", quadrupled.Get())

	// Example 4: Subscribe to computed signals
	fmt.Println("\nSubscribing to computed signal...")
	unsubComputed := fullName.SubscribeForever(func(v string) {
		fmt.Println("Full name changed:", v)
	})
	defer unsubComputed()

	lastName.Set("Smith")
}

func demoEffects() {
	fmt.Println("\n=== Phase 3: Effects ===")
	demoBasicEffect()
	demoEffectMultipleDeps()
	demoEffectWithCleanup()
	demoEffectWithComputed()
	demoEffectSideEffects()
}

func demoBasicEffect() {
	// Example 1: Basic effect (runs immediately!)
	effectCount := signals.New(0)
	fmt.Println("\nCreating effect (will run immediately)...")

	eff1 := signals.Effect(
		func() {
			fmt.Printf("Effect running! Count is: %d\n", effectCount.Get())
		},
		effectCount.AsReadonly(),
	)
	defer eff1.Stop()

	// Change the signal (effect runs again)
	effectCount.Set(5)
	effectCount.Set(10)
}

func demoEffectMultipleDeps() {
	// Example 2: Effect with multiple dependencies
	fmt.Println("\nEffect with multiple dependencies:")
	x := signals.New(3)
	y := signals.New(4)

	eff2 := signals.Effect(
		func() {
			sum := x.Get() + y.Get()
			fmt.Printf("x=%d, y=%d, sum=%d\n", x.Get(), y.Get(), sum)
		},
		x.AsReadonly(),
		y.AsReadonly(),
	)
	defer eff2.Stop()

	x.Set(5) // Effect runs: x=5, y=4, sum=9
	y.Set(6) // Effect runs: x=5, y=6, sum=11
}

func demoEffectWithCleanup() {
	// Example 3: Effect with cleanup
	fmt.Println("\nEffect with cleanup:")
	timer := signals.New(0)

	eff3 := signals.EffectWithCleanup(
		func() func() {
			currentValue := timer.Get()
			fmt.Printf("Starting timer with value: %d\n", currentValue)

			return func() {
				fmt.Printf("Cleaning up timer value: %d\n", currentValue)
			}
		},
		timer.AsReadonly(),
	)

	timer.Set(1)
	timer.Set(2)
	eff3.Stop()
}

func demoEffectWithComputed() {
	// Example 4: Effect with computed signal
	fmt.Println("\nEffect with computed signal:")
	base := signals.New(10)
	tripled := signals.Computed(
		func() int {
			return base.Get() * 3
		},
		base.AsReadonly(),
	)

	eff4 := signals.Effect(
		func() {
			fmt.Printf("Base=%d, Tripled=%d\n", base.Get(), tripled.Get())
		},
		tripled,
	)
	defer eff4.Stop()

	base.Set(20)
}

func demoEffectSideEffects() {
	// Example 5: Effect for side effects (logging, metrics, etc.)
	fmt.Println("\nEffect for side effects (simulating analytics):")
	userName := signals.New("Alice")
	userAge := signals.New(25)

	eff5 := signals.Effect(
		func() {
			fmt.Printf("[Analytics] User profile viewed: %s, age %d\n", userName.Get(), userAge.Get())
		},
		userName.AsReadonly(),
		userAge.AsReadonly(),
	)
	defer eff5.Stop()

	userName.Set("Bob")
	userAge.Set(30)
}
